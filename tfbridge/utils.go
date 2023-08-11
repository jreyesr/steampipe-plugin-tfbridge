package tfbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/jreyesr/steampipe-plugin-tfbridge/logging"
	tfplugin "github.com/jreyesr/steampipe-plugin-tfbridge/plugin"
	tfplugin6 "github.com/jreyesr/steampipe-plugin-tfbridge/plugin6"
	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	spPlugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var Handshake = plugin.HandshakeConfig{
	// This comes directly from Terraform's repo
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

func getPluginConnection(pluginPath string) (providers.Interface, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: tfplugin.VersionedPlugins,
		Cmd:              exec.Command("sh", "-c", pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger:           logging.NewProviderLogger(""),
		SyncStdout:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stdout", "dns")),
		SyncStderr:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stderr", "dns")),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	x, err := rpcClient.Dispense("provider") // Grab the provider
	if err != nil {
		return nil, err
	}
	// store the client so that the plugin can kill the child process
	protoVer := client.NegotiatedVersion()
	switch protoVer {
	case 5:
		p := x.(*tfplugin.GRPCProvider)
		p.PluginClient = client
		return p, nil
	case 6:
		p := x.(*tfplugin6.GRPCProvider)
		p.PluginClient = client
		return p, nil
	default:
		return nil, fmt.Errorf("can't cast %v (%T) to GRPCProvider", x, x)
	}
}

func getProviderSchema(ctx context.Context, provider providers.Interface) hcldec.Spec {
	schema := provider.GetProviderSchema()
	spec := schema.Provider.Block.DecoderSpec()

	return spec
}

func configureProvider(ctx context.Context, provider providers.Interface, rawConfig string) error {
	spPlugin.Logger(ctx).Debug("configureProvider", "rawConfig", rawConfig)

	// grab config schema from provider
	spec := getProviderSchema(ctx, provider)

	// parse HCL string using provider's schema as blueprint
	// if it fails, user wrote incorrect HCL string on .spc file
	parser := hclparse.NewParser()
	f, err := parser.ParseHCL([]byte(rawConfig), "config.hcl")
	if err != nil {
		spPlugin.Logger(ctx).Error("configureProvider.ParseHCL", "rawConfig", rawConfig, "err", err)
		return err
	}
	cfgVal, err := hcldec.Decode(f.Body, spec, &hcl.EvalContext{})
	if err != nil {
		spPlugin.Logger(ctx).Error("configureProvider.Decode", "body", f.Body, "spec", spec)
		return err
	}

	configType := hcldec.ImpliedType(spec)
	spPlugin.Logger(ctx).Debug("configureProvider", "parsedConfigType", configType, "parsedConfig", cfgVal)

	// ACTUALLY send the configure RPC to provider binary
	configureResponse := provider.ConfigureProvider(providers.ConfigureProviderRequest{
		TerraformVersion: "999.0.0",
		Config:           cfgVal,
	})
	if configureResponse.Diagnostics.HasErrors() {
		spPlugin.Logger(ctx).Error("configureProvider.ConfigureProvider", "config", cfgVal, "err", configureResponse.Diagnostics.Err())
		return configureResponse.Diagnostics.Err()
	}

	return nil
}

func getDataSources(provider providers.Interface) (map[string]providers.Schema, error) {
	schema := provider.GetProviderSchema()
	if schema.Diagnostics.HasErrors() {
		return nil, schema.Diagnostics.Err()
	}

	return schema.DataSources, nil
}

func getDataSourceSchema(provider providers.Interface, dataSourceName string) (*providers.Schema, error) {
	schemas, err := getDataSources(provider)
	if err != nil {
		return nil, err
	}

	schema, ok := schemas[dataSourceName] // grab the schema of this datasource
	if !ok {
		return nil, fmt.Errorf("data source %s not found", dataSourceName)
	}

	return &schema, nil
}

func readDataSource(ctx context.Context, provider providers.Interface, dataSourceName string, quals map[string]*proto.QualValue) (*cty.Value, error) {
	dsSchema, err := getDataSourceSchema(provider, dataSourceName)
	if err != nil {
		spPlugin.Logger(ctx).Warn("readDataSource.getDataSourceSchema", "provider", provider, "dataSource", dataSourceName)
		return nil, err
	}
	dsSchemaType := dsSchema.Block.ImpliedType()
	spPlugin.Logger(ctx).Debug("readDataSource", "dsSchema", dsSchema, "dsSchemaType", dsSchemaType)

	simpleQuals := make(map[string]any)
	for k, v := range quals {
		if attr, ok := dsSchema.Block.Attributes[k]; ok {
			switch attr.Type {
			case cty.Number:
				simpleQuals[k] = v.GetDoubleValue()
			case cty.String:
				simpleQuals[k] = v.GetStringValue()
			case cty.Bool:
				simpleQuals[k] = v.GetBoolValue()
			default:
				errmsg := fmt.Errorf("type %v can't be handled by quals", attr.Type)
				spPlugin.Logger(ctx).Warn("readDataSource.makeSimpleQuals.unsupported", "qualName", k, "qual", v, "typeInSchema", attr.Type, "err", errmsg)
				return nil, errmsg
			}
		}
		if _, ok := dsSchema.Block.BlockTypes[k]; ok {
			// if qual matches a nested block, it must have come packed in a JSONB field
			var d map[string]any
			err := json.Unmarshal([]byte(v.GetJsonbValue()), &d)
			if err != nil {
				return nil, err
			}
			simpleQuals[k] = d
		}
	}
	// do the dance: marshal quals into JSON string...
	qualsString, err := json.Marshal(simpleQuals)
	if err != nil {
		spPlugin.Logger(ctx).Warn("readDataSource.jsonMarshal", "err", err, "quals", simpleQuals)
		return nil, err
	}

	// ... and then deserialize into cty.Value, using the expected type as a guide
	// https://pkg.go.dev/github.com/zclconf/go-cty@v1.13.2/cty/json#Unmarshal
	dsSchemaVal, err := ctyjson.Unmarshal(qualsString, dsSchemaType)
	if err != nil {
		spPlugin.Logger(ctx).Warn("readDataSource.ctyUnmarshal", "err", err, "quals", string(qualsString), "schema", dsSchemaType)
		return nil, err
	}

	spPlugin.Logger(ctx).Debug("readDataSource", "quals", simpleQuals, "readConfig", dsSchemaVal)

	// now provide the cty.Value to the RPC interface
	readResponse := provider.ReadDataSource(providers.ReadDataSourceRequest{
		TypeName:     dataSourceName,
		Config:       dsSchemaVal,
		ProviderMeta: cty.EmptyObjectVal,
	})
	if readResponse.Diagnostics.HasErrors() {
		return nil, readResponse.Diagnostics.Err()
	}
	spPlugin.Logger(ctx).Debug("readDataSource.response", "response", readResponse.State)

	return &readResponse.State, nil
}

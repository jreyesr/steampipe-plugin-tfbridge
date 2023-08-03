package tfbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/jreyesr/steampipe-plugin-tfbridge/logging"
	tfplugin "github.com/jreyesr/steampipe-plugin-tfbridge/plugin"
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

func getPluginConnection(pluginPath string) (*tfplugin.GRPCProvider, error) {
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
	provider, valid := x.(*tfplugin.GRPCProvider)
	if !valid {
		return nil, fmt.Errorf("can't cast %v (%T) to GRPCProvider", provider, provider)
	}

	return provider, nil
}

func getDataSources(provider *tfplugin.GRPCProvider) (map[string]providers.Schema, error) {
	schema := provider.GetProviderSchema()
	if schema.Diagnostics.HasErrors() {
		return nil, schema.Diagnostics.Err()
	}

	return schema.DataSources, nil
}

func getDataSourceSchema(provider *tfplugin.GRPCProvider, dataSourceName string) (*providers.Schema, error) {
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

func readDataSource(ctx context.Context, provider *tfplugin.GRPCProvider, dataSourceName string, quals map[string]*proto.QualValue) (*cty.Value, error) {
	dsSchema, err := getDataSourceSchema(provider, dataSourceName)
	if err != nil {
		spPlugin.Logger(ctx).Warn("readDataSource.getDataSourceSchema", "provider", provider, "dataSource", dataSourceName)
		return nil, err
	}
	dsSchemaType := dsSchema.Block.ImpliedType()

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

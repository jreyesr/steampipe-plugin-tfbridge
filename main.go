package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/jreyesr/steampipe-plugin-tfbridge/logging"
	tfplugin "github.com/jreyesr/steampipe-plugin-tfbridge/plugin"
	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
)

var Handshake = plugin.HandshakeConfig{
	// The ProtocolVersion is the version that must match between TF core
	// and TF plugins. This should be bumped whenever a change happens in
	// one or the other that makes it so that they can't safely communicate.
	// This could be adding a new interface value, it could be how
	// helper/schema computes diffs, etc.
	// ProtocolVersion: 4,

	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

func main() {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: tfplugin.VersionedPlugins,
		// Cmd:              exec.Command("sh", "-c", "./terraform-provider-dns_v3.2.4_x5"),
		Cmd:              exec.Command("sh", "-c", "./terraform-provider-tfe_v0.47.0_x5"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger:           logging.NewProviderLogger(""),
		SyncStdout:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stdout", "dns")),
		SyncStderr:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stderr", "dns")),
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	x, _ := rpcClient.Dispense("provider") // Grab the provider
	provider := x.(*tfplugin.GRPCProvider)

	schema := provider.GetProviderSchema()
	spec := schema.Provider.Block.DecoderSpec()
	// fmt.Printf("PROVIDER: %v %d\n", schema.Provider.Block.BlockTypes["update"].Attributes["port"], len(schema.Provider.Block.BlockTypes))

	// type UpdateBlock struct {
	// 	GSSAPI []struct {
	// 		Keytab   string `cty:"keytab"`
	// 		Password string `cty:"password"`
	// 		Realm    string `cty:"realm"`
	// 		Username string `cty:"username"`
	// 	} `cty:"gssapi"`
	// 	KeyAlgorithm string `cty:"key_algorithm"`
	// 	KeyName      string `cty:"key_name"`
	// 	KeySecret    string `cty:"key_secret"`
	// 	Port         int    `cty:"port"`
	// 	Retries      int    `cty:"retries"`
	// 	Server       string `cty:"server"`
	// 	Timeout      string `cty:"timeout"`
	// 	Transport    string `cty:"transport"`
	// }
	// type ProviderConfig struct {
	// 	Update []UpdateBlock `cty:"update"`
	// }
	// type TFEConfig struct {
	// 	Hostname      string `cty:"hostname"`
	// 	Token         string `cty:"token"`
	// 	SSLSkipVerify bool   `cty:"ssl_skip_verify"`
	// 	Organization  string `cty:"organization"`
	// }
	// cfg := ProviderConfig{Update: []UpdateBlock{
	// 	{Port: 1553, Server: "123"},
	// }}
	// cfg := TFEConfig{
	// 	Token:    "invalid",
	// 	Hostname: "192.168.100.1",
	// }
	// configType, _ := gocty.ImpliedType(cfg)
	// cfgVal, _ := gocty.ToCtyValue(cfg, configType)

	config := `
		organization      = "my-org"
		# hostname   = "localhost"
		token = "secret-token"
		ssl_skip_verify  = true
	  
	`
	parser := hclparse.NewParser()
	f, _ := parser.ParseHCL([]byte(config), "config.hcl")
	cfgVal, _ := hcldec.Decode(f.Body, spec, &hcl.EvalContext{})
	configType := hcldec.ImpliedType(spec)
	fmt.Printf("%v %v\n", configType, cfgVal)

	configureResponse := provider.ConfigureProvider(providers.ConfigureProviderRequest{
		TerraformVersion: "999.0.0",
		Config:           cfgVal,
	})

	if configureResponse.Diagnostics.HasErrors() {
		fmt.Println("Error:", configureResponse.Diagnostics.Err().Error())
		os.Exit(1)
	}

	type ARecordConfig struct {
		Host  string   `cty:"host"`
		Addrs []string `cty:"addrs"`
		ID    string   `cty:"id"`
	}
	type TFEProject struct {
		Name         string   `cty:"name"`
		Organization string   `cty:"organization"`
		ID           string   `cty:"id"`
		WorkspaceIDs []string `cty:"workspace_ids"`
	}
	// readConfig := ARecordConfig{Host: "192.168.1.1.traefik.me"}
	readConfig := TFEProject{Name: "my-project-name", Organization: "my-org-name"}
	readConfigType, _ := gocty.ImpliedType(readConfig)
	readConfigVal, _ := gocty.ToCtyValue(readConfig, readConfigType)
	readResponse := provider.ReadDataSource(providers.ReadDataSourceRequest{
		TypeName:     "tfe_project",
		Config:       readConfigVal,
		ProviderMeta: cty.EmptyObjectVal,
	})

	var results ARecordConfig
	gocty.FromCtyValue(readResponse.State, &results)
	fmt.Printf("%v\n", readResponse)

	// resp, err := provider.ConfigureProvider(ConfigureProviderRequest{
	// 	TerraformVersion: "1.2.3.4",
	// 	Config: cty.MapVal(map[string]cty.Value{
	// 		"update": cty.ListVal(
	// 			[]cty.Value{
	// 				cty.MapVal(map[string]cty.Value{"port2": cty.NumberIntVal(0)}),
	// 			},
	// 		)}),
	// })
	// fmt.Println(resp, err)
}

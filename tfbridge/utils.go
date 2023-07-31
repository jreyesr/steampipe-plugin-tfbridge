package tfbridge

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/jreyesr/steampipe-plugin-tfbridge/logging"
	tfplugin "github.com/jreyesr/steampipe-plugin-tfbridge/plugin"
	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
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
		// Cmd:              exec.Command("sh", "-c", "./terraform-provider-dns_v3.2.4_x5"),
		// Cmd:              exec.Command("sh", "-c", "./terraform-provider-tfe_v0.47.0_x5"),
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

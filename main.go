package main

import (
	"github.com/jreyesr/steampipe-plugin-tfbridge/tfbridge"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: tfbridge.Plugin})
}

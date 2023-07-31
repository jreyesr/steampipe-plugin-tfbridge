package tfbridge

import (
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type TFBridgeConfig struct {
	Provider       *string `cty:"provider"`
	Version        *string `cty:"version"`
	ProviderConfig *string `cty:"provider_config"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"provider":        {Type: schema.TypeString},
	"version":         {Type: schema.TypeString},
	"provider_config": {Type: schema.TypeString},
}

func ConfigInstance() interface{} {
	return &TFBridgeConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) TFBridgeConfig {
	if connection == nil || connection.Config == nil {
		return TFBridgeConfig{}
	}
	config, _ := connection.Config.(TFBridgeConfig)
	return config
}

func (c TFBridgeConfig) String() string {
	return fmt.Sprintf(
		"TFBridgeConfig{provider=%s v%s, other_config=%v}",
		*c.Provider, *c.Version, *c.ProviderConfig)
}

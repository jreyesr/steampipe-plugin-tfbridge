package tfbridge

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             "steampipe-plugin-tfbridge",
		DefaultTransform: transform.FromGo().NullIfZero(),
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		SchemaMode:   plugin.SchemaModeDynamic,
		TableMapFunc: PluginTables,
	}
	return p
}

type key string

const (
	keyDataSource key = "dataSource"
	keySchema     key = "schema"
)

func PluginTables(ctx context.Context, d *plugin.TableMapData) (map[string]*plugin.Table, error) {
	// Initialize tables
	tables := map[string]*plugin.Table{}

	config := GetConfig(d.Connection)

	// Download requested provider to tempdir
	pluginBinaryPath, err := DownloadProvider(ctx, *config.Provider, *config.Version, d)
	if err != nil {
		plugin.Logger(ctx).Error("tfbridge.PluginTables", "download_provider_error", err, "provider", *config.Provider)
		return nil, err
	}
	plugin.Logger(ctx).Info("tfbridge.PluginTables", "plugin_download_path", pluginBinaryPath)

	// Establish connection with downloaded provider
	conn, err := getPluginConnection(pluginBinaryPath)
	if err != nil {
		plugin.Logger(ctx).Error("tfbridge.PluginTables", "get_connection_error", err, "provider", *config.Provider)
		return nil, err
	}

	dataSources, err := getDataSources(conn)
	if err != nil {
		plugin.Logger(ctx).Error("tfbridge.PluginTables", "get_data_sources_error", err)
		return nil, err
	}
	plugin.Logger(ctx).Debug("tfbridge.PluginTables.getDataSources", "ds", dataSources)
	for k, i := range dataSources {
		// Nested WithValue: set two keys on the same context
		tableCtx := context.WithValue(context.WithValue(ctx, keyDataSource, k), keySchema, i)
		table, err := tableTFBridge(tableCtx, d.Connection, pluginBinaryPath)
		if err != nil {
			plugin.Logger(ctx).Error("tfbridge.PluginTables", "create_table_error", err, "datasource", k)
			return nil, err
		}

		plugin.Logger(ctx).Debug("tfbridge.PluginTables.makeTables", "name", k, "table", table)
		tables[k] = table
	}
	plugin.Logger(ctx).Debug("tfbridge.PluginTables.makeTables", "tables", tables)
	// paths, err := csvList(ctx, p)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, i := range paths {
	// 	tableCtx := context.WithValue(ctx, "path", i)
	// 	base := filepath.Base(i)
	// 	// tableCSV returns a *plugin.Table type
	// 	tables[base[0:len(base)-len(filepath.Ext(base))]] = tableCSV(tableCtx, p)
	// }

	return tables, nil
}

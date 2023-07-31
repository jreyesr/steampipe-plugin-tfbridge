package tfbridge

import (
	"context"
	"fmt"

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
		// TableMap: map[string]*plugin.Table{
		// 	"tfbridge_ds_1": tableTFBridgeZone(),
		// },
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

	// Search for CSV files to create as tables
	config := GetConfig(d.Connection)
	conn, err := getPluginConnection(*config.Provider)
	if err != nil {
		plugin.Logger(ctx).Error("tfbridge.PluginTables", "get_connection_error", err, "provider", *config.Provider)
		return nil, err
	}

	dataSources, err := getDataSources(conn)
	if err != nil {
		plugin.Logger(ctx).Error("tfbridge.PluginTables", "get_data_sources_error", err)
		return nil, err
	}
	for k, i := range dataSources {
		fmt.Printf("%s %v", k, i)
		// Nested WithValue: set two keys on the same context
		tableCtx := context.WithValue(context.WithValue(ctx, keyDataSource, k), keySchema, i)
		table, err := tableTFBridge(tableCtx, d.Connection)
		if err != nil {
			plugin.Logger(ctx).Error("tfbridge.PluginTables", "create_table_error", err, "datasource", k)
			return nil, err
		}

		tables[k] = table
	}
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

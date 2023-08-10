package tfbridge

import (
	"context"
	"fmt"

	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
)

func tableTFBridge(ctx context.Context, connection *plugin.Connection, pluginLocation string) (*plugin.Table, error) {
	name := ctx.Value(keyDataSource).(string)
	schema := ctx.Value(keySchema).(providers.Schema)

	return &plugin.Table{
		Name:        name,
		Description: fmt.Sprintf("%s: %s", name, schema.Block.Description),
		List: &plugin.ListConfig{
			Hydrate:    ListDataSource(name, pluginLocation),
			KeyColumns: makeKeyColumns(ctx, schema),
		},
		Columns: makeColumns(ctx, schema),
	}, nil

	// plugin.Logger(ctx).Error("tfbridge.PluginTables", "create_table_error", err, "datasource", k)
}

func makeColumns(ctx context.Context, schema providers.Schema) []*plugin.Column {
	columns := []*plugin.Column{}

	// First the attributes (atomic/leaf params, with no nested business)
	for k, i := range schema.Block.Attributes {
		columns = append(columns, &plugin.Column{
			Name:        k,
			Type:        attrTypeToColumnType(ctx, i.Type),
			Description: i.Description,
			Transform:   FromCtyMapKey(k),
		})
	}

	// Now the nested blocks, they will be JSON no questions asked
	for k, i := range schema.Block.BlockTypes {
		columns = append(columns, &plugin.Column{
			Name:        k,
			Type:        proto.ColumnType_JSON,
			Description: i.Description,
			Transform:   FromCtyMapKey(k),
		})
	}

	return columns
}

func makeKeyColumns(ctx context.Context, schema providers.Schema) plugin.KeyColumnSlice {
	mandatoryKeyColumns := []string{}
	optionalKeyColumns := []string{}

	for k, i := range schema.Block.Attributes {
		if childAttributeIsRequired(i) {
			mandatoryKeyColumns = append(mandatoryKeyColumns, k)
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "mandatory")
		} else if childAttributeIsOptional(i) {
			optionalKeyColumns = append(optionalKeyColumns, k)
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "optional")
		} else if childAttributeIsReadOnly(i) {
			// Read-only attrs don't become KeyColumns
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "ignore")
		} else {
			// should never happen, right?
			plugin.Logger(ctx).Error("makeKeyColumns", "column", k, "data", i, "disposition", "INVALID")
		}
	}

	// Remember that all nested blocks become JSONB columns on Steampipe, no matter what
	for k, i := range schema.Block.BlockTypes {
		if childBlockIsRequired(i) {
			mandatoryKeyColumns = append(mandatoryKeyColumns, k)
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "mandatory")
		} else if childBlockIsOptional(i) {
			optionalKeyColumns = append(optionalKeyColumns, k)
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "optional")
		} else if childBlockIsReadOnly(i) {
			// Read-only nested blocks don't become KeyColumns
			plugin.Logger(ctx).Debug("makeKeyColumns", "column", k, "data", i, "disposition", "ignore")
		} else {
			// should never happen, right?
			plugin.Logger(ctx).Error("makeKeyColumns", "column", k, "data", i, "disposition", "INVALID")
		}
	}

	var all = make([]*plugin.KeyColumn, 0, len(mandatoryKeyColumns)+len(optionalKeyColumns))
	for _, c := range mandatoryKeyColumns {
		all = append(all, &plugin.KeyColumn{
			Name:      c,
			Operators: []string{"="},
			Require:   plugin.Required, // Magic is here
		})
	}
	for _, c := range optionalKeyColumns {
		all = append(all, &plugin.KeyColumn{
			Name:      c,
			Operators: []string{"="},
			Require:   plugin.Optional, // Magic is here
		})
	}

	plugin.Logger(ctx).Info("makeKeyColumns.done", "len", len(all), "mand", mandatoryKeyColumns, "opt", optionalKeyColumns, "val", all)
	return all
}

func attrTypeToColumnType(ctx context.Context, attrType cty.Type) proto.ColumnType {
	switch attrType {
	case cty.Number:
		return proto.ColumnType_DOUBLE
	case cty.String:
		return proto.ColumnType_STRING
	case cty.Bool:
		return proto.ColumnType_BOOL
	default:
		// fear the unknown, cast as JSON
		// this catches tuple, list, set (array-likes), object, map (dict-likes) and probably others?
		// plz no capsule types
		plugin.Logger(ctx).Warn("tfbridge.attrTypeToColumnType", "unknown_type_on_attr", attrType.FriendlyName())
		return proto.ColumnType_JSON
	}
}

func ListDataSource(name, pluginLocation string) func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		config := GetConfig(d.Connection)

		plugin.Logger(ctx).Info("tfbridge.ListDataSource", "equalsQuals", d.EqualsQuals)
		plugin.Logger(ctx).Info("tfbridge.ListDataSource", "location", pluginLocation)
		conn, err := getPluginConnection(pluginLocation)
		if err != nil {
			plugin.Logger(ctx).Warn("tfbridge.ListDataSource.getPluginConnection", "provider", *config.Provider, "err", err)
			return nil, err
		}
		err = configureProvider(ctx, conn, *config.ProviderConfig)
		if err != nil {
			plugin.Logger(ctx).Warn("tfbridge.ListDataSource.configureProvider", "provider", *config.Provider, "config", config.ProviderConfig, "err", err)
			return nil, err
		}

		response, err := readDataSource(ctx, conn, name, d.EqualsQuals)
		if err != nil {
			plugin.Logger(ctx).Warn("tfbridge.ListDataSource.readDataSource", "name", name)
			return nil, err
		}
		responseMap := response.AsValueMap()
		plugin.Logger(ctx).Info("tfbridge.ListDataSource.response", "data", responseMap)
		d.StreamListItem(ctx, responseMap)

		return nil, nil
	}
}

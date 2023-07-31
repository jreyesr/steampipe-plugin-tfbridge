package tfbridge

import (
	"context"
	"fmt"

	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
)

func tableTFBridge(ctx context.Context, connection *plugin.Connection) (*plugin.Table, error) {
	name := ctx.Value(keyDataSource).(string)
	schema := ctx.Value(keySchema).(providers.Schema)

	return &plugin.Table{
		Name:        name,
		Description: fmt.Sprintf("%s: %s", name, schema.Block.Description),
		List: &plugin.ListConfig{
			Hydrate:    ListDataSource(name),
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
			// Transform: ???,
		})
	}

	// Now the nested blocks, they will be JSON no questions asked
	for k, i := range schema.Block.BlockTypes {
		columns = append(columns, &plugin.Column{
			Name:        k,
			Type:        proto.ColumnType_JSON,
			Description: i.Description,
			// Transform: ???,
		})
	}

	return columns
}

func makeKeyColumns(ctx context.Context, schema providers.Schema) plugin.KeyColumnSlice {
	keyColumns := []string{}

	// TODO How does the TF doc generator detect Arguments vs. Other Fields???
	for k, i := range schema.Block.Attributes {
		if i.Required && !i.Computed {
			keyColumns = append(keyColumns, k)
		}
	}
	for k, i := range schema.Block.BlockTypes {
		if i.MinItems > 0 {
			keyColumns = append(keyColumns, k)
		}
	}

	var all = make([]*plugin.KeyColumn, len(keyColumns))
	for i, c := range keyColumns {
		all[i] = &plugin.KeyColumn{
			Name:      c,
			Operators: []string{"="},
			Require:   plugin.Required,
		}
	}
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
		plugin.Logger(ctx).Error("tfbridge.attrTypeToColumnType", "unknown_type_on_attr", attrType.FriendlyName())
		return proto.ColumnType_JSON
	}
}

func ListDataSource(path string) func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
		plugin.Logger(ctx).Info("tfbridge.ListDataSource", "quals", d.Quals)
		return nil, nil
	}
}

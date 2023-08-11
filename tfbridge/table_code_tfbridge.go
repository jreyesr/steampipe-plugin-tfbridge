package tfbridge

import (
	"context"
	"fmt"
	"sort"

	"github.com/jreyesr/steampipe-plugin-tfbridge/providers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/slices"
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
}

func makeColumns(ctx context.Context, schema providers.Schema) []*plugin.Column {
	columns := []*plugin.Column{}
	columnsAttrs := make([]string, 0, len(schema.Block.Attributes))
	columnsBlocks := make([]string, 0, len(schema.Block.BlockTypes))

	// First the attributes (atomic/leaf params, with no nested business)
	for k, i := range schema.Block.Attributes {
		postgresType := attrTypeToColumnType(ctx, i.Type, i.NestedType.ImpliedType())
		if postgresType == proto.ColumnType_UNKNOWN {
			plugin.Logger(ctx).Warn("tfbridge.makeColumns.atomic", "msg", "unknown type, skipping column!", "field", k, "type", i.Type)
			continue
		}
		columns = append(columns, &plugin.Column{
			Name:        k,
			Type:        postgresType,
			Description: i.Description,
			Transform:   FromCtyMapKey(k),
		})
		columnsAttrs = append(columnsAttrs, k)
	}

	// Now the nested blocks, they will be JSON no questions asked
	for k, i := range schema.Block.BlockTypes {
		columns = append(columns, &plugin.Column{
			Name:        k,
			Type:        proto.ColumnType_JSON,
			Description: i.Description,
			Transform:   FromCtyMapKey(k),
		})
		columnsBlocks = append(columnsBlocks, k)
	}

	// this bit is required for sorting like the TF docs do, see below
	// colTypes := {"id": "readonly", "name": "required", "type": "optional", "othercol": "readonly", ...}
	// map keys are the column names, which match datasource's attributes
	// the only valid values are "required", "optional" and "readonly"
	colTypes := make(map[string]string)
	for _, col := range columns {
		var group string
		if slices.Contains(columnsAttrs, col.Name) { // it's an atomic attr
			if childAttributeIsRequired(schema.Block.Attributes[col.Name]) {
				group = "required"
			} else if childAttributeIsOptional(schema.Block.Attributes[col.Name]) {
				group = "optional"
			} else if childAttributeIsReadOnly(schema.Block.Attributes[col.Name]) {
				group = "readonly"
			}
		} else if slices.Contains(columnsBlocks, col.Name) { // it's a nested block (complex attr)
			if childBlockIsRequired(schema.Block.BlockTypes[col.Name]) {
				group = "required"
			} else if childBlockIsOptional(schema.Block.BlockTypes[col.Name]) {
				group = "optional"
			} else if childBlockIsReadOnly(schema.Block.BlockTypes[col.Name]) {
				group = "readonly"
			}
		}
		if group == "" { // we didn't hit any of the cases above, should be a bug
			panic(fmt.Errorf("column %v wasn't found on either Attrs or Blocks", col))
		}
		colTypes[col.Name] = group
	}

	// as courtesy to the user, sort the columns in required -> optional -> readonly, and inside each group sort alphabetically
	// this pays no attention to atomic attrs vs. nested/complex attrs
	sort.SliceStable(columns, func(i, j int) bool {
		groupOrder := map[string]int{"required": 0, "optional": 1, "readonly": 2}
		colTypeI, colTypeJ := colTypes[columns[i].Name], colTypes[columns[j].Name]
		if colTypeI != colTypeJ {
			// cols [i] and [j] are on different main groups, so that takes precedence
			return groupOrder[colTypeI] < groupOrder[colTypeJ]
		}
		// otherwise, we know that cols [i] and [j] belong to same group, so we look at their names
		return columns[i].Name < columns[j].Name
	})

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

func attrTypeToColumnType(ctx context.Context, attrType cty.Type, nestedAttrType cty.Type) proto.ColumnType {
	plugin.Logger(ctx).Debug("tfbridge.attrTypeToColumnType", "type", attrType)

	switch attrType {
	case cty.Number:
		return proto.ColumnType_DOUBLE
	case cty.String:
		return proto.ColumnType_STRING
	case cty.Bool:
		return proto.ColumnType_BOOL
	case cty.NilType: // could indicate nested, since those have Type = nil and NestedType not nil
		if nestedAttrType != cty.NilType {
			return proto.ColumnType_JSON
		}
		// if both Type and NestedType are nil, should be a bug somewhere, right?
		return proto.ColumnType_UNKNOWN
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

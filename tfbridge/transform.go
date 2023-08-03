package tfbridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"
)

/*
ctyValToSteampipeVal takes a row, returned from the List hydrate func, and converts it into Steampipe-friendly Go native types
The values streamed from List are cty.Value instances, which are inspected and converted to Go types
Mapping rules are:

* cty.Numbers are converted to float64
* cty.Strings are converted to string
* cty.Bools are converted to bool
* cty.List(ANYTHING), cty.Set(ANYTHING), cty.Tuple(ANYTHING) are converted to slices []any
* cty.Map(ANYTHING), cty.Object are converted to maps map[string]any
*/
func ctyValToSteampipeVal(ctx context.Context, tf *transform.TransformData) (interface{}, error) {
	entireItem := tf.HydrateItem.(map[string]cty.Value)
	key := tf.Param.(string)
	plugin.Logger(ctx).Info("ctyValToSteampipeVal", "k", key, "item", entireItem)

	val, ok := entireItem[key]
	if !ok {
		return nil, fmt.Errorf("cty.ValueAsMap %v has no field %s", entireItem, key)
	}

	// in this switch, the primary thing that changes is the type of x
	// then gocty.FromCtyValue saves into x and x is returned
	switch {
	// primitive types are easy
	case val.Type() == cty.Number:
		var x float64
		gocty.FromCtyValue(val, &x)
		return x, nil
	case val.Type() == cty.String:
		var x string
		gocty.FromCtyValue(val, &x)
		return x, nil
	case val.Type() == cty.Bool:
		var x bool
		gocty.FromCtyValue(val, &x)
		return x, nil
	// array-like types, save into generic-est slice
	case val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType():
		asJsonList, err := ctyJson.SimpleJSONValue{Value: val}.MarshalJSON()
		if err != nil {
			return nil, err
		}

		var x []any
		json.Unmarshal(asJsonList, &x)
		return x, nil
	// map-like types, save into generic-est map
	case val.Type().IsMapType() || val.Type().IsTupleType():
		asJson, err := ctyJson.SimpleJSONValue{Value: val}.MarshalJSON()
		if err != nil {
			return nil, err
		}

		var x map[string]any
		json.Unmarshal(asJson, &x)
		plugin.Logger(ctx).Info("ctyValToSteampipeVal.json", "asJson", asJson, "x", x)
		return x, nil
	default:
		return nil, fmt.Errorf("value %v with type %v is not recognized", val, val.Type())
	}
}

func FromCtyMapKey(key string) *transform.ColumnTransforms {
	return &transform.ColumnTransforms{Transforms: []*transform.TransformCall{
		{Transform: ctyValToSteampipeVal, Param: key},
	}}
}

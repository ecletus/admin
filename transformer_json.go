package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/aghape/roles"
	"github.com/moisespsena-go/valuesmap"
)

// JSONTransformer json transformer
type JSONTransformer struct{}

// CouldEncode check if encodable
func (JSONTransformer) CouldEncode(encoder Encoder) bool {
	return true
}

// Encode encode encoder to writer as JSON
func (JSONTransformer) Encode(writer io.Writer, encoder Encoder) error {
	var (
		context = encoder.Context
		res     = encoder.Resource
	)

	js, err := json.MarshalIndent(convertObjectToJSONMap(res, context, encoder.Result, encoder.Layout), "", "\t")
	if err != nil {
		result := make(map[string]string)
		result["error"] = err.Error()
		js, _ = json.Marshal(result)
	}

	if w, ok := writer.(http.ResponseWriter); ok {
		w.Header().Set("Content-Type", "application/json")
	}

	_, err = writer.Write(js)
	return err
}

func convertObjectToJSONMap(res *Resource, context *Context, value interface{}, layout string) interface{} {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Slice:
		values := []interface{}{}
		for i := 0; i < reflectValue.Len(); i++ {
			indexValue := reflect.Indirect(reflectValue.Index(i))
			if k := indexValue.Kind(); k == reflect.Struct || k == reflect.Interface {
				if k == reflect.Interface {
					indexValue = reflect.Indirect(indexValue.Elem())
				}
				indexValue = indexValue.Addr()
				values = append(values, convertObjectToJSONMap(res, context, indexValue.Interface(), layout))
			} else {
				values = append(values, fmt.Sprint(indexValue.Interface()))
			}
		}
		return values
	case reflect.Struct:
		if getter, ok := value.(valuesmap.Getter); ok {
			return convertObjectToJSONMap(res, context, getter.Get(), layout)
		}
		metas, metaNames := res.MetasFromLayoutContext(layout, context, value, roles.Read)
		values := map[string]interface{}{}
		for i, meta := range metas {
			// has_one, has_many checker to avoid dead loop
			if meta.Resource != nil && (meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && (meta.FieldStruct.Relationship.Kind == "has_one" || meta.FieldStruct.Relationship.Kind == "has_many" || meta.Type == "single_edit" || meta.Type == "collection_edit")) {
				values[metaNames[i].GetEncodedNameOrDefault()] = convertObjectToJSONMap(meta.Resource, context, context.RawValueOf(value, meta), layout)
			} else {
				formattedValue := context.FormattedValueOf(value, meta)
				if meta.ForceShowZero || !meta.IsZero(value, formattedValue) {
					values[metaNames[i].GetEncodedNameOrDefault()] = formattedValue
				}
			}
		}
		return values
	case reflect.Map:
		for _, key := range reflectValue.MapKeys() {
			reflectValue.SetMapIndex(key, reflect.ValueOf(convertObjectToJSONMap(res, context, reflectValue.MapIndex(key).Interface(), layout)))
		}
		return reflectValue.Interface()
	default:
		return value
	}
}

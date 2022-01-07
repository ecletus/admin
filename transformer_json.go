package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/ecletus/roles"
	"github.com/moisespsena-go/valuesmap"
)

// JSONTransformer json transformer
type JSONTransformer struct{}

var JSONTransformerType = reflect.TypeOf(JSONTransformer{})

// CouldEncode check if encodable
func (JSONTransformer) CouldEncode(*Encoder) bool {
	return true
}

func (JSONTransformer) IsType(t reflect.Type) bool {
	return JSONTransformerType == t
}

// Encode encode encoder to writer as JSON
func (JSONTransformer) Encode(writer io.Writer, encoder *Encoder) (err error) {
	var (
		context = encoder.Context
		res     = encoder.Resource
	)

	var js []byte

	if encoder.Result == nil {
		js = []byte("null")
	} else if js, err = json.MarshalIndent(convertObjectToJSONMap(res, context, encoder.Result, encoder.Layout), "", "\t"); err != nil {
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

func convertObjectToJSONMap(res *Resource, ctx *Context, value interface{}, layout string) interface{} {
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
				values = append(values, convertObjectToJSONMap(res, ctx.CreateChild(res, indexValue.Interface()), indexValue.Interface(), layout))
			} else {
				values = append(values, fmt.Sprint(indexValue.Interface()))
			}
		}
		return values
	case reflect.Struct:
		if getter, ok := value.(valuesmap.Getter); ok {
			return convertObjectToJSONMap(res, ctx, getter.Get(), layout)
		}
		metas, metaNames := res.MetasFromLayoutNameContext(layout, ctx, value, roles.Read)
		var (
			values = map[string]interface{}{}
			val    interface{}
		)
		for i, meta := range metas {
			pop := ctx.MetaStack.Push(meta)
			fv := meta.FormattedValue(ctx.Context, value)
			if fv == nil {
				continue
			}

			// has_one, has_many checker to avoid dead loop
			if meta.Resource != nil && (meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && (meta.FieldStruct.Relationship.Kind.IsHasN() || meta.Type == "single_edit" || meta.Type == "collection_edit")) {
				if val := convertObjectToJSONMap(meta.Resource, ctx, fv.Raw, layout); val == nil {
					fv = res.GetDefinedMeta(META_STRINGIFY).FormattedValue(ctx.Context, val)
					if fv == nil {
						continue
					}
				}
				val = fv.Raw
			} else if fv.Value == "" && fv.SafeValue != "" {
				val = fv.SafeValue
			} else {
				val = fv.Value
			}
			if meta.ForceShowZero || !meta.IsZero(value, val) {
				values[metaNames[i].GetEncodedNameOrDefault()] = val
			}
			pop()
		}
		if len(values) == 0 {
			return nil
		}
		return values
	case reflect.Map:
		for _, key := range reflectValue.MapKeys() {
			reflectValue.SetMapIndex(key, reflect.ValueOf(convertObjectToJSONMap(res, ctx, reflectValue.MapIndex(key).Interface(), layout)))
		}
		return reflectValue.Interface()
	default:
		return value
	}
}

package admin

import (
	"reflect"

	"github.com/moisespsena-go/aorm"
)

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &aorm.Scope{Value: res.Value}

Fields:
	for _, field := range scope.GetModelStruct().StructFields {
		for _, meta := range res.Metas {
			if field.Name == meta.FieldName {
				attrs = append(attrs, meta.Name)
				continue Fields
			}
		}

		if field.IsForeignKey {
			continue
		}

		for _, value := range []string{"CreatedAt", "UpdatedAt", "DeletedAt"} {
			if value == field.Name {
				continue Fields
			}
		}

		if (field.IsNormal || field.Relationship != nil) && !field.IsIgnored {
			attrs = append(attrs, field.Name)
			continue
		}

		fieldType := field.Struct.Type
		for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			attrs = append(attrs, field.Name)
		}
	}

MetaIncluded:
	for _, meta := range res.Metas {
		if meta.Name[0] != '_' {
			for _, attr := range attrs {
				if attr == meta.FieldName || attr == meta.Name {
					continue MetaIncluded
				}
			}
			attrs = append(attrs, meta.Name)
		}
	}

	return attrs
}

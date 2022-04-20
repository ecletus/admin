package admin

import (
	"reflect"

	"github.com/go-aorm/aorm"
)

func (this *Resource) allAttrs() []string {
	var attrs []string

	for _, field := range aorm.StructOf(this.Value).Fields {
		if meta := this.GetDefinedMeta(field.Name); meta != nil {
			if meta.DefaultInvisible {
				continue
			}
			attrs = append(attrs, meta.Name)
			continue
		} else if tags := ParseMetaTags(field.Struct.Tag); tags.Hidden() || tags.DefaultInvisible() {
			continue
		}

		if field.IsPrimaryKey || field.IsForeignKey || aorm.IsAuditedSdField(field.Name) {
			continue
		}

		if (field.IsNormal || field.Relationship != nil) && (field.IsEmbedded || !field.IsIgnored) {
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
	for _, meta := range this.Metas {
		if !meta.DefaultInvisible && meta.Name[0] != '_' {
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

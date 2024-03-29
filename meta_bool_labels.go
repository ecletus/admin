package admin

import (
	"reflect"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

type BoolLabelsConfig struct {
	Metas []string
	metaConfig
}

func (s *BoolLabelsConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "bool_labels"
	meta.SetFormattedValuer(func(record interface{}, ctx *core.Context) *FormattedValue {
		adminContext := ContextFromCoreContext(ctx)
		var labels []string
		for _, name := range s.Metas {
			m := meta.BaseResource.GetMeta(name)
			v := m.Value(ctx, record)
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case *bool:
				if t != nil && *t {
					labels = append(labels, m.GetRecordLabel(adminContext, record))
				}
			case bool:
				if t {
					labels = append(labels, m.GetRecordLabel(adminContext, record))
				}
			}
		}
		return &FormattedValue{Record: record, Slice: true, Raw: labels, Value: strings.Join(labels, ", "), IsZeroF: func(record, value interface{}) bool {
			return len(value.([]string)) == 0
		}}
	})

	if len(s.Metas) == 0 {
		for _, f := range meta.BaseResource.ModelStruct.Fields {
			if f.Struct.Type.Kind() == reflect.Bool || (f.Struct.Type.Kind() == reflect.Ptr && f.Struct.Type.Elem().Kind() == reflect.Bool) {
				s.Metas = append(s.Metas, f.Name)
			}
		}
	}
}

func init() {
	RegisterMetaConfigor("bool_labels", func(meta *Meta) {
		if meta.Config == nil {
			cfg := &BoolLabelsConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	})
}

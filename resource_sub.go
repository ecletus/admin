package admin

import (
	"fmt"
	"reflect"
)

type SubSliceFieldConfig struct {
	*Config
	NotMeta bool
	Meta    *Meta
}

func (res *Resource) AddSliceField(fieldName string, config ...*SubSliceFieldConfig) *Resource {
	sliceField, ok := reflect.TypeOf(res.Value).Elem().FieldByName(fieldName)
	if !ok {
		panic(fmt.Errorf("%T does not have field %q", res.Value, fieldName))
	}
	if sliceField.Type.Kind() != reflect.Slice {
		panic(fmt.Errorf("%T.%s is not slice", res.Value, fieldName))
	}
	fieldItemType := sliceField.Type.Elem()
	for fieldItemType.Kind() == reflect.Ptr {
		fieldItemType = fieldItemType.Elem()
	}
	fieldItemValue := reflect.New(fieldItemType).Interface()

	var subCfg *SubSliceFieldConfig
	if len(config) >= 0 && config[0] != nil {
		subCfg = config[0]
	} else {
		subCfg = &SubSliceFieldConfig{}
	}

	cfg := subCfg.Config
	if cfg == nil {
		cfg = &Config{}
	}

	var meta *Meta
	if !subCfg.NotMeta {
		if subCfg.Meta == nil {
			meta = &Meta{}
		}
		if meta.Name == "" {
			meta.Name = fieldName
		}
	}

	oldSetup := cfg.Setup
	cfg.Setup = func(sub *Resource) {
		if meta != nil {
			res.SetMeta(meta)
		}
		if oldSetup != nil {
			oldSetup(sub)
		}
	}
	if cfg.LabelKey == "" {
		cfg.LabelKey = res.ChildrenLabelKey(fieldName)
	}

	if !res.registered {
		res.afterRegister = append(res.afterRegister, func() {
			res.AddResource(&SubConfig{FieldName: fieldName}, fieldItemValue, cfg)
		})
		return nil
	}

	return res.AddResource(&SubConfig{FieldName: fieldName}, fieldItemValue, cfg)
}

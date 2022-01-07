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

func (this *Resource) AddSliceField(fieldName string, config ...*SubSliceFieldConfig) *Resource {
	sliceField, ok := reflect.TypeOf(this.Value).Elem().FieldByName(fieldName)
	if !ok {
		panic(fmt.Errorf("%T does not have field %q", this.Value, fieldName))
	}
	if sliceField.Type.Kind() != reflect.Slice {
		panic(fmt.Errorf("%T.%s is not slice", this.Value, fieldName))
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
			this.SetMeta(meta)
		}
		if oldSetup != nil {
			oldSetup(sub)
		}
	}
	if cfg.LabelKey == "" {
		cfg.LabelKey = this.ChildrenLabelKey(fieldName)
	}

	if !this.initialized {
		this.postInitializeCallbacks = append(this.postInitializeCallbacks, func() {
			this.AddResource(&SubConfig{FieldName: fieldName}, fieldItemValue, cfg)
		})
		return nil
	}

	return this.AddResource(&SubConfig{FieldName: fieldName}, fieldItemValue, cfg)
}

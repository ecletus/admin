package admin

import (
	"reflect"

	"github.com/ecletus/core/utils"
)

// NewResource initialize a new ecletus resource, won't add it to admin, just initialize it
func (this *Resource) NewResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = this
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}

	return this.Admin.NewResource(value, config[0])
}

// AddResource register sub-resource with optional config into admin
func (this *Resource) AddResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = this
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}
	return this.AddResourceConfig(value, config[0])
}

// AddResourceConfig register sub-resource with config into admin
func (this *Resource) AddResourceConfig(value interface{}, cfg *Config) *Resource {
	if cfg.Sub == nil {
		cfg.Sub = &SubConfig{}
	}
	cfg.Sub.Parent = this
	return this.Admin.AddResource(value, cfg)
}

// AddResourceFieldConfig register sub-resource from field type with config into admin. Value is optional.
func (this *Resource) AddResourceFieldConfig(fieldName string, value interface{}, cfg *Config) *Resource {
	if value == nil {
		field, _ := utils.IndirectType(this.Value).FieldByName(fieldName)
		fieldType := utils.IndirectType(field.Type)

		if fieldType.Kind() == reflect.Slice {
			fieldType = utils.IndirectType(fieldType.Elem())
		}

		value = reflect.New(fieldType).Interface()
	}
	setup := cfg.Setup
	cfg.Setup = func(child *Resource) {
		this.SetMeta(&Meta{Name: fieldName, Resource: child})
		if setup != nil {
			setup(child)
		}
	}
	defer func() {
		this.Meta(&Meta{Name: fieldName, Resource: this})
	}()
	return this.AddResource(&SubConfig{FieldName: fieldName}, value, cfg)
}

// AddResourceField register sub-resource from field type into admin. Value and setup function is optional.
func (this *Resource) AddResourceField(fieldName string, value interface{}, setup ...func(res *Resource)) *Resource {
	return this.AddResourceFieldConfig(fieldName, value, &Config{
		Setup: func(res *Resource) {
			for _, s := range setup {
				s(res)
			}
		},
	})
}

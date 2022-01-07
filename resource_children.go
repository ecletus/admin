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

// NewResource initialize a new ecletus resource, won't add it to admin, just initialize it
func (this *Resource) NewActionResource(value interface{}, config ...*Config) *Resource {
	var (
		cfg  *Config
		name = indirectType(reflect.TypeOf(value)).Name()
	)
	for _, cfg = range config {
	}
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.UID = this.UID + ":actions:" + name
	cfg.ID = "Actions." + name
	cfg.PrependSetup(func(res *Resource) {
		res.I18nPrefix = this.I18nPrefix + ".action_resources." + name
	})
	return this.Admin.NewResource(value, cfg)
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
func (this *Resource) AddResourceFieldConfig(fieldName string, value interface{}, cfg *Config, addMetaDisabled ...bool) (res *Resource) {
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
		if cfg.Permission == nil {
			child.Permissioner(this)
		}
		this.SetMeta(&Meta{Name: fieldName, Resource: child})
		if setup != nil {
			setup(child)
		}
	}
	res = this.AddResource(&SubConfig{FieldName: fieldName}, value, cfg)
	for _, dis := range addMetaDisabled {
		if dis {
			return
		}
	}
	this.Meta(&Meta{Name: fieldName, Resource: res})
	return
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

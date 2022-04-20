package admin

import (
	"reflect"

	"github.com/moisespsena-go/maps"

	"github.com/ecletus/core/resource"
	"github.com/ecletus/roles"
	"github.com/go-aorm/aorm"
)

type MetaConfig struct {
	Name                  string
	EncodedName           string
	Alias                 string
	FieldName             string
	FieldStruct           *aorm.StructField
	ContextResourcer      resource.FContextResourcer
	Setter                resource.FSetter
	Valuer                resource.FValuer
	FormattedValuer       resource.FFormattedValuer
	Config                MetaConfigInterface
	BaseResource          *Resource
	Resource              *Resource
	Permission            *roles.Permission
	Help                  string
	HelpLong              string
	SaveID                bool
	Inline                bool
	Required              bool
	Icon                  bool
	Typ                   reflect.Type
	Type                  string
	TypeHandler           func(meta *Meta, recorde interface{}, context *Context) string
	Enabled               MetaEnabled
	SkipDefaultLabel      bool
	DefaultLabel          string
	Label                 string
	GetMetasFunc          func() []resource.Metaor
	Collection            interface{}
	Dependency            []interface{}
	Fragment              *Fragment
	Options               maps.Map
	OutputFormattedValuer MetaOutputValuer
	DefaultValueFunc      MetaValuer
}

func (this *MetaConfig) Meta(m *Meta) *Meta {
	if m == nil {
		m = &Meta{
			Meta: &resource.Meta{
				FieldName:        this.FieldName,
				Setter:           this.Setter,
				Valuer:           this.Valuer,
				FormattedValuer:  this.FormattedValuer,
				BaseResource:     this.BaseResource,
				ContextResourcer: this.ContextResourcer,
				Resource:         this.Resource,
				Permission:       this.Permission,
				Config:           this.Config,
				Required:         this.Required,
				Inline:           this.Inline,
				Icon:             this.Icon,
				Typ:              this.Typ,
			},
		}
	}
	if this.EncodedName != "" {
		m.EncodedName = this.EncodedName
	}

	if this.Type != "" {
		m.Type = this.Type
	}

	if this.TypeHandler != nil {
		m.TypeHandler = this.TypeHandler
	}

	if this.Enabled != nil {
		m.Enabled = this.Enabled
	}

	if this.SkipDefaultLabel {
		m.SkipDefaultLabel = true
	}

	if this.DefaultLabel != "" {
		m.DefaultLabel = this.DefaultLabel
	}

	if this.Label != "" {
		m.Label = this.Label
	}

	if this.FieldName != "" {
		m.FieldName = this.FieldName
	}

	if this.Setter != nil {
		m.Setter = this.Setter
	}

	if this.Valuer != nil {
		m.Valuer = this.Valuer
	}

	if this.FormattedValuer != nil {
		m.FormattedValuer = this.FormattedValuer
	}

	if this.Resource != nil {
		m.Resource = this.Resource
	}

	if this.Permission != nil {
		m.Permission = this.Permission
	}

	if this.Config != nil {
		m.Config = this.Config
	}

	if this.Collection != nil {
		m.Collection = this.Collection
	}

	if len(this.Dependency) > 0 {
		m.Dependency = this.Dependency
	}

	if this.Fragment != nil {
		m.Fragment = this.Fragment
	}

	if this.Options != nil {
		m.Options.Update(this.Options)
	}

	if this.Alias != "" {
		if m.Alias == nil {
			m.Alias = &resource.MetaName{
				Name:        this.Alias,
				EncodedName: this.Name,
			}
		} else {
			m.Alias.Name = this.Alias
		}
	}
	return m
}

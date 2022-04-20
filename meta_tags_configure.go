package admin

import (
	"fmt"
	"reflect"

	"github.com/ecletus/core"
	"github.com/go-aorm/aorm"
)

var MetaConfigureTagsHandlers []func(meta *Meta, tags *MetaTags)

func RegisterMetaConfigureTagsHandler(f func(meta *Meta, tags *MetaTags)) {
	MetaConfigureTagsHandlers = append(MetaConfigureTagsHandlers, f)
}

func (this *Meta) tagsConfigure() {
	if this.tagsInitialized {
		return
	}
	this.tagsInitialized = true

	this.UITags = map[string]string{}
	if this.FieldStruct == nil {
		return
	}
	var tags = ParseMetaTags(this.FieldStruct.Tag)

	if this.FieldStruct.Struct.Type.Implements(reflect.TypeOf((*DefaultFieldMetaTagger)(nil)).Elem()) {
		for key, value := range reflect.New(IndirectRealType(this.FieldStruct.Struct.Type)).Interface().(DefaultFieldMetaTagger).AdminDefaultMetaTags(this.FieldStruct, tags) {
			if _, ok := tags.Tags[key]; !ok {
				tags.Tags[key] = value
			}
		}
	}

	if this.FieldStruct.Struct.Type.Implements(reflect.TypeOf((*DefaultMetaTagger)(nil)).Elem()) {
		for key, value := range reflect.New(IndirectRealType(this.FieldStruct.Struct.Type)).Interface().(DefaultMetaTagger).AdminDefaultMetaTags() {
			if _, ok := tags.Tags[key]; !ok {
				tags.Tags[key] = value
			}
		}
	}

	if tags.Empty() {
		return
	}
	this.UITags = tags.UI()

	if !this.HiddenLabel {
		this.HiddenLabel = tags.Flag("-LABEL")
	}

	if this.Label == "" {
		this.Label = tags.Label()
	}
	if this.DefaultFormat == "" {
		this.DefaultFormat = tags.Fmt()
	}
	if this.Type == "" {
		this.Type = tags.Type()
	}
	if tags.DefaultInvisible() {
		this.DefaultInvisible = true
	}
	if tags.Required() {
		this.Meta.Required = true
		this.Required = true
	}
	if tags.Readonly() {
		this.ReadOnly = true
	}
	if tags.ReadonlyStringer() {
		this.ReadOnlyStringer = true
	}
	if tags.NilAsZero() {
		this.NilAsZero = true
	}
	if sv := tags.Severity(); sv != "" {
		this.Severity.Parse(sv)
	}
	if lockedField := tags.LockedField(); lockedField != "" {
		typ := indirectType(reflect.TypeOf(this.BaseResource.Value))
		f, ok := typ.FieldByName(lockedField)
		if !ok {
			panic(fmt.Errorf("locked field %s.%s#%s does not exists", typ.PkgPath(), typ.Name(), lockedField))
		}
		fieldIndex := f.Index
		this.LockedFunc = func(_ *Meta, _ *Context, record interface{}) bool {
			return reflect.Indirect(reflect.ValueOf(record)).FieldByIndex(fieldIndex).Bool()
		}
	}
	if tags.Filter() {
		// TODO
	}
	if tags.Sort() {
		// TODO
	}
	if tags.Search() {
		// TODO
	}
	if tags.UnZero() {
		if this.IsZeroFunc == nil {
			this.IsZeroFunc = MetaUnzeroCheck
		}
	}
	if !this.ForceShowZero {
		this.ForceShowZero = tags.ZeroRender()
	}
	if this.Help == "" {
		this.Help = tags.Help()
	}
	if this.Config == nil {
		if cfg, resID, advanced, opt := tags.SelectOne(); cfg != nil {
			if advanced {
				this.Config = cfg
				if resID == "" {
					this.AfterUpdate(func() {
						SelectOneOption(opt, this.BaseResource, NameCallback{Name: this.Name})
					})
				} else {
					this.AfterUpdate(func() {
						this.BaseResource.Admin.OnResourcesAdded(func(e *ResourceEvent) error {
							this.Config = cfg
							this.Resource = e.Resource
							SelectOneOption(opt, this.BaseResource, NameCallback{Name: this.Name})
							return nil
						}, resID)
					})
				}
			} else if resID != "" {
				this.BaseResource.Admin.OnResourcesAdded(func(e *ResourceEvent) error {
					this.Config = cfg
					this.Resource = e.Resource
					this.updateMeta()
					return nil
				}, resID)
			} else {
				this.Config = cfg
			}
		}

		if cfg, resID, advanced, opt := tags.SelectMany(); cfg != nil {
			if advanced {
				if resID == "" {
					this.AfterUpdate(func() {
						SelectManyOption(opt, this.BaseResource, NameCallback{Name: this.Name})
					})
				} else {
					this.AfterUpdate(func() {
						this.BaseResource.Admin.OnResourcesAdded(func(e *ResourceEvent) error {
							this.Config = cfg
							this.Resource = e.Resource
							SelectManyOption(opt, this.BaseResource, NameCallback{Name: this.Name})
							return nil
						}, resID)
					})
				}
			} else if resID != "" {
				this.BaseResource.Admin.OnResourcesAdded(func(e *ResourceEvent) error {
					this.Config = cfg
					this.Resource = e.Resource
					this.updateMeta()
					return nil
				}, resID)
			} else {
				this.Config = cfg
			}
		}
	}

	if edit := tags.Edit(); edit != nil {
		if edit.ReadOnly() {
			this.ReadOnlyFunc = func(meta *Meta, ctx *Context, record interface{}) bool {
				if ctx.Type.Has(EDIT) {
					return true
				}
				return false
			}
		}
	}

	if stringify := tags.GetString("STRINGIFY"); stringify != "" {
		m, ok := reflect.PtrTo(this.BaseResource.ModelStruct.Type).MethodByName(stringify)
		if !ok {
			panic("Tag STRINGIFY: method " + stringify + " does not exists")
		}
		if m.Type.NumOut() != 2 {
			panic("Tag STRINGIFY: method " + stringify + " does not returns (interface{}, string)")
		}
		this.FormattedValuer = func(record interface{}, ctx *core.Context) *FormattedValue {
			values := reflect.ValueOf(record).Method(m.Index).Call(nil)
			if !values[0].IsValid() || values[0].IsNil() || (values[0].Kind() == reflect.Interface && values[0].Elem().IsNil()) {
				return nil
			}
			return (&FormattedValue{Record: record, Raw: values[0].Interface(), Value: values[1].String()}).SetNonZero()
		}
	} else if m, ok := reflect.PtrTo(this.BaseResource.ModelStruct.Type).MethodByName("Get" + this.Name + "String"); ok {
		if m.Type.NumOut() != 2 {
			panic("Tag STRINGIFY: method " + stringify + " does not returns (interface{}, string)")
		}
		this.FormattedValuer = func(record interface{}, ctx *core.Context) *FormattedValue {
			values := reflect.ValueOf(record).Method(m.Index).Call(nil)
			if !values[0].IsValid() || values[0].IsNil() || (values[0].Kind() == reflect.Interface && values[0].Elem().IsNil()) {
				return nil
			}
			return (&FormattedValue{Record: record, Raw: values[0].Interface(), Value: values[1].String()}).SetNonZero()
		}
	}

	for _, f := range MetaConfigureTagsHandlers {
		f(this, &tags)
	}
	this.Tags = tags
}

func MetaUnzeroCheck(m *Meta, record, value interface{}) bool {
	if len(m.BaseResource.PrimaryFields) > 0 {
		return m.BaseResource.GetKey(record).IsZero()
	}
	return aorm.IsZero(value)
}

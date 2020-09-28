package admin

import (
	"fmt"
	"reflect"
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
	if tags.Empty() {
		return
	}
	this.UITags = tags.UI()

	if this.Label == "" {
		this.Label = tags.Label()
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
	if tags.NilAsZero() {
		this.NilAsZero = true
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
	for _, f := range MetaConfigureTagsHandlers {
		f(this, &tags)
	}
	this.Tags = tags
}

package admin

var MetaConfigureTagsHandlers []func(meta *Meta, tags *MetaTags)

func RegisterMetaConfigureTagsHandler(f func(meta *Meta, tags *MetaTags)) {
	MetaConfigureTagsHandlers = append(MetaConfigureTagsHandlers, f)
}

func (this *Meta) tagsConfigure() {
	if this.FieldStruct == nil {
		return
	}
	var tags = ParseMetaTags(this.FieldStruct.Tag)
	if tags.Empty() {
		return
	}
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
		if cfg, resID := tags.SelectOne(); cfg != nil {
			if resID != "" {
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

		if cfg, resID := tags.SelectMany(); cfg != nil {
			if resID != "" {
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

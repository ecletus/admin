package admin

import (
	"github.com/ecletus/core/resource"
)

// CollectionEditConfig meta configuration used for collection edit
type CollectionEditConfig struct {
	Template     string
	TemplateFunc func(context *Context, record interface{}, readOnly bool) (string, error)
	Max          uint
	metaConfig
	meta *Meta

	SectionsAttribute
}

func (this *CollectionEditConfig) BeforeRender1(ctx *MetaContext, record interface{}) {

}

func (this *CollectionEditConfig) GetUserTypeTemplateName(context *Context, record interface{}, readOnly bool) (string, error) {
	if this.TemplateFunc != nil {
		return this.TemplateFunc(context, record, readOnly)
	}
	return this.Template, nil
}

// ConfigureQorMeta configure collection edit meta
func (this *CollectionEditConfig) ConfigureQorMeta(metaor resource.Metaor) {
	this.meta = metaor.(*Meta)
	this.meta.IsCollection = true
	this.Template = metaor.(*Meta).UITags["TEMPLATE"]
	if this.Sections == nil {
		this.SectionsAttribute = this.meta.Resource.SectionsAttribute
	}
}

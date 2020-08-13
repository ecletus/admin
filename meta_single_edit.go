package admin

import (
	"errors"

	"github.com/moisespsena-go/assetfs"

	"github.com/ecletus/core/resource"
)

// SingleEditConfig meta configuration used for single edit
type SingleEditConfig struct {
	Template string
	metaConfig
	meta *Meta
	ExcludeEditAttrs []string
	ExcludeNewAttrs []string
	ExcludeShowAttrs []string
	AfterParseMetaValues func(record interface{}, context *Context)
}

func (this *SingleEditConfig) EditSections(ctx *Context, record interface{}) []*Section {
	var attrs []interface{}
	for _, a := range this.meta.Resource.EditAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeEditAttrs {
		attrs = append(attrs, "-" + a)
	}
	return this.meta.Resource.SectionsList(attrs...)
}

func (this *SingleEditConfig) NewSections(ctx *Context) []*Section {
	var attrs []interface{}
	for _, a := range this.meta.Resource.NewAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeNewAttrs {
		attrs = append(attrs, "-" + a)
	}
	return this.meta.Resource.SectionsList(attrs...)
}

func (this *SingleEditConfig) ShowSections(ctx *Context, record interface{}) []*Section {
	var attrs []interface{}
	for _, a := range this.meta.Resource.ShowSections(ctx, record) {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeShowAttrs {
		attrs = append(attrs, "-" + a)
	}
	return this.meta.Resource.SectionsList(attrs...)
}

// GetTemplate get template for single edit
func (this *SingleEditConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && this.Template != "" {
		return context.Asset(this.Template)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure single edit meta
func (this *SingleEditConfig) ConfigureQorMeta(metaor resource.Metaor) {
	this.meta = metaor.(*Meta)
}

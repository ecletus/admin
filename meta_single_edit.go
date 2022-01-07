package admin

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/moisespsena-go/assetfs"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

type SingleEditConfigSelfPageLink struct {
	Label,
	Uri func(context *Context, record interface{}) string
}

var selfPageLinkCount uint32

// SingleEditConfig meta configuration used for single edit
type SingleEditConfig struct {
	Template string
	metaConfig
	meta                 *Meta
	ExcludeEditAttrs     []string
	ExcludeNewAttrs      []string
	ExcludeShowAttrs     []string
	AfterParseMetaValues func(record interface{}, context *Context)
	showSelfPageLink     *SingleEditConfigSelfPageLink
	selfPageLinkName     string
	SectionLayout        string
}

func (this *SingleEditConfig) ShowSelfPageLink() *SingleEditConfigSelfPageLink {
	return this.showSelfPageLink
}

func (this *SingleEditConfig) SetShowSelfPageLink(showSelfPageLink *SingleEditConfigSelfPageLink) {
	this.showSelfPageLink = showSelfPageLink
	this.selfPageLinkName = "SelfPageLink" + fmt.Sprint(atomic.AddUint32(&selfPageLinkCount, 1))
	if this.meta != nil {
		this.meta.Resource.Meta(&Meta{
			Name:             this.selfPageLinkName,
			DefaultInvisible: true,
			Config: &UrlConfig{
				Copy: true,
				LabelFunc: func(ctx *Context, record interface{}) string {
					return this.showSelfPageLink.Label(ctx, record)
				},
			},
			Valuer: func(record interface{}, context *core.Context) interface{} {
				return this.showSelfPageLink.Uri(ContextFromContext(context), record)
			},
		})
	}
}

func (this *SingleEditConfig) EditSections(ctx *Context, record interface{}) []*Section {
	var attrs []interface{}
	for _, a := range this.meta.Resource.EditAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeEditAttrs {
		attrs = append(attrs, "-"+a)
	}
	return this.meta.Resource.SectionsList(attrs...)
}

func (this *SingleEditConfig) NewSections(ctx *Context) []*Section {
	var attrs []interface{}
	for _, a := range this.meta.Resource.NewAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeNewAttrs {
		attrs = append(attrs, "-"+a)
	}
	return this.meta.Resource.SectionsList(attrs...)
}

func (this *SingleEditConfig) ShowSections(ctx *Context) []*Section {
	var attrs []interface{}
	if this.showSelfPageLink != nil {
		attrs = append(attrs, this.selfPageLinkName)
	}
	for _, a := range this.meta.Resource.ShowSections(ctx, ctx.ResourceRecord) {
		attrs = append(attrs, a)
	}
	for _, a := range this.ExcludeShowAttrs {
		attrs = append(attrs, "-"+a)
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
	if tags := this.meta.Tags.GetTags("SINGLE_EDIT"); tags != nil {
		this.SectionLayout = tags.Get("LAYOUT")
	}
}

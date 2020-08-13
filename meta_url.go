package admin

import (
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type UrlConfig struct {
	Target string
	Download,
	Copy, // Copy to clipboard
	NoLink bool
	Label string

	LabelFunc func(ctx *Context, record interface{}) string
	WrapFunc  func(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML
}

func (this *UrlConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "url"
}

func (this *UrlConfig) GetLabel(ctx *Context, record interface{}) string {
	if this.LabelFunc != nil {
		return this.LabelFunc(ctx, record)
	}
	return this.Label
}

func (this *UrlConfig) Wrap(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML {
	if this.WrapFunc != nil {
		return this.WrapFunc(s, ctx, record, value)
	}
	return value
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &UrlConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("url", cfg)
}

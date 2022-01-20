package admin

import (
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type UrlWithLabel struct {
	Label, Url string
}

type UrlWithLabelConfig struct {
	Target string
	Download,
	Copy, // Copy to clipboard
	NoLink bool

	ReadonlyLabelEnabled bool
	WrapFunc             func(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML
	meta                 *Meta
}

func (this *UrlWithLabelConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if this.meta == nil {
		meta := metaor.(*Meta)
		this.meta = meta
		meta.Type = "url_with_label"
	}
}

func (this *UrlWithLabelConfig) Wrap(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML {
	if this.WrapFunc != nil {
		return this.WrapFunc(s, ctx, record, value)
	}
	return value
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &UrlWithLabelConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("url_with_label", cfg)
}

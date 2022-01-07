package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena/template/html/template"
)

type (
	ResourceMetaUpdater interface {
		AdminMetaUpdate(meta *Meta)
	}

	MetaSingleTyper interface {
		AdminMetaType() string
	}

	MetaTyper interface {
		AdminMetaType(meta *Meta) string
	}

	MetaRecordTyper interface {
		AdminMetaType(meta *Meta, ctx *Context) string
	}

	MetaTemplateNameGetter interface {
		GetTemplateName(context *Context, record interface{}, kind string, readOnly bool) (string, error)
	}

	MetaTemplateExecutorGetter interface {
		GetTemplateExecutor(context *Context, record interface{}, kind string, readOnly bool) (*template.Executor, error)
	}

	MetaUserTypeTemplateNameGetter interface {
		GetUserTypeTemplateName(context *Context, record interface{}, readOnly bool) (string, error)
	}

	MetaFormattedValuer interface {
		FormattedValue(ctx *Context, meta *Meta, record interface{}) string
	}

	MetaorFormattedValuer interface {
		FormattedValue(ctx *core.Context, meta resource.Metaor, record interface{}) string
	}

	MetaSingleFormattedValuer interface {
		FormattedValue() string
	}
)

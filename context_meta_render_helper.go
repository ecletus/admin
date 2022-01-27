package admin

import (
	"context"
	"fmt"

	"github.com/ecletus/core"
	"github.com/moisespsena/template/text/template"
)

type MetaConfigHelper struct {
	AnyIndexPath string
}

func NewMetaConfigHelper(anyIndexPath string) *MetaConfigHelper {
	return &MetaConfigHelper{AnyIndexPath: anyIndexPath}
}

func (h *MetaConfigHelper) Setter(ctx *core.LocalContext) *MetaConfigHelperSetter {
	return &MetaConfigHelperSetter{h, ctx}
}

func (h *MetaConfigHelper) KeyOf(name string) string {
	return fmt.Sprintf("meta:%s:%s", h.AnyIndexPath, name)
}

func (h *MetaConfigHelper) GetValue(ctx context.Context, name string, defaul ...interface{}) (value interface{}) {
	if value = ctx.Value(h.KeyOf(name)); value == nil {
		for _, value = range defaul {
			return
		}
	}
	return
}

func (h *MetaConfigHelper) KeyOfTypeName() string {
	return fmt.Sprintf("meta:%s:type", h.AnyIndexPath)
}

func (h *MetaConfigHelper) GetTypeName(ctx *core.LocalContext) string {
	if v, ok := ctx.Get(h.KeyOfTypeName()); ok {
		return v.(string)
	}
	return ""
}

func (h *MetaConfigHelper) KeyOfSectionLayout() string {
	return fmt.Sprintf("meta:%s:section_layout", h.AnyIndexPath)
}

func (h *MetaConfigHelper) GetSectionLayout(ctx *core.LocalContext) string {
	if v, ok := ctx.Get(h.KeyOfSectionLayout()); ok {
		return v.(string)
	}
	return ""
}

func (h *MetaConfigHelper) KeyOfTemplate() string {
	return fmt.Sprintf("meta:%s:template", h.AnyIndexPath)
}

func (h *MetaConfigHelper) GetTemplate(ctx *core.LocalContext) string {
	if v, ok := ctx.Get(h.KeyOfTemplate()); ok {
		return v.(string)
	}
	return ""
}

func (h *MetaConfigHelper) KeyOfTemplateExecutor() string {
	return fmt.Sprintf("meta:%s:template_executor", h.AnyIndexPath)
}

func (h *MetaConfigHelper) GetTemplateExecutor(ctx *Context, kind string, m *Meta, fv *FormattedValue) (*template.Executor, error) {
	if v, ok := ctx.Get(h.KeyOfTemplateExecutor()); ok {
		switch t := v.(type) {
		case *template.Executor:
			return t, nil
		case func(ctx *Context, kind string, m *Meta, v *FormattedValue) (*template.Executor, error):
			return t(ctx, kind, m, fv)
		}
	}
	return nil, nil
}

type MetaConfigHelperSetter struct {
	Helper *MetaConfigHelper
	Ctx    *core.LocalContext
}

func (h *MetaConfigHelperSetter) TemplateExecutorF(tef func(ctx *Context, kind string, m *Meta, v *FormattedValue) (*template.Executor, error)) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOfTemplateExecutor(), tef)
	return h
}

func (h *MetaConfigHelperSetter) TemplateExecutor(te *template.Executor) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOfTemplateExecutor(), te)
	return h
}

func (h *MetaConfigHelperSetter) Template(templateName string) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOfTemplate(), templateName)
	return h
}

func (h *MetaConfigHelperSetter) Value(name string, value interface{}) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOf(name), value)
	return h
}

func (h *MetaConfigHelperSetter) TypeName(typeName string) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOfTypeName(), typeName)
	return h
}

func (h *MetaConfigHelperSetter) SectionLayout(typeName string) *MetaConfigHelperSetter {
	h.Ctx.SetValue(h.Helper.KeyOfSectionLayout(), typeName)
	return h
}

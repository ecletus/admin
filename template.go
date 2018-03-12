package admin

import (
	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
	"io"
)

func (context *Context) TemplateExecutor(t *template.Template, funcMaps... funcs.FuncMap) *template.Executor {
	return t.CreateExecutor().FuncsValues(context.FuncValues()).Funcs(funcMaps...)
}

func (context *Context) ExecuteTemplate(t *template.Template, out io.Writer, data interface{}, funcMaps... funcs.FuncMap) error {
	return context.TemplateExecutor(t).Funcs(funcMaps...).Execute(out, data)
}

func (context *Context) ExecuteTemplateError(t *template.Template, out io.Writer, data interface{}, funcMaps... funcs.FuncMap) error {
	return context.TemplateExecutor(t).WriteError().Funcs(funcMaps...).Execute(out, data)
}



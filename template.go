package admin

import (
	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
	"io"
)

func (this *Context) TemplateExecutor(t *template.Template, funcMaps... funcs.FuncMap) *template.Executor {
	return t.CreateExecutor().FuncsValues(this.FuncValues()).Funcs(funcMaps...)
}

func (this *Context) ExecuteTemplate(t *template.Template, out io.Writer, data interface{}, funcMaps... funcs.FuncMap) error {
	return this.TemplateExecutor(t).Funcs(funcMaps...).Execute(out, data)
}

func (this *Context) ExecuteTemplateError(t *template.Template, out io.Writer, data interface{}, funcMaps... funcs.FuncMap) error {
	return this.TemplateExecutor(t).WriteError().Funcs(funcMaps...).Execute(out, data)
}



package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena/template/html/template"
)

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

// T call i18n backend to translate
func (this *Admin) T(context *core.Context, key string, value interface{}, values ...interface{}) template.HTML {
	return template.HTML(this.Ts(context, key, value, values...))
}

// T call i18n backend to translate
func (this *Admin) Ts(context *core.Context, key string, value interface{}, values ...interface{}) string {
	if len(values) > 1 {
		panic("Values has many args.")
	}

	t := context.GetI18nContext().T(key).Default(value)

	if len(values) == 1 {
		t.Data(values[0])
	}

	return t.Get()
}

// TT call i18n backend to translate template
func (this *Admin) TT(context *core.Context, key string, data interface{}, defaul ...string) template.HTML {
	t := context.GetI18nContext().T(key).Data(data)
	if len(defaul) > 0 {
		t = t.Default(defaul[0])
	}
	return template.HTML(t.Get())
}

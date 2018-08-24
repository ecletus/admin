package admin

import (
	"github.com/moisespsena/template/html/template"
	"github.com/aghape/core"
)

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

// T call i18n backend to translate
func (admin *Admin) T(context *core.Context, key string, value string, values ...interface{}) template.HTML {
	if len(values) > 1 {
		panic("Values has many args.")
	}

	t := context.GetI18nContext().T(key).Default(value)

	if len(values) == 1 {
		t.Data(values[0])
	}

	return template.HTML(t.Get())
}

// TT call i18n backend to translate template
func (admin *Admin) TT(context *core.Context, key string, data interface{}, defaul ...string) template.HTML {
	t := context.GetI18nContext().T(key).Data(data)
	if len(defaul) > 0 {
		t = t.Default(defaul[0])
	}
	return template.HTML(t.Get())
}
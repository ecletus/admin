package admin

import (
	"bytes"
	"strings"

	"github.com/moisespsena/template/html/template"
)

func (this *Meta) formatTemplateString(ctx *Context, templateName, v string) string {
	if strings.Contains(v, "{{") {
		tmpl, err := template.New("meta::" + templateName + "{" + this.Name + "}").Parse(v)
		if err != nil {
			return "[[parse template failed: " + err.Error() + "]]"
		}
		var w bytes.Buffer
		if err = ctx.ExecuteTemplate(tmpl, &w, this); err != nil {
			return "[[execute template failed: " + err.Error() + "]]"
		}
		return w.String()
	}
	return v
}

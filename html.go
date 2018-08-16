package admin

import (
	"strings"

	"github.com/moisespsena/template/html/template"
)

var htmlWithIconsDefaultTemplate, _ = template.New("html_with_icons").Parse(`
{{$iconSize := .Value.IconSize}}
{{range $icon := .Value.Icons}}
<img height="{{$iconSize}}" width="{{$iconSize}}" src="{{.}}"/>
{{end}}
{{.Value.Body}}
`)

type HTMLWithIcons struct {
	Icons    []string
	Body     template.HTML
	Template string
	IconSize int
}

func (h *HTMLWithIcons) Icon(icon ...string) {
	h.Icons = append(h.Icons, icon...)
}

func (h HTMLWithIcons) Htmlify(context *Context) template.HTML {
	if h.IconSize == 0 {
		h.IconSize = 24
	}
	data := map[string]interface{}{
		"Context": context,
		"Value":   h,
	}
	if h.Template == "" {
		r, err := htmlWithIconsDefaultTemplate.ExecuteString(data)
		if err != nil {
			panic(err)
		}
		return template.HTML(strings.TrimSpace(r))
	}
	t, err := context.GetTemplate(h.Template)
	if err != nil {
		panic(err)
	}
	r, err := t.ExecuteString(data)
	if err != nil {
		panic(err)
	}
	return template.HTML(r)
}

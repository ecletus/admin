package admin

import (
	"bytes"

	"github.com/ecletus/core"
	"github.com/moisespsena/template/html/template"
)

// Execute execute template with layout
func (this *Context) Execute(name string, result interface{}) {
	if name == "" {
		name = this.Type.S()
	}

	if this.Action == "" {
		this.Action = name
	}

	var (
		executor *template.Executor
		err      error
		layout   = "layout"
	)

	if this.Anonymous() {
		layout = AnonymousDirName + "/" + layout
	}

	if executor, err = this.GetTemplate(layout); err != nil {
		panic(err)
	}

	this.Result = result
	this.TemplateName = name

	var buf bytes.Buffer

	//defer this.Write(buf.Bytes())

	defer this.templatesStack.Add(layout)()

	if err := executor.Execute(&buf, this); err != nil {
		this.Errors = core.Errors{}
		panic(err)
	}
	this.Write(buf.Bytes())
}

// RawValueOf return raw value of a meta for current resource
func (this *Context) RawValueOf(value interface{}, meta *Meta) interface{} {
	return this.valueOf(meta.GetValuer(), value, meta)
}

// FormattedValueOf return formatted value of a meta for current resource
func (this *Context) FormattedValueOf(value interface{}, meta *Meta) interface{} {
	result := this.valueOf(meta.GetFormattedValuer(), value, meta)
	if result == nil {
		return ""
	}
	return result
}

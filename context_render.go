package admin

import (
	"bytes"
	"fmt"
	"io"

	"github.com/moisespsena-go/tracederror"
	"github.com/moisespsena/template/html/template"
	"github.com/pkg/errors"

	"github.com/moisespsena-go/assetfs/assetfsapi"
)

// renderWithF render template based on data
func (this *Context) renderWithF(out io.Writer, name string, data interface{}) (err error) {
	pth := this.templatesStack.Abs(name)
	if pth == "" {
		return fmt.Errorf("bad template name %q", name)
	}
	defer this.templatesStack.Add(pth)()
	var executor *template.Executor
	if executor, err = this.GetTemplate(pth); err != nil {
		return errors.Wrapf(err, "get template %q", pth)
	}
	return errors.Wrapf(executor.Execute(out, data), "execute template %q", pth)
}
// renderWithF render template based on data
func (this *Context) renderExecutor(executor *template.Executor, out io.Writer, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := this.templatesStack.StringMessage("panic at")
			if et, ok := r.(tracederror.TracedError); ok {
				err = tracederror.Wrap(et, msg)
			} else if err2, ok := r.(error); ok {
				err = tracederror.New(errors.Wrap(err2, msg), et.Trace())
			} else {
				err = tracederror.New(errors.Wrap(fmt.Errorf("recoverd_error %T: %v", r, r), msg))
			}
		} else if err != nil && len(*this.templatesStack) > 0 {
			err = errors.Wrap(err, this.templatesStack.String())
		}
	}()
	return executor.Execute(out, data)
}

// renderWith render template based on data
func (this *Context) renderWith(name string, data interface{}) template.HTML {
	var w bytes.Buffer
	if err := this.renderWithF(&w, name, data); err != nil {
		w.Write([]byte("<pre>" + err.Error() + "</pre>"))
	}
	return template.HTML(w.String())
}

// renderWithInfoF render template based on FileInfo and data
func (this *Context) renderWithInfoF(out io.Writer, info assetfsapi.FileInfo, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = tracederror.TracedWrap(r, "panic of %q", info.RealPath())
		}
	}()
	var executor *template.Executor
	if executor, err = this.GetTemplateInfo(info); err != nil {
		return errors.Wrapf(err, "get template info %q", info.RealPath())
	}
	return errors.Wrapf(executor.Execute(out, data), "execute template %q", info.RealPath())
}

// RenderF render template based on context
func (this *Context) RenderF(out io.Writer, name string, results ...interface{}) error {
	clone := this.Clone()
	if len(results) > 0 {
		clone.Result = results[0]
	}
	return clone.renderWithF(out, name, clone)
}

// Render render template based on context
func (this *Context) RenderHtml(name string, results ...interface{}) template.HTML {
	var w bytes.Buffer
	if err := this.RenderF(&w, name, results...); err != nil {
		w.Write([]byte("<pre>" + err.Error() + "</pre>"))
	}
	return template.HTML(w.String())
}

// Include render template based on context
func (this *Context) Include(w io.Writer, name string, results ...interface{}) {
	if err := this.RenderF(w, name, results...); err != nil {
		w.Write([]byte("<pre>" + err.Error() + "</pre>"))
		panic(err)
	}
}

// Include render template based on context
func (this *Context) defaultYield(w io.Writer, results ...interface{}) {
	if err := this.RenderF(w, this.TemplateName, results...); err != nil {
		w.Write([]byte("<pre>" + err.Error() + "</pre>"))
		panic(err)
	}
}

// UseTheme append used themes into current context, will load those theme's stylesheet, javascripts in admin pages
func (this *Context) UseTheme(name string) {
	this.usedThemes = append(this.usedThemes, name)
}

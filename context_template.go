package admin

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	oscommon "github.com/moisespsena-go/os-common"
	"github.com/moisespsena/template/cache"
	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
	"github.com/pkg/errors"
)

var TemplateGlob = assetfs.NewGlobPattern("\f*.tmpl")

// renderText render text based on data
func (this *Context) renderText(text string, data interface{}) template.HTML {
	var (
		err    error
		tmpl   *template.Template
		result bytes.Buffer
	)

	if tmpl, err = template.New("").Parse(text); err == nil {
		if err = this.ExecuteTemplate(tmpl, &result, data); err == nil {
			return template.HTML(result.String())
		}
	}

	return template.HTML(err.Error())
}

func (this *Context) LoadTemplate(name string) (exc *template.Executor, err error) {
	if exc, err = this.LoadSiteTemplate(name); err == nil {
		return
	} else if oscommon.IsNotFound(err) {
		var asset assetfs.AssetInterface
		if asset, err = this.Asset(name + ".tmpl"); err != nil {
			return
		}
		tmpl, err := template.New(name).SetPath(asset.Path()).Parse(assetfs.MustDataS(asset))
		if err != nil {
			return nil, err
		}
		exc = tmpl.CreateExecutor()
		exc.Context = this
	}
	return
}

func (this *Context) LoadTemplateInfo(info assetfsapi.FileInfo) (*template.Executor, error) {
	if info.Size() > 30*2014 {
		return nil, fmt.Errorf("template %q is too long", info.RealPath())
	}
	data, err := assetfs.DataS(info)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New(info.Name()).SetPath(info.RealPath()).Parse(data)
	if err != nil {
		return nil, err
	}
	exc := tmpl.CreateExecutor()
	exc.Context = this
	return exc, nil
}

func (this *Context) GetTemplateOrDefault(name string, defaul *template.Executor, others ...string) (t *template.Executor, err error) {
	if t, err = cache.Cache.LoadOrStoreNames(name, this.LoadTemplate, others...); err != nil {
		if !oscommon.IsNotFound(err) {
			return
		}
		t = defaul
		err = nil
	}
	t.Context = this
	return t.FuncsValues(this.FuncValues()), nil
}

// renderWith render template based on data
func (this *Context) GetTemplate(name string, others ...string) (t *template.Executor, err error) {
	if t, err = cache.Cache.LoadOrStoreNames(name, this.LoadTemplate, others...); err != nil {
		return
	}
	if t == nil && err == nil {
		var msg string
		if len(others) > 0 {
			msg = "Templates with \"" + strings.Join(append([]string{name}, others...), "\", \"") + "\" does not exists."
		} else {
			msg = "Template \"" + name + "\" not exists."
		}
		return nil, errors.New(msg)
	}
	return t.FuncsValues(this.FuncValues()), nil
}

// GetTemplateInfo
func (this *Context) GetTemplateInfo(info assetfsapi.FileInfo, others ...assetfsapi.FileInfo) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreInfos(info, this.LoadTemplateInfo, others...)
	if err != nil {
		return nil, err
	}
	return t.FuncsValues(this.FuncValues()), nil
}

// FuncValues return funcs FuncValues
func (this *Context) FuncValues() funcs.FuncValues {
	if this.funcValues == nil {
		v, err := funcs.CreateValuesFunc(this.FuncMaps()...)
		if err != nil {
			panic(err)
		}
		this.funcValues = v
	}
	return this.funcValues
}

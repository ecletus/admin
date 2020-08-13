package admin

import (
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena/template/html/template"
)

func (this *Context) GetSiteTemplateInfo(name string) (exc *template.Executor, err error) {
	exc, err = this.LoadSiteTemplate(name)
	if err != nil {
		return nil, err
	}
	return exc.FuncsValues(this.FuncValues()), nil
}

func (this *Context) LoadSiteTemplate(name string) (*template.Executor, error) {
	if info, err := this.LoadSiteTemplateInfo(name + ".tmpl"); err != nil {
		return nil, err
	} else {
		return this.LoadTemplateInfo(info)
	}
}

func (this *Context) LoadSiteTemplateInfo(name string) (assetfsapi.FileInfo, error) {
	return this.SiteTemplateFS.AssetInfo(name)
}

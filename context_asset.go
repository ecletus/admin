package admin

import (
	"path/filepath"

	oscommon "github.com/moisespsena-go/os-common"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

type assetfsCache struct {
	assetfsapi.Interface
}

func (this *Context) templateFsS() []assetfs.Interface {
	return []assetfs.Interface{this.SiteTemplateFS, assetfsCache{this.Admin.TemplateFS}}
}
func (this *Context) staticFsS() []assetfs.Interface {
	return []assetfs.Interface{this.SiteStaticFS, assetfsCache{this.Admin.StaticFS}}
}

func (this *Context) Asset(layouts ...string) (asset assetfs.AssetInterface, err error) {
	return this.getAsset(this.templateFsS(), layouts...)
}

func (this *Context) StaticAsset(layouts ...string) (asset assetfs.AssetInterface, err error) {
	return this.getAsset(this.staticFsS(), layouts...)
}

func (this *Context) glob(pattern assetfs.GlobPatter, cb func(info assetfsapi.FileInfo), fs ...assetfs.Interface) {
	for _, fs := range fs {
		fs.NewGlob(pattern).Info(func(info assetfsapi.FileInfo) error {
			cb(info)
			return nil
		})
	}
}

func (this *Context) GlobTemplate(pattern assetfs.GlobPatter, cb func(info assetfsapi.FileInfo)) {
	this.glob(pattern, cb, this.SiteTemplateFS, this.Admin.TemplateFS)
}

func (this *Context) GlobStatic(pattern assetfs.GlobPatter, cb func(info assetfsapi.FileInfo)) {
	this.glob(pattern, cb, this.SiteStaticFS, this.Admin.StaticFS)
}

func (this *Context) findAsset(fs []assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, afs assetfs.Interface, err error) {
	var prefixes, themes []string

	if this.Request != nil {
		if theme := this.Request.URL.Query().Get("theme"); theme != "" {
			themes = append(themes, theme)
		}
	}

	if len(themes) == 0 && this.Resource != nil {
		for _, theme := range this.Resource.Config.Themes {
			themes = append(themes, theme.GetName())
		}
	}

	if resourcePath := this.resourcePath(); resourcePath != "" {
		for _, theme := range themes {
			prefixes = append(prefixes, filepath.Join("themes", theme, resourcePath))
		}
		prefixes = append(prefixes, resourcePath)
	}

	for _, theme := range themes {
		prefixes = append(prefixes, filepath.Join("themes", theme))
	}

	return this.findAssetThemes(layouts, prefixes, fs...)
}

func (this *Context) findAssetThemes(layouts, prefixes []string, fs ...assetfs.Interface) (asset assetfs.AssetInterface, afs assetfs.Interface, err error) {
	for _, layout := range layouts {
		for _, prefix := range prefixes {
			for _, afs = range fs {
				if asset, err = afs.Asset(filepath.Join(prefix, layout)); err == nil {
					return
				}
			}
		}
		for _, afs = range fs {
			if asset, err = afs.Asset(layout); err == nil {
				return
			}
		}
	}

	return nil, nil, oscommon.ErrNotFound(PathStack(layouts).String(), "theme not found")
}

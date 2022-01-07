// +build !dev

package admin

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

func (this *Context) javaScriptTag(names ...string) template.HTML {
	var results []string
	prefix := this.JoinStaticURL("javascripts")
	cache := &this.Admin.Cache
	for _, name := range names {
		pth := cache.LoadOrFactory(name+".js", func() interface{} {
			for _, ext := range []string{".min", ""} {
				var file = filepath.Join("javascripts", name+ext+".js")
				if _, err := this.StaticAsset(file); err == nil {
					return prefix + "/" + name + ext + ".js"
				}
			}
			return ""
		})
		if pth != "" {
			results = append(results, fmt.Sprintf(`<script src="%s"></script>`, pth))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

func (this *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := this.JoinStaticURL("stylesheets")
	cache := &this.Admin.Cache
	for _, name := range names {
		var media string
		if pos := strings.IndexByte(name, '@'); pos > 0 {
			media, name = name[pos+1:], name[:pos]
		}
		pth := cache.LoadOrFactory(name+".css", func() interface{} {
			for _, ext := range []string{".min", ""} {
				var file = filepath.Join("stylesheets", name+ext+".css")
				if _, err := this.StaticAsset(file); err == nil {
					return prefix + "/" + name + ext + ".css"
				}
			}
			return ""
		})
		if pth != "" {
			if media != "" {
				media = `media="` + media + `" `
			}
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s" %s/>`, pth, media))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

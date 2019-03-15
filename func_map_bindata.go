// +build assetfs_bindata

package admin

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

func (context *Context) javaScriptTag(names ...string) template.HTML {
	var results []string
	prefix := context.GenStaticURL("javascripts")
	cache := &context.Admin.Cache
	for _, name := range names {
		pth := cache.LoadOrFactory(name+".js", func() interface{} {
			for _, ext := range []string{".min", ""} {
				var file = filepath.Join("javascripts", name+ext+".js")
				if _, err := context.StaticAsset(file); err == nil {
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

func (context *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := context.GenStaticURL("stylesheets")
	cache := &context.Admin.Cache
	for _, name := range names {
		pth := cache.LoadOrFactory(name+".css", func() interface{} {
			for _, ext := range []string{".min", ""} {
				var file = filepath.Join("stylesheets", name+ext+".css")
				if _, err := context.StaticAsset(file); err == nil {
					return prefix + "/" + name + ext + ".css"
				}
			}
			return ""
		})
		if pth != "" {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, pth))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

// +build !assetfs_bindata

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
	for _, name := range names {
		if _, err := context.StaticAsset(filepath.Join("javascripts", name+".js")); err == nil {
			results = append(results, fmt.Sprintf(`<script src="%s"></script>`, prefix+"/"+name+".js"))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := context.GenStaticURL("stylesheets")
	for _, name := range names {
		if _, err := context.StaticAsset(filepath.Join("stylesheets", name+".css")); err == nil {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s/%s.css">`, prefix, name))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

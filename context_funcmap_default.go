// +build dev

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
	for _, name := range names {
		if _, err := this.StaticAsset(filepath.Join("javascripts", name+".js")); err == nil {
			results = append(results, fmt.Sprintf(`<script src="%s"></script>`, prefix+"/"+name+".js"))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

func (this *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := this.JoinStaticURL("stylesheets")
	for _, name := range names {
		var media string
		if pos := strings.IndexByte(name, '@'); pos > 0 {
			media, name = `media="`+name[pos+1:]+`" `, name[:pos]
		}
		if _, err := this.StaticAsset(filepath.Join("stylesheets", name+".css")); err == nil {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s/%s.css" %s/>`, prefix, name, media))
		}
	}
	return template.HTML(strings.Join(results, ""))
}

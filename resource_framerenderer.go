package admin

import (
	"github.com/moisespsena-go/maps"
	"github.com/moisespsena/template/text/template"

	"github.com/ecletus/core"
)

type PathTemplater interface {
	core.Configor
	GetTemplatePaths() []string
	GetData() maps.Interface
}

type frameRendererKey struct{ name string }

func SetFrameRenderer(ptr PathTemplater, name string, renderer FrameRenderer) {
	ptr.GetData().Set(frameRendererKey{name}, renderer)
}

func GetFrameRenderer(ptr PathTemplater, name string) (renderer FrameRenderer) {
	if v, ok := ptr.GetData().Get(frameRendererKey{name}); ok && v != nil {
		return v.(FrameRenderer)
	}
	return
}

func GetFrameRendererTemplateName(ptr PathTemplater, ctx *Context, name string) (templateNames []string) {
	for _, pth := range ptr.GetTemplatePaths() {
		templateNames = append(templateNames, pth+"/frames/"+name)
		if f := GetOptFrameRendererTemplateNames(ptr, name); f != nil {
			templateNames = f(ctx, templateNames)
		}
	}
	return
}

type FrameRenderer interface {
	Render(ctx *Context, state *template.State) error
}

type FrameRendererFunc = func(ctx *Context, state *template.State) error
type funcFrameRenderer FrameRendererFunc

func (this funcFrameRenderer) Render(ctx *Context, state *template.State) error {
	return this(ctx, state)
}

func NewFrameRenderer(f FrameRendererFunc) FrameRenderer {
	return funcFrameRenderer(f)
}

func OptFrameRendererTemplateNames(frameName string, f func(ctx *Context, names []string) (templateNames []string)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("frame_renderer:"+frameName+":template", f)
	})
}
func GetOptFrameRendererTemplateNames(configor core.Configor, frameName string) (f func(ctx *Context, names []string) (templateNames []string)) {
	if value, ok := configor.ConfigGet("frame_renderer:" + frameName + ":template"); ok {
		f = value.(func(ctx *Context, names []string) []string)
	}
	return
}

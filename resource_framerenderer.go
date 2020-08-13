package admin

import (
	"github.com/moisespsena/template/text/template"

	"github.com/ecletus/core"
)

type frameRendererKey struct{ name string }

func (this *Resource) SetFrameRenderer(name string, renderer FrameRenderer) {
	this.Data.Set(frameRendererKey{name}, renderer)
}

func (this *Resource) GetFrameRenderer(name string) (renderer FrameRenderer) {
	if v := this.Data.MustGetInterface(frameRendererKey{name}); v != nil {
		return v.(FrameRenderer)
	}
	return
}

func (this *Resource) GetFrameRendererTemplateName(ctx *Context, name string) (templateNames []string) {
	templateNames = append(templateNames, this.TemplatePath + "/frames/" + name)
	if f := GetOptFrameRendererTemplateNames(this, name); f != nil {
		templateNames = f(ctx, templateNames)
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
		f = value.(func(ctx *Context, names []string) ([]string))
	}
	return
}

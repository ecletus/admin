package admin

import (
	"time"

	"github.com/ecletus/core"
)

func GetOptLocation(configor core.Configor) func(ctx *Context) *time.Location {
	if configor != nil {
		if v, ok := configor.ConfigGet("form:update.button.label"); ok {
			return v.(func(ctx *Context) *time.Location)
		}
	}
	return nil
}

func OptGetLocation(configor core.Configor, ctx *Context) *time.Location {
	if f := GetOptLocation(configor); f != nil {
		return f(ctx)
	}
	return ctx.TimeLocation()
}

func OptLocationF(f func(ctx *Context) *time.Location) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("location", f)
	})
}

func OptLocation(loc *time.Location) core.Option {
	return OptLocationF(func(ctx *Context) *time.Location {
		return loc
	})
}

func OptActionTitleFormatter(f func(arg *ActionArgument, title string) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("action:title_formatter", f)
	})
}

func GetOptActionTitleFormatter(configor core.Configor) func(arg *ActionArgument, title string) string {
	if v, ok := configor.ConfigGet("action:title_formatter"); ok {
		return v.(func(arg *ActionArgument, title string) string)
	}
	return nil
}

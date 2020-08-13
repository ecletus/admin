package admin

import "github.com/ecletus/core"

func OptShowEditDisabledF(f func(ctx *Context) bool) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:show.edit_disabled", f)
	})
}

func OptShowEditDisabled() core.Option {
	return OptShowEditDisabledF(func(ctx *Context) bool {
		return true
	})
}

func GetOptShowEditDisabled(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("form:show.edit_disabled"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}
func OptShowActionsDisabledF(f func(ctx *Context) bool) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:show.actions_disabled", f)
	})
}

func OptShowActionsDisabled() core.Option {
	return OptShowEditDisabledF(func(ctx *Context) bool {
		return true
	})
}

func GetOptShowActionsDisabled(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("form:show.actions_disabled"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}

type ShowConfig struct {
	EditDisabled    bool
	ActionsDisabled bool
}

func (this *ShowConfig) Load(configor core.Configor, ctx *Context) {
	if f := GetOptShowEditDisabled(configor); f != nil {
		this.EditDisabled = f(ctx)
	}
	if f := GetOptShowActionsDisabled(configor); f != nil {
		this.ActionsDisabled = f(ctx)
	}
}

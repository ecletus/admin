package admin

import "github.com/ecletus/core"

func OptFormUpdateButtonLabel(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.button.label", f)
	})
}
func OptFormUpdateButtonLabelString(label string) core.Option {
	return OptFormUpdateButtonLabel(func(ctx *Context) string {
		return label
	})
}
func GetOptUpdateButtonLabel(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:update.button.label"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}
func OptFormUpdateButtonCancelUrl(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.button.cancel.url", f)
	})
}

func GetOptUpdateButtonCancelUrl(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:update.button.cancel.url"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormUpdateRedirectTo(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.redirect_to", f)
	})
}

func GetOptUpdateRedirectTo(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:update.redirect_to"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormUpdateAction(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.form_action", f)
	})
}

func GetOptUpdateAction(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:update.form_action"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormUpdateContinueEditingDisabledF(f func(ctx *Context) bool) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.continue_editing_disabled", f)
	})
}

func OptFormUpdateContinueEditingDisabled() core.Option {
	return OptFormUpdateContinueEditingDisabledF(func(ctx *Context) bool {
		return true
	})
}

func GetOptUpdateContinueEditingDisabled(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("form:update.continue_editing_disabled"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}

func OptFormUpdateSuccess(f func(ctx *Context, old, record interface{}, message *string)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.success_cb", f)
	})
}

func GetOptFormUpdateSuccess(configor core.Configor) func(ctx *Context, old, record interface{}, message *string) {
	if v, ok := configor.ConfigGet("form:update.success_cb"); ok {
		return v.(func(ctx *Context, old, record interface{}, message *string))
	}
	return nil
}

func OptFormUpdatePrepare(f func(ctx *Context)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.prepare", f)
	})
}

func GetOptFormUpdatePrepare(configor core.Configor) func(ctx *Context) {
	if v, ok := configor.ConfigGet("form:update.prepare"); ok {
		return v.(func(ctx *Context))
	}
	return nil
}

func OptFormUpdateError(f func(ctx *Context, err error)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:update.error_cb", f)
	})
}

func GetOptFormUpdateError(configor core.Configor) func(ctx *Context, err error) {
	if v, ok := configor.ConfigGet("form:update.error_cb"); ok {
		return v.(func(ctx *Context, err error))
	}
	return nil
}

type UpdateConfig struct {
	ButtonLabel             string
	RedirectTo              string
	CancelUrl               string
	FormAction              string
	ContinueEditingDisabled bool
	SuccessCallback         func(ctx *Context, old, record interface{}, message *string)
	ErrorCallback           func(ctx *Context, err error)
	Prepare                 func(ctx *Context)
}

func (this *UpdateConfig) Load(configor core.Configor, ctx *Context) {
	if f := GetOptUpdateRedirectTo(configor); f != nil {
		this.RedirectTo = f(ctx)
	}
	if f := GetOptUpdateButtonLabel(configor); f != nil {
		this.ButtonLabel = f(ctx)
	}
	if f := GetOptUpdateButtonCancelUrl(configor); f != nil {
		this.CancelUrl = f(ctx)
	}
	if f := GetOptUpdateAction(configor); f != nil {
		this.FormAction = f(ctx)
	}
	if f := GetOptUpdateContinueEditingDisabled(configor); f != nil {
		this.ContinueEditingDisabled = f(ctx)
	}
	if f := GetOptFormUpdateSuccess(configor); f != nil {
		this.SuccessCallback = f
	}
	if f := GetOptFormUpdateError(configor); f != nil {
		this.ErrorCallback = f
	}
	if f := GetOptFormUpdatePrepare(configor); f != nil {
		this.Prepare = f
	}
}

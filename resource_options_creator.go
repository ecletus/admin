package admin

import "github.com/ecletus/core"

func OptFormCreateButtonLabel(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.button.label", f)
	})
}
func OptFormCreateButtonLabelString(label string) core.Option {
	return OptFormCreateButtonLabel(func(ctx *Context) string {
		return label
	})
}

func GetOptCreateButtonLabel(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:create.button.label"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormCreateRedirectTo(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.redirect_to", f)
	})
}

func GetOptCreateRedirectTo(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:create.redirect_to"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormCreateButtonCancelUrl(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.button.cancel.url", f)
	})
}

func GetOptCreateButtonCancelUrl(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:create.button.cancel.uri"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormCreateAction(f func(ctx *Context) string) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.form_action", f)
	})
}

func GetOptCreateAction(configor core.Configor) func(ctx *Context) string {
	if v, ok := configor.ConfigGet("form:create.form_action"); ok {
		return v.(func(ctx *Context) string)
	}
	return nil
}

func OptFormCreateContinueEditingDisabledF(f func(ctx *Context) bool) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.continue_editing_disabled", f)
	})
}
func OptFormCreateContinueEditingDisabled() core.Option {
	return OptFormCreateContinueEditingDisabledF(func(ctx *Context) bool {
		return true
	})
}

func GetOptCreateContinueEditingDisabled(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("form:create.continue_editing_disabled"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}

func OptFormCreateSuccess(f func(ctx *Context, message string)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.success_cb", f)
	})
}

func GetOptFormCreateSuccess(configor core.Configor) func(ctx *Context, message string) {
	if v, ok := configor.ConfigGet("form:create.success_cb"); ok {
		return v.(func(ctx *Context, message string))
	}
	return nil
}

func OptFormCreateError(f func(ctx *Context, err error)) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:create.error_cb", f)
	})
}

func GetOptFormCreateError(configor core.Configor) func(ctx *Context, err error) {
	if v, ok := configor.ConfigGet("form:create.error_cb"); ok {
		return v.(func(ctx *Context, err error))
	}
	return nil
}

type CreateConfig struct {
	ButtonLabel             string
	RedirectTo              string
	CancelUrl               string
	FormAction              string
	ContinueEditingDisabled bool
	SuccessCallback         func(ctx *Context, message string)
	ErrorCallback           func(ctx *Context, err error)
}

func (this *CreateConfig) Load(configor core.Configor, ctx *Context) {
	if f := GetOptCreateRedirectTo(configor); f != nil {
		this.RedirectTo = f(ctx)
	}
	if f := GetOptCreateButtonLabel(configor); f != nil {
		this.ButtonLabel = f(ctx)
	}
	if f := GetOptCreateButtonCancelUrl(configor); f != nil {
		this.CancelUrl = f(ctx)
	}
	if f := GetOptCreateAction(configor); f != nil {
		this.FormAction = f(ctx)
	}
	if f := GetOptCreateContinueEditingDisabled(configor); f != nil {
		this.ContinueEditingDisabled = f(ctx)
	}
	if f := GetOptFormCreateSuccess(configor); f != nil {
		this.SuccessCallback = f
	}
	if f := GetOptFormCreateError(configor); f != nil {
		this.ErrorCallback = f
	}
}

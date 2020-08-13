package admin

import "github.com/ecletus/core"

func (this *Context) Options(opt ...core.Option) *Context {
	for _, opt := range opt {
		opt.Apply(this)
	}
	return this
}

func (this *Context) TypeConfig() interface{} {
	return GetOptContextTypeConfig(this)
}

func OptContextActionsDisabledF(f func(ctx *Context) bool) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("context:actions_disabled", f)
	})
}

func OptContextActionsDisabled() core.Option {
	return OptContextActionsDisabledF(func(ctx *Context) bool {
		return true
	})
}

func GetOptContextActionsDisabled(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("context:actions_disabled"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}

func OptContextTypeConfig(f func(ctx *Context) interface{}) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("context:type_config", f)
	})
}

func GetOptContextTypeConfig(configor core.Configor) func(ctx *Context) bool {
	if v, ok := configor.ConfigGet("context:type_config"); ok {
		return v.(func(ctx *Context) bool)
	}
	return nil
}

func OptContextRecordLoaded(f func(ctx *Context, record interface{})) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("context:record_loaded", append(GetOptContextRecordLoaded(configor), f))
	})
}

func GetOptContextRecordLoaded(configor ...core.Configor) (result []func(ctx *Context, record interface{})) {
	for _, configor := range configor {
		if v, ok := configor.ConfigGet("context:record_loaded"); ok {
			result = append(result, v.([]func(ctx *Context, record interface{}))...)
		}
	}
	return
}

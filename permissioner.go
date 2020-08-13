package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/roles"
)

type Permissioner interface {
	HasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm)
}

type Permissioners []Permissioner

func (this Permissioners) HasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	for _, permissioner := range this {
		if perm = permissioner.HasPermission(mode, ctx); perm != roles.UNDEF {
			return
		}
	}
	return
}

type PermissionerFuncType = func(mode roles.PermissionMode, ctx *Context) (perm roles.Perm)

type PermissionerFunc PermissionerFuncType

func (this PermissionerFunc) HasPermission(mode roles.PermissionMode, ctx *Context) roles.Perm {
	return this(mode, ctx)
}

func NewPermissioner(f PermissionerFuncType, fn ...PermissionerFuncType) Permissioner {
	if len(fn) == 0 {
		return PermissionerFunc(f)
	}
	fn = append([]PermissionerFuncType{f}, fn...)
	return PermissionerFunc(func(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
		for _, f := range fn {
			if perm = f(mode, ctx); perm != roles.UNDEF {
				return
			}
		}
		return
	})
}

func HasPermission(mode roles.PermissionMode, ctx *Context, permissioner ...Permissioner) (perm roles.Perm) {
	return Permissioners(permissioner).HasPermission(mode, ctx)
}

func PermissionerOf(p core.Permissioner, pn ...core.Permissioner) Permissioner {
	permissioners := core.Permissioners(append([]core.Permissioner{p}, pn...)...)
	return NewPermissioner(func(mode roles.PermissionMode, ctx *Context) roles.Perm {
		return permissioners.HasPermission(mode, ctx.Context)
	})
}

func ToCorePermissioner(p Permissioner) core.Permissioner {
	return core.NewPermissioner(func(mode roles.PermissionMode, ctx *core.Context) roles.Perm {
		c := ContextFromCoreContext(ctx)
		return p.HasPermission(mode, c)
	})
}

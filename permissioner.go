package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/roles"
)

type Permissioner interface {
	AdminHasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm)
}

type permissioners []Permissioner

func (this permissioners) AdminHasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	for _, p := range this {
		if perm = p.AdminHasPermission(mode, ctx); perm != roles.UNDEF {
			return
		}
	}
	return
}

func Permissioners(p ...Permissioner) Permissioner {
	var result permissioners
	for _, p := range p {
		if p == nil {
			continue
		}
		switch t := p.(type) {
		case permissioners:
			result = append(result, t...)
		default:
			result = append(result, p)
		}
	}
	if len(result) == 1 {
		return result[0]
	}
	return result
}

type PermissionerFuncType = func(mode roles.PermissionMode, ctx *Context) (perm roles.Perm)

type PermissionerFunc PermissionerFuncType

func (this PermissionerFunc) AdminHasPermission(mode roles.PermissionMode, ctx *Context) roles.Perm {
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
	return permissioners(permissioner).AdminHasPermission(mode, ctx)
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
		return p.AdminHasPermission(mode, c)
	})
}

type RecordPermissioner interface {
	AdminHasRecordPermission(mode roles.PermissionMode, ctx *Context, record interface{}) (perm roles.Perm)
}

type RecordPermissionerFunc func(mode roles.PermissionMode, ctx *Context, record interface{}) roles.Perm

func (f RecordPermissionerFunc) AdminHasRecordPermission(mode roles.PermissionMode, ctx *Context, record interface{}) roles.Perm {
	return f(mode, ctx, record)
}

func NewRecordPermissioner(f func(mode roles.PermissionMode, ctx *Context, record interface{}) (perm roles.Perm)) RecordPermissioner {
	return RecordPermissionerFunc(f)
}

type allowedPermissioners []Permissioner

func (this allowedPermissioners) AdminHasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	for _, p := range this {
		if !p.AdminHasPermission(mode, ctx).Allow() {
			return roles.DENY
		}
	}
	return roles.ALLOW
}

func AllowedPermissioners(p ...Permissioner) Permissioner {
	permr := Permissioners(p...)
	if items, ok := permr.(permissioners); ok {
		return allowedPermissioners(items)
	}
	return permr
}

func RolePermissioner(permissioner roles.Permissioner) Permissioner {
	return NewPermissioner(func(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
		return permissioner.HasPermission(ctx, mode, ctx.Roles.Interfaces()...)
	})
}

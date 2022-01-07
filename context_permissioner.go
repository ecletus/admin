package admin

import (
	"reflect"

	"github.com/ecletus/roles"
)

func (this *Context) DefaultDenyMode() bool {
	if this.Roles.Has(roles.Anyone) {
		return false
	}
	return this.Admin.DefaultDenyMode
}

func (this *Context) HasPermission(permissioner Permissioner, mode roles.PermissionMode, modeN ...roles.PermissionMode) (ok bool) {
	if res, ok := permissioner.(*Resource); ok {
		if ok := res.ControllerBuilder.HasModes(mode, modeN...); ok != nil && !*ok {
			return false
		}
		if mode == roles.Create {
			if res.Config.CreationAllowed != nil && !res.Config.CreationAllowed(this) {
				return false
			}
		}
	}
	if rp, ok := permissioner.(RecordPermissioner); ok && this.Result != nil && indirectType(reflect.TypeOf(this.Result)).Kind() == reflect.Struct {
		if perm := rp.AdminHasRecordPermission(mode, this, this.Result); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	if this.IsSuperUser() {
		return true
	}
	if perm := permissioner.AdminHasPermission(mode, this); perm != roles.UNDEF {
		return perm.Allow()
	}
	for _, mode := range modeN {
		if perm := permissioner.AdminHasPermission(mode, this); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	if _, ok := permissioner.(*Meta); ok {
		return true
	}
	ok = !this.DefaultDenyMode()
	return
}

func (this *Context) HasRecordPermission(resource *Resource, record interface{}, mode roles.PermissionMode) (ok bool) {
	if !this.HasPermission(resource, mode) {
		return false
	}
	if perm := resource.AdminHasRecordPermission(mode, this, record); perm == roles.UNDEF {
		return true
	} else {
		return perm.Allow()
	}
}
func (this *Context) HasAnyPermission(permissioner Permissioner, mode ...roles.PermissionMode) (ok bool) {
	return this.HasAnyPermissionDefault(permissioner, !this.DefaultDenyMode(), mode...)
}

func (this *Context) HasAnyPermissionDefault(permissioner Permissioner, defaul bool, mode ...roles.PermissionMode) (ok bool) {
	if res, ok := permissioner.(*Resource); ok {
		if ok := res.ControllerBuilder.HasModes(mode[0], mode[1:]...); ok != nil && !*ok {
			return false
		}
	}
	if this.IsSuperUser() {
		return true
	}
	for _, mode := range mode {
		if perm := permissioner.AdminHasPermission(mode, this); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	return defaul
}

func (this *Context) HasRolePermission(permissioner roles.Permissioner, mode roles.PermissionMode, modeN ...roles.PermissionMode) (ok bool) {
	if this.IsSuperUser() {
		return true
	}
	if perm := permissioner.HasPermission(this, mode); perm != roles.UNDEF {
		return perm.Allow()
	}
	for _, mode := range modeN {
		if perm := permissioner.HasPermission(this, mode); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	return !this.DefaultDenyMode()
}

func (this *Context) hasCreatePermission(permissioner Permissioner) bool {
	return this.HasPermission(permissioner, roles.Create)
}

func (this *Context) hasReadPermission(permissioner Permissioner) bool {
	return this.HasPermission(permissioner, roles.Read)
}

func (this *Context) hasUpdatePermission(permissioner Permissioner) bool {
	return this.HasPermission(permissioner, roles.Update)
}

func (this *Context) hasDeletePermission(permissioner Permissioner) bool {
	return this.HasPermission(permissioner, roles.Delete)
}

func (this *Context) readPermissionFilter(permissioners []Permissioner, result ...interface{}) (filtered []interface{}) {
	if len(result) > 0 {
		defer this.WithResult(result[0])()
	}
	for _, permissioner := range permissioners {
		if ok := this.HasPermission(permissioner, roles.Read); ok {
			filtered = append(filtered, permissioner)
		}
	}
	return
}

// AllowedActions return allowed actions based on context
func (this *Context) AllowedActions(actions []*Action, mode string, records ...interface{}) []*Action {
	var allowedActions []*Action
	for _, action := range actions {
		if len(action.Modes) > 0 {
			var ok bool
			for _, m := range action.Modes {
				if m == mode {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}

		if action.IsAllowed(this, records...) {
			allowedActions = append(allowedActions, action)
		}
	}
	return Actions(allowedActions).Sort()
}

type ContextPermissioner interface {
	AdminHasContextPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm)
}

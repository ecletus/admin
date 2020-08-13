package admin

import (
	"github.com/ecletus/roles"

	"github.com/ecletus/core"
)

func (this *Context) DefaultDenyMode() bool {
	if this.Roles.Has(roles.Anyone) {
		return false
	}
	return this.Admin.DefaultDenyMode
}

func (this *Context) HasPermission(permissioner core.Permissioner, mode roles.PermissionMode, modeN ...roles.PermissionMode) (ok bool) {
	if res, ok := permissioner.(*Resource); ok {
		if ok := res.ControllerBuilder.HasModes(mode, modeN...); ok != nil && !*ok {
			return false
		}
	}
	if rp, ok := permissioner.(core.RecordPermissioner); ok {
		if perm := rp.HasRecordPermission(mode, this.Context, this.Result); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	if this.IsSuperUser() {
		return true
	}
	if perm := permissioner.HasPermission(mode, this.Context); perm != roles.UNDEF {
		return perm.Allow()
	}
	for _, mode := range modeN {
		if perm := permissioner.HasPermission(mode, this.Context); perm != roles.UNDEF {
			return perm.Allow()
		}
	}
	ok = !this.DefaultDenyMode()
	return
}

func (this *Context) HasAnyPermission(permissioner core.Permissioner, mode ...roles.PermissionMode) (ok bool) {
	if res, ok := permissioner.(*Resource); ok {
		if ok := res.ControllerBuilder.HasModes(mode[0], mode[1:]...); ok != nil && !*ok {
			return false
		}
	}
	if this.IsSuperUser() {
		return true
	}
	for _, mode := range mode {
		if perm := permissioner.HasPermission(mode, this.Context); perm.Allow() {
			return true
		}
	}
	return !this.DefaultDenyMode()
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

func (this *Context) hasCreatePermission(permissioner core.Permissioner) bool {
	return this.HasPermission(permissioner, roles.Create)
}

func (this *Context) hasReadPermission(permissioner core.Permissioner) bool {
	return this.HasPermission(permissioner, roles.Read)
}

func (this *Context) hasUpdatePermission(permissioner core.Permissioner) bool {
	return this.HasPermission(permissioner, roles.Update)
}

func (this *Context) hasDeletePermission(permissioner core.Permissioner) bool {
	return this.HasPermission(permissioner, roles.Delete)
}

func (this *Context) readPermissionFilter(permissioners []core.Permissioner, result ...interface{}) (filtered []interface{}) {
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

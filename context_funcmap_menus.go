package admin

import (
	"sort"
	"strings"

	"github.com/ecletus/roles"
	"github.com/moisespsena/template/html/template"
)

type menu struct {
	*Menu
	Label    template.HTML
	Active   bool
	SubMenus []*menu
	URL      string
}

func (this menu) sortChildren() {
	// sort now because translated labels
	sort.Slice(this.SubMenus, func(i, j int) bool {
		var mi, mj = this.SubMenus[i], this.SubMenus[j]
		if mi.Priority == 0 {
			if mj.Priority == 0 {
				// 0 and 0
				return mi.Label < mj.Label
			} else if mj.Priority > 0 {
				// 0 and 2
				return false
			}
			// 0 and -1
			return true
		}
		if mi.Priority > 0 {
			if mj.Priority > 0 {
				// 2 and 3 or 3 and 2
				return mi.Priority < mj.Priority
			}
			// 2 and 0 or 2 and -1
			return false
		}
		// -10 and -1
		// -10 and -12
		if mj.Priority < 0 {
			return mi.Priority > mj.Priority
		}
		// -10 and 0
		return false
	})
}

func (this *Context) getMenus() (menus []*menu) {
	var (
		globalMenu        = &menu{}
		mostMatchedMenu   *menu
		mostMatchedLength int
		addMenu           func(*menu, []*Menu)
		path              = this.RequestPath()
	)

	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			if m.Enabled != nil {
				if !m.Enabled(m, this) {
					continue
				}
			}
			if this.HasPermission(m, roles.Read) {
				var menu = &menu{Menu: m}

				if strings.HasPrefix(path, m.URI) && len(m.URI) > mostMatchedLength {
					mostMatchedMenu = menu
					mostMatchedLength = len(m.URI)
				}

				menu.URL = m.URL(this)
				menu.Label = this.Tt(m)

				addMenu(menu, menu.GetSubMenus())
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}

		parent.sortChildren()
	}

	addMenu(globalMenu, this.Admin.GetMenus())

	if this.Action != "search_center" && mostMatchedMenu != nil {
		mostMatchedMenu.Active = true
	}

	return globalMenu.SubMenus
}

func (this *Context) getResourceItemMenus() (menus []*menu) {
	var (
		globalMenu = &menu{}
		addMenu    func(*menu, []*Menu)
	)

	var parents []interface{}
	for _, parentID := range this.ParentResourceID {
		parents = append(parents, parentID)
	}

	if this.ResourceID != nil {
		parents = append(parents, this.ResourceID)
	}

	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			if m.Enabled != nil {
				if !m.Enabled(m, this) {
					continue
				}
			}
			if this.HasPermission(m, roles.Read) {
				var menu = &menu{Menu: m}
				menu.URL = m.URL(this, parents...)
				menu.Label = this.Tt(m)
				addMenu(menu, menu.GetSubMenus())
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}

		parent.sortChildren()
	}

	if this.Resource != nil && (this.Resource.Config.Singleton || this.ResourceID != nil) {
		addMenu(globalMenu, this.Resource.GetItemMenus())
	}

	return globalMenu.SubMenus
}

func (this *Context) getResourceMenus() (menus []*menu) {
	var (
		globalMenu = &menu{}
		addMenu    func(*menu, []*Menu)
	)

	var parents []interface{}
	for _, parentID := range this.ParentResourceID {
		parents = append(parents, parentID)
	}

	if this.ResourceID != nil {
		parents = append(parents, this.ResourceID)
	}

	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			if m.Enabled != nil {
				if !m.Enabled(m, this) {
					continue
				}
			}
			if this.HasPermission(m, roles.Read) {
				var menu = &menu{Menu: m}
				menu.URL = m.URL(this, parents...)
				menu.Label = this.Tt(m)
				addMenu(menu, menu.GetSubMenus())
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}

		parent.sortChildren()
	}

	if this.Resource != nil {
		addMenu(globalMenu, this.Resource.GetMenus())
	}

	return globalMenu.SubMenus
}

func (this *Context) getResourceMenuActions() interface{} {
	if this.Resource != nil && (this.Resource.IsSingleton() || (this.ResourceID != nil && !this.IsResultSlice() && !this.Resource.GetKey(this.Result).IsZero())) {
		actions := Actions(this.AllowedActions(this.Resource.Actions, "menu_item", this.Result)).Sort()

		return &struct {
			Context  *Context
			Actions  []*Action
			Resource *Resource
			Result   interface{}
		}{this, actions, this.Resource, this.Result}
	}
	return nil
}

type scope struct {
	*Scope
	Active bool
}

type scopeMenu struct {
	Group  string
	Scopes []scope
}

// GetScopes get scopes from current context
func (this *Context) GetScopes() (menus []*scopeMenu) {
	if this.Resource == nil {
		return
	}

	scopes := this.Request.URL.Query()["scopes[]"]

OUT:
	for _, s := range this.Scheme.scopes {
		if s.Visible != nil && !s.Visible(this) {
			continue
		}

		menu := scope{Scope: s}

		for _, s := range scopes {
			if s == menu.Name {
				menu.Active = true
			}
		}

		if !menu.Default {
			if menu.Group != "" {
				for _, m := range menus {
					if m.Group == menu.Group {
						m.Scopes = append(m.Scopes, menu)
						continue OUT
					}
				}
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			} else {
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			}
		}
	}
	return menus
}

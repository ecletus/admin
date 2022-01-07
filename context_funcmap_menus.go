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
		addMenu           func(*menu, []*Menu) int
		path              = this.RequestPath()
	)

	addMenu = func(parent *menu, menus []*Menu) (i int) {
		for _, m := range menus {
			if m.EnabledFunc != nil {
				if !m.EnabledFunc(m, this) {
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

				if addMenu(menu, menu.GetSubMenus()) == 0 && menu.Dir {
					continue
				}
				i++
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}

		parent.sortChildren()
		return i
	}

	addMenu(globalMenu, this.Admin.GetMenus())

	if this.Action != "search_center" && mostMatchedMenu != nil {
		mostMatchedMenu.Active = true
	}

	return globalMenu.SubMenus
}

func (this *Context) getResourceItemMenus(item interface{}) (menus []*menu) {
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
			if this.HasPermission(m, roles.Read) {
				var menu = &menu{Menu: m}
				menu.URL = m.ItemUrl(this, item, parents...)
				menu.Label = this.Tt(m)
				addMenu(menu, menu.GetSubMenusForItem(this, item))
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}

		parent.sortChildren()
	}

	if this.Resource != nil && (this.Resource.Config.Singleton || this.ResourceID != nil) {
		addMenu(globalMenu, this.Resource.GetItemMenusOf(this, item))
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
			if m.Disablers.Disabled(MenuMain, m, this) {
				continue
			}
			if m.EnabledFunc != nil {
				if !m.EnabledFunc(m, this) {
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
	Label  string
	Active bool
}

type scopeMenu struct {
	Group, Label string
	Scopes       []scope
}

// GetScopes get scopes from current context
func (this *Context) GetScopes(advanced bool) (menus []*scopeMenu) {
	if this.Resource == nil {
		return
	}

	scopes := this.Request.URL.Query()["scope[]"]

OUT:
	for _, s := range this.Scheme.MustGetScopes() {
		if advanced != s.Advanced(this) {
			continue
		}
		if s.Visible != nil && !s.Visible(this) {
			continue
		}

		menu := scope{Scope: s, Label: s.GetLabel(this)}

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
				menus = append(menus, &scopeMenu{Group: menu.Group, Label: menu.GetGroupLabel(this), Scopes: []scope{menu}})
			} else {
				menus = append(menus, &scopeMenu{Scopes: []scope{menu}})
			}
		}
	}
	return menus
}

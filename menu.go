package admin

import (
	"path"

	"github.com/aghape/core"
	"github.com/aghape/roles"
)

// GetMenus get all sidebar menus for admin
func (admin Admin) GetMenus() []*Menu {
	return admin.menus
}

// AddMenu add a menu to admin sidebar
func (admin *Admin) AddMenu(menu *Menu) *Menu {
	menu.prefix = admin.Config.MountPath
	admin.menus = appendMenu(admin.menus, menu.Ancestors, menu)
	return menu
}

// GetMenu get sidebar menu with name
func (admin Admin) GetMenu(name string) *Menu {
	return getMenu(admin.menus, name)
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar Menu
////////////////////////////////////////////////////////////////////////////////

// Menu admin sidebar menu definiation
type Menu struct {
	Name         string
	Label        string
	LabelFunc    func() string
	Link         string
	Icon         string
	RelativePath string
	Priority     int
	Ancestors    []string
	Permissioner core.Permissioner
	Permission   *roles.Permission
	Class        string
	Enabled      func(menu *Menu, context *Context) bool
	Resource     *Resource

	subMenus []*Menu
	prefix   string
	MakeLink func(context *Context, args ...interface{}) string
}

// GetLabel return menu's Label
func (menu Menu) GetLabel() string {
	if menu.LabelFunc != nil {
		return menu.LabelFunc()
	}
	if menu.Label != "" {
		return menu.Label
	}
	return I18NGROUP + ".menus." + menu.Name
}

// GetLabel return menu's Label
func (menu Menu) GetIcon() string {
	if menu.Icon != "" {
		return menu.Icon
	}
	return menu.Name
}

// URL return menu's URL
func (menu Menu) RealURL() string {
	if menu.Link != "" {
		return menu.Link
	}

	if (menu.prefix != "") && (menu.RelativePath != "") {
		return path.Join(menu.prefix, menu.RelativePath)
	}

	return menu.RelativePath
}

// URL return menu's URL
func (menu Menu) URL(context *Context, args ...interface{}) string {
	if menu.MakeLink != nil {
		return menu.MakeLink(context, args...)
	}

	if menu.Link != "" {
		return menu.Link
	}

	return context.GenURL(menu.RelativePath)
}

// HasPermission check menu has permission or not
func (menu Menu) HasPermissionE(mode roles.PermissionMode, context *core.Context) (bool, error) {
	if menu.Permission != nil {
		var roles_ = []interface{}{}
		for _, role := range context.Roles {
			roles_ = append(roles_, role)
		}
		return roles.HasPermissionDefaultE(true, menu.Permission, mode, roles_...)
	}

	if menu.Permissioner != nil {
		return core.HasPermissionDefaultE(true, menu.Permissioner, mode, context)
	}

	return true, roles.ErrDefaultPermission
}

// GetSubMenus get submenus for a menu
func (menu *Menu) GetSubMenus() []*Menu {
	return menu.subMenus
}

func getMenu(menus []*Menu, name string) *Menu {
	for _, m := range menus {
		if m.Name == name {
			return m
		}

		if len(m.subMenus) > 0 {
			if mc := getMenu(m.subMenus, name); mc != nil {
				return mc
			}
		}
	}

	return nil
}

func generateMenu(menus []string, menu *Menu) *Menu {
	menuCount := len(menus)
	for index := range menus {
		menu = &Menu{Name: menus[menuCount-index-1], subMenus: []*Menu{menu}}
	}

	return menu
}

func appendMenu(menus []*Menu, ancestors []string, menu *Menu) []*Menu {
	if len(ancestors) > 0 {
		for _, m := range menus {
			if m.Name != ancestors[0] {
				continue
			}

			if len(ancestors) > 1 {
				m.subMenus = appendMenu(m.subMenus, ancestors[1:], menu)
			} else {
				m.subMenus = appendMenu(m.subMenus, []string{}, menu)
			}

			return menus
		}
	}

	var newMenu = generateMenu(ancestors, menu)
	var added bool
	if len(menus) == 0 {
		menus = append(menus, newMenu)
	} else if newMenu.Priority > 0 {
		for idx, menu := range menus {
			if menu.Priority > newMenu.Priority || menu.Priority <= 0 {
				menus = append(menus[0:idx], append([]*Menu{newMenu}, menus[idx:]...)...)
				added = true
				break
			}
		}
		if !added {
			menus = append(menus, menu)
		}
	} else {
		if newMenu.Priority < 0 {
			for idx := len(menus) - 1; idx >= 0; idx-- {
				menu := menus[idx]
				if menu.Priority < newMenu.Priority || menu.Priority == 0 {
					menus = append(menus[0:idx+1], append([]*Menu{newMenu}, menus[idx+1:]...)...)
					added = true
					break
				}
			}

			if !added {
				menus = append(menus, menu)
			}
		} else {
			for idx := len(menus) - 1; idx >= 0; idx-- {
				menu := menus[idx]
				if menu.Priority >= 0 {
					menus = append(menus[0:idx+1], append([]*Menu{newMenu}, menus[idx+1:]...)...)
					added = true
					break
				}
			}

			if !added {
				menus = append([]*Menu{menu}, menus...)
			}
		}
	}

	return menus
}

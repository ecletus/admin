package admin

import (
	"github.com/ecletus/roles"

	"github.com/ecletus/core"
)

// GetMenus get all sidebar itemMenus for admin
func (this Admin) GetMenus() []*Menu {
	return this.menus
}

// AddItemMenu add a menu to admin sidebar
func (this *Admin) AddMenu(menu *Menu) *Menu {
	menu.prefix = this.Config.MountPath
	this.menus = appendMenu(this.menus, menu.Ancestors, menu)
	return menu
}

// GetItemMenu get sidebar menu with name
func (this Admin) GetMenu(name string) *Menu {
	return getMenu(this.menus, name)
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar Menu
////////////////////////////////////////////////////////////////////////////////

// Menu admin sidebar menu definiation
type Menu struct {
	Name         string
	Label        string
	LabelKey     string
	LabelFunc    func() string
	URI          string
	Icon         string
	MdlIcon      string
	Priority     int
	Ancestors    []string
	Permissioner core.Permissioner
	Permission   *roles.Permission
	Class        string
	Enabled      func(menu *Menu, context *Context) bool
	Resource     *Resource
	BaseResource *Resource

	subMenus []*Menu
	prefix   string
	MakeLink func(context *Context, args ...interface{}) string
	AjaxLoad bool
}

func (menu Menu) GetLabelPair() (keys []string, value string) {
	if menu.LabelFunc != nil {
		return []string{menu.LabelFunc()}, menu.Label
	}
	if menu.LabelKey != "" {
		return []string{menu.LabelKey}, menu.Label
	}
	res := menu.Resource
	if res == nil {
		res = menu.BaseResource
	}
	if res != nil {
		keys = append(keys, res.I18nPrefix+".itemMenus."+menu.Name)
	}
	keys = append(keys, I18NGROUP+".itemMenus."+menu.Name)
	if menu.Label != "" {
		value = menu.Label
	} else {
		value = menu.Name
	}
	return
}

// GetLabel return menu's Label
func (menu Menu) GetIcon() string {
	if menu.Icon != "" {
		return menu.Icon
	}
	return menu.Name
}

// URL return menu's URL
func (menu *Menu) URL(context *Context, args ...interface{}) string {
	if menu.MakeLink != nil {
		return menu.MakeLink(context, args...)
	}

	if menu.URI != "" {
		if menu.URI[0] == '/' {
			return context.Path(menu.URI)
		}
		return menu.URI
	}
	return ""
}

// HasContextPermission check menu has permission or not
func (menu Menu) HasPermission(mode roles.PermissionMode, context *core.Context) (perm roles.Perm) {
	if menu.Permissioner != nil {
		return menu.Permissioner.HasPermission(mode, context)
	}

	if menu.Permission != nil {
		return menu.Permission.HasPermission(context, mode, context.Roles.Interfaces()...)
	}

	return
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
		menu = &Menu{
			Name:     menus[menuCount-index-1],
			subMenus: []*Menu{menu},
		}
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

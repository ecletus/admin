package admin

import (
	"github.com/ecletus/roles"
)

const (
	MenuMain = "main"
)

// GetMenus get all sidebar itemMenus for admin
func (this Admin) GetMenus() []*Menu {
	return this.menus
}

// AddItemMenu add a menu to admin sidebar
func (this *Admin) AddMenu(menu *Menu) *Menu {
	menu.prefix = this.Config.MountPath
	this.menus = appendMenu(this.menus, menu.Ancestors, menu)
	for _, cb := range this.onMenuAdded {
		cb(menu)
	}
	menu.DefaultDenyMode = func(ctx *Context) bool {
		return this.DefaultDenyMode
	}
	return menu
}

// GetItemMenu get sidebar menu with name
func (this *Admin) OnMenuAdded(name string, cb func(menu *Menu)) {
	for _, m := range this.menus {
		if m.Name == name {
			cb(m)
			return
		}
	}
	this.onMenuAdded = append(this.onMenuAdded, cb)
}

// GetItemMenu get sidebar menu with name
func (this Admin) GetMenu(name string) *Menu {
	return getMenu(this.menus, name)
}

// //////////////////////////////////////////////////////////////////////////////
// Sidebar Menu
// //////////////////////////////////////////////////////////////////////////////

type MenuDisablers map[string]func(menu *Menu, context *Context) bool

func (this *MenuDisablers) Set(name string, disabler func(menu *Menu, context *Context) bool) {
	if *this == nil {
		*this = map[string]func(menu *Menu, context *Context) bool{}
	}
	(*this)[name] = disabler
}

func (this *MenuDisablers) Disabled(name string, menu *Menu, context *Context) (ok bool) {
	if *this == nil {
		return
	}

	var disabler func(menu *Menu, context *Context) bool
	disabler, ok = (*this)[name]
	if ok {
		return disabler(menu, context)
	}
	return
}

// Menu admin sidebar menu definiation
type Menu struct {
	Name            string
	Label           string
	LabelKey        string
	LabelFunc       func() string
	URI             string
	Icon            string
	MdlIcon         string
	Priority        int
	Ancestors       []string
	Permissioner    Permissioner
	Permission      *roles.Permission
	Class           string
	EnabledFunc     func(menu *Menu, context *Context) bool
	ItemEnabledFunc func(menu *Menu, context *Context, item interface{}) bool
	Disablers       MenuDisablers
	Resource        *Resource
	BaseResource    *Resource

	subMenus        []*Menu
	prefix          string
	MakeLink        func(context *Context, args ...interface{}) string
	MakeItemLink    func(context *Context, item interface{}, args ...interface{}) string
	AjaxLoad        bool
	Dir             bool
	DefaultDenyMode func(ctx *Context) bool
	Parent          *Menu
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

func (menu *Menu) Enabled(context *Context) bool {
	if menu.EnabledFunc != nil {
		return menu.EnabledFunc(menu, context)
	}
	return true
}

func (menu *Menu) ItemEnabled(context *Context, item interface{}) bool {
	if !menu.Enabled(context) {
		return false
	}
	if menu.ItemEnabledFunc != nil {
		return menu.ItemEnabledFunc(menu, context, item)
	}
	return true
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

// ItemUrl return item menu's URL
func (menu *Menu) ItemUrl(context *Context, item interface{}, args ...interface{}) string {
	if menu.MakeItemLink != nil {
		return menu.MakeItemLink(context, item, args...)
	}
	return menu.URL(context, args...)
}

// AdminHasContextPermission check menu has permission or not
func (menu Menu) AdminHasPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	if menu.Dir {
		return roles.ALLOW
	}

	if menu.Permissioner != nil {
		return menu.Permissioner.AdminHasPermission(mode, context)
	}

	if menu.Permission != nil {
		return menu.Permission.HasPermission(context, mode, context.Roles.Interfaces()...)
	}

	if menu.DefaultDenyMode != nil && menu.DefaultDenyMode(context) {
		return roles.DENY
	}

	return
}

// GetSubMenus get submenus for a menu
func (menu *Menu) GetSubMenus() []*Menu {
	return menu.subMenus
}

// GetMenu return submenu from name if exists, other else, nil
func (menu *Menu) GetMenu(name string) *Menu {
	for _, m := range menu.subMenus {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func (menu *Menu) GetSubMenusForItem(ctx *Context, item interface{}) (menus []*Menu) {
	for _, m := range menu.subMenus {
		if m.ItemEnabled(ctx, item) {
			menus = append(menus, m)
		}
	}
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
	var (
		menuCount = len(menus)
		old       *Menu
	)
	for index := range menus {
		old = menu
		menu = &Menu{
			Name:     menus[menuCount-index-1],
			subMenus: []*Menu{old},
			Dir:      true,
		}
		old.Parent = menu
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

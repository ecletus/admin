package admin

import "github.com/jinzhu/inflection"

func (res *Resource) CreateMenu(plural bool) *Menu {
	menuName := res.Name

	if plural {
		menuName = inflection.Plural(menuName)
	}

	menu := &Menu{
		Name:         menuName,
		Label:        res.GetLabelKey(plural),
		Permissioner: res,
		Priority:     res.Config.Priority,
		Ancestors:    res.Config.Menu,
		RelativePath: res.GetIndexURI(),
		Enabled:      res.Config.MenuEnabled,
		Resource:     res,
	}

	if res.ParentResource != nil {
		menu.MakeLink = func(context *Context, args ...interface{}) string {
			var parentKeys []string
			for _, arg := range args {
				switch t := arg.(type) {
				case string:
					if t != "" {
						parentKeys = append(parentKeys, t)
					}
				case []string:
					parentKeys = append(parentKeys, t...)
				}
			}
			if len(parentKeys) == 0 {
				return res.GetContextIndexURI(context.Context, context.ParentResourceID...)
			}
			return res.GetContextIndexURI(context.Context, parentKeys...)
		}
	}

	return menu
}

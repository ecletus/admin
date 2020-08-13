package admin

import "github.com/moisespsena-go/aorm"

func (this *Resource) CreateMenu(plural bool) *Menu {
	menuName := this.Name

	if plural {
		menuName = this.PluralName
	}

	menu := &Menu{
		Name:         menuName,
		Label:        menuName,
		LabelKey:     this.GetLabelKey(plural),
		Permissioner: this,
		Priority:     this.Config.Priority,
		Ancestors:    this.Config.Menu,
		URI:          this.GetIndexURI(),
		Enabled:      this.Config.MenuEnabled,
		Resource:     this,
		BaseResource: this,
		subMenus:     make([]*Menu, 0),
	}

	if this.ParentResource != nil {
		menu.MakeLink = func(context *Context, args ...interface{}) string {
			var parentKeys = aorm.IDSlice(args...)
			if len(parentKeys) == 0 {
				return this.GetContextIndexURI(context.Context, context.ParentResourceID...)
			}
			return this.GetContextIndexURI(context.Context, parentKeys...)
		}
	}

	return menu
}

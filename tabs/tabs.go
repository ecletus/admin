package tabs

import (
	"github.com/aghape/core"
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/admin"
)

type Tab struct {
	Title    string
	Path     string
	TitleKey string
	Handler  func(res *admin.Resource, context *core.Context, db *aorm.DB) *aorm.DB
	Default  bool
}

func (s *Tab) URL(res *admin.Resource, context *core.Context) string {
	if s.Default {
		return res.GetContextIndexURI(context)
	}
	return res.GetContextIndexURI(context) + "/" + s.Path
}

type Tabs []*Tab

type TabsData struct {
	Tabs   Tabs
	ByPath map[string]*Tab
}

func GetTabPath(context *core.Context) string {
	if scope, ok := context.Data().GetOk(KEY_TAB); ok {
		return scope.(*Tab).Path
	}
	return ""
}

func GetTab(context *core.Context) *Tab {
	if tab, ok := context.Data().GetOk(KEY_TAB); ok {
		return tab.(*Tab)
	}
	return nil
}

func TabHandler(res *admin.Resource, config *admin.RouteConfig, indexHandler admin.Handler, scope *Tab) *admin.RouteHandler {
	return admin.NewHandler(func(c *admin.Context) {
		c.Breadcrumbs().Append(core.NewBreadcrumb(res.GetContextIndexURI(c.Context), res.GetLabelKey(true), ""))
		c.Data().Set("page_title", c.T(scope.TitleKey, scope.Title))
		c.Data().Set(KEY_TAB, scope)
		indexHandler(c)
	}, config)
}

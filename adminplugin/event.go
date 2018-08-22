package adminplugin

import (
	"github.com/aghape/admin"
	"github.com/aghape/plug"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-route"
)

var (
	E_ADMIN                = PREFIX + ".admin"
	E_ADMIN_DONE           = E_ADMIN + ".done"
	E_ADMIN_FUNC_MAP       = E_ADMIN + ".funcMap"
	E_ADMIN_ROUTE          = E_ADMIN + ".route"
	E_ADMIN_INIT_RESOURCES = E_ADMIN + ".initResources"
)

type AdminEvent struct {
	plug.PluginEventInterface
	Admin       *admin.Admin
	AdminName   string
	PluginEvent plug.PluginEventInterface
}

type AdminFuncMapEvent struct {
	*AdminEvent
}

func (afm *AdminFuncMapEvent) Register(name string, value interface{}) {
	afm.Admin.RegisterFuncMap(name, value)
}

type AdminRouterEvent struct {
	*AdminEvent
	Mux *route.Mux
}

func (are *AdminRouterEvent) Router() *admin.Router {
	return are.Admin.Router
}

func EAdmin(adminKey string) string {
	if adminKey == "" {
		panic("adminKey is blank")
	}
	return E_ADMIN + ":" + adminKey
}

func EDone(adminKey string) string {
	if adminKey == "" {
		panic("adminKey is blank")
	}
	return E_ADMIN_DONE + ":" + adminKey
}

func EFuncMap(adminKey string) string {
	if adminKey == "" {
		panic("adminKey is blank")
	}
	return E_ADMIN_FUNC_MAP + ":" + adminKey
}

func ERoute(adminName string) string {
	if adminName == "" {
		panic("AdminName is blank")
	}
	return E_ADMIN_ROUTE + ":" + adminName
}

func EInitResources(adminName string) string {
	if adminName == "" {
		panic("AdminName is blank")
	}
	return E_ADMIN_INIT_RESOURCES + ":" + adminName
}

func (admins *Admins) Trigger(d plug.PluginEventDispatcherInterface) error {
	return admins.Each(func(adminName string, Admin *admin.Admin) (err error) {
		e := &AdminEvent{plug.NewPluginEvent(E_ADMIN), Admin, adminName, nil}
		if err = d.TriggerPlugins(e); err != nil {
			return errwrap.Wrap(err, "Admin %q: event %q", adminName, e.Name())
		}
		return nil
	})
}

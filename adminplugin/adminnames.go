package adminplugin

import (
	"github.com/aghape/plug"
)

type AdminNames struct {
	Names []string
}

func (dn *AdminNames) EachE(cb func(adminName string) error) (err error) {
	for _, name := range dn.Names {
		err = cb(name)
		if err != nil {
			return
		}
	}
	return
}

func (dn *AdminNames) Each(cb func(adminName string)) {
	dn.EachE(func(adminName string) error {
		cb(adminName)
		return nil
	})
}

func (dn *AdminNames) EachOrDefaultE(cb func(adminName string) error) (err error) {
	if len(dn.Names) == 0 {
		dn.Names = []string{DEFAULT_ADMIN}
	}
	return dn.EachE(cb)
}

func (dn *AdminNames) EachOrDefault(cb func(adminName string)) {
	dn.EachOrDefaultE(func(adminName string) error {
		cb(adminName)
		return nil
	})
}

func (a *AdminNames) GetNames() []string {
	if len(a.Names) == 0 {
		a.Names = []string{DEFAULT_ADMIN}
	}
	return a.Names
}

func (a *AdminNames) on(ef func(name string) string, dis plug.EventDispatcherInterface, cb interface{}) (err error) {
	for _, name := range a.GetNames() {
		err = dis.OnE(ef(name), cb)
		if err != nil {
			return
		}
	}
	return
}

func (a *AdminNames) OnAdmin(dis plug.EventDispatcherInterface, cb func(e *AdminEvent)) (err error) {
	return a.on(EAdmin, dis, func(e plug.PluginEventInterface) {
		cb(e.(*AdminEvent))
	})
}

func (a *AdminNames) OnAdminE(dis plug.EventDispatcherInterface, cb func(e *AdminEvent) error) (err error) {
	return a.on(EAdmin, dis, func(e plug.PluginEventInterface) error {
		return cb(e.(*AdminEvent))
	})
}

func (a *AdminNames) OnInitResources(dis plug.EventDispatcherInterface, cb func(e *AdminEvent)) (err error) {
	return a.on(EInitResources, dis, func(e plug.PluginEventInterface) {
		cb(e.(*AdminEvent))
	})
}

func (a *AdminNames) OnInitResourcesE(dis plug.EventDispatcherInterface, cb func(e *AdminEvent) error) (err error) {
	return a.on(EInitResources, dis, func(e plug.PluginEventInterface) error {
		return cb(e.(*AdminEvent))
	})
}

func (a *AdminNames) OnFuncMap(dis plug.EventDispatcherInterface, cb func(e *AdminFuncMapEvent)) (err error) {
	return a.on(EFuncMap, dis, func(e plug.PluginEventInterface) {
		cb(e.(*AdminFuncMapEvent))
	})
}

func (a *AdminNames) OnFuncMapE(dis plug.EventDispatcherInterface, cb func(e *AdminFuncMapEvent) error) (err error) {
	return a.on(EFuncMap, dis, func(e plug.PluginEventInterface) error {
		return cb(e.(*AdminFuncMapEvent))
	})
}

func (a *AdminNames) OnRoute(dis plug.EventDispatcherInterface, cb func(e *AdminRouterEvent)) (err error) {
	return a.on(ERoute, dis, func(e plug.PluginEventInterface) {
		cb(e.(*AdminRouterEvent))
	})
}

func (a *AdminNames) OnRouteE(dis plug.EventDispatcherInterface, cb func(e *AdminRouterEvent) error) (err error) {
	return a.on(ERoute, dis, func(e plug.PluginEventInterface) error {
		return cb(e.(*AdminRouterEvent))
	})
}

package adminplugin

import (
	"path/filepath"

	"github.com/aghape/admin"
	"github.com/aghape/plug"
	"github.com/moisespsena/go-default-logger"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-path-helpers"
	"github.com/aghape/router"
	"strings"
)

const DEFAULT_ADMIN = "default"

var log = defaultlogger.NewLogger(PREFIX)

type DisAdminNames struct {
	Dis       plug.EventDispatcherInterface
	Names     *AdminNames
	AdminsKey string
}

func NewDisAdminNames(names *AdminNames, dis plug.EventDispatcherInterface, adminsKey ...string) *DisAdminNames {
	if len(adminsKey) == 0 {
		adminsKey = append(adminsKey, "")
	}
	return &DisAdminNames{dis, names, adminsKey[0]}
}

func (dn *DisAdminNames) EachOrAll(e plug.PluginEventInterface, cb func(adminName string, Admin *admin.Admin) (err error)) (err error) {
	Admins := e.Options().GetInterface(dn.AdminsKey).(*Admins)
	if len(dn.Names.Names) == 0 {
		for adminName, Admin := range Admins.ByName {
			err = cb(adminName, Admin)
			if err != nil {
				return
			}
		}
	} else {
		for _, adminName := range dn.Names.Names {
			err = cb(adminName, Admins.ByName[adminName])
			if err != nil {
				return
			}
		}
	}
	return
}

func (dn *DisAdminNames) OnE(cb func(e *AdminEvent) error) *DisAdminNames {
	dn.Names.EachOrDefault(func(adminName string) {
		dn.Dis.On(EAdmin(adminName), func(e plug.PluginEventInterface) error {
			return cb(e.(*AdminEvent))
		})
	})
	return dn
}

func (dn *DisAdminNames) On(cb func(e *AdminEvent)) *DisAdminNames {
	return dn.OnE(func(e *AdminEvent) error {
		cb(e)
		return nil
	})
}

func (dn *DisAdminNames) OnDoneE(cb func(e *AdminEvent) error) *DisAdminNames {
	dn.Names.EachOrDefault(func(adminName string) {
		dn.Dis.On(EDone(adminName), func(e plug.PluginEventInterface) error {
			return cb(e.(*AdminEvent))
		})
	})
	return dn
}

func (dn *DisAdminNames) OnDone(cb func(e *AdminEvent)) *DisAdminNames {
	return dn.OnDoneE(func(e *AdminEvent) error {
		cb(e)
		return nil
	})
}

func (dn *DisAdminNames) OnInitResourcesE(cb func(e *AdminEvent) error) *DisAdminNames {
	dn.Names.EachOrDefault(func(adminName string) {
		dn.Dis.On(EInitResources(adminName), func(e plug.PluginEventInterface) error {
			return cb(e.(*AdminEvent))
		})
	})
	return dn
}

func (dn *DisAdminNames) OnInitResources(cb func(e *AdminEvent)) *DisAdminNames {
	return dn.OnInitResourcesE(func(e *AdminEvent) error {
		cb(e)
		return nil
	})
}

func (dn *DisAdminNames) OnFuncMapE(cb func(e *AdminFuncMapEvent) error) *DisAdminNames {
	dn.Names.EachOrDefault(func(adminName string) {
		dn.Dis.On(EFuncMap(adminName), func(e plug.PluginEventInterface) error {
			return cb(e.(*AdminFuncMapEvent))
		})
	})
	return dn
}

func (dn *DisAdminNames) OnFuncMap(cb func(e *AdminFuncMapEvent)) *DisAdminNames {
	return dn.OnFuncMapE(func(e *AdminFuncMapEvent) error {
		cb(e)
		return nil
	})
}

func (dn *DisAdminNames) OnRouterE(cb func(e *AdminRouterEvent) error) *DisAdminNames {
	dn.Names.EachOrDefault(func(adminName string) {
		dn.Dis.On(ERoute(adminName), func(e plug.PluginEventInterface) error {
			return cb(e.(*AdminRouterEvent))
		})
	})
	return dn
}

func (dn *DisAdminNames) OnRouter(cb func(e *AdminRouterEvent)) *DisAdminNames {
	return dn.OnRouterE(func(e *AdminRouterEvent) error {
		cb(e)
		return nil
	})
}

func (adms *Admins) Each(cb func(adminName string, Admin *admin.Admin) (err error)) (err error) {
	for adminName, Admin := range adms.ByName {
		err = cb(adminName, Admin)
		if err != nil {
			return errwrap.Wrap(err, "Admin %q", adminName)
		}
	}
	return nil
}

type Plugin struct {
	plug.EventDispatcher
	AdminsKey string
}

func (p *Plugin) RequireOptions() []string {
	return []string{p.AdminsKey}
}

func (p *Plugin) NameSpace() string {
	return filepath.Dir(path_helpers.GetCalledDir(false))
}

func (p *Plugin) AssetsRootPath() string {
	return filepath.Dir(path_helpers.GetCalledDir(true))
}

func (p *Plugin) OnRegister() {
	adminsCalled := map[string]bool {}
	p.On(E_ADMIN, func(e plug.PluginEventInterface) (err error) {
		adminEvent := e.(*AdminEvent)
		if _, ok := adminsCalled[adminEvent.AdminName]; ok {
			return nil
		}
		adminsCalled[adminEvent.AdminName] = true
		adminName, Admin := adminEvent.AdminName, adminEvent.Admin
		log.Debugf("trigger AdminEvent")
		if err = e.PluginDispatcher().TriggerPlugins(&AdminEvent{plug.NewPluginEvent(EAdmin(adminName)), Admin, adminName, e}); err != nil {
			return errwrap.Wrap(err, "AdminEvent")
		}
		log.Debugf("trigger AdminInitResourcesEvent")
		if err = e.PluginDispatcher().TriggerPlugins(&AdminEvent{plug.NewPluginEvent(EInitResources(adminName)), Admin, adminName, e}); err != nil {
			return errwrap.Wrap(err, "AdminInitResourcesEvent")
		}
		log.Debugf("trigger AdminFuncMapEvent")
		if err = e.PluginDispatcher().TriggerPlugins(&AdminFuncMapEvent{&AdminEvent{plug.NewPluginEvent(EFuncMap(adminName)), Admin, adminName, e}}); err != nil {
			return errwrap.Wrap(err, "AdminFuncMapEvent")
		}
		log.Debugf("trigger AdminDone")
		if err = Admin.TriggerDone(&admin.AdminEvent{plug.NewEvent(admin.E_DONE), Admin}); err != nil {
			return errwrap.Wrap(err, "Admin.Done")
		}
		if err = e.PluginDispatcher().TriggerPlugins(&AdminEvent{plug.NewPluginEvent(EDone(adminName)), Admin, adminName, e}); err != nil {
			return errwrap.Wrap(err, "AdminDone")
		}
		return nil
	})

	router.OnRouteE(p, func(e *router.RouterEvent) (err error) {
		admins := e.Options().GetInterface(p.AdminsKey).(*Admins)
		err = admins.Trigger(e.PluginDispatcher())
		if err != nil {
			return errwrap.Wrap(err, "Trigger Admins [%s]", strings.Join(admins.Names(), ", "))
		}
		return admins.Each(func(adminName string, Admin *admin.Admin) error {
			log.Debugf("[admin=%q] mounted on %v", adminName, Admin.Config.MountPath)
			mux := Admin.NewServeMux()
			e.Router.Mux.Mount(Admin.Config.MountPath, mux)
			return errwrap.Wrap(e.PluginDispatcher().TriggerPlugins(&AdminRouterEvent{&AdminEvent{
				plug.NewPluginEvent(ERoute(adminName)), Admin,
				adminName, e}, mux}),
				"AdminRouterEvent")
		})
	})
}

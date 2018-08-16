package tabs

import (
	"github.com/aghape/admin/adminplugin"
	"github.com/aghape/plug"
)

type Plugin struct {
	plug.EventDispatcher
	AdminsKey string
}

func (p *Plugin) RequireOptions() []string {
	return []string{p.AdminsKey}
}

type interfaceGetter interface {
	GetInterface(key interface{}) interface{}
}

type interfaceStringGetter interface {
	GetInterface(key string) interface{}
}

type interfaceStringDefaultsGetter interface {
	GetInterface(key string, defaul ...interface{}) interface{}
}

func (p *Plugin) Init(options *plug.Options) {
	Admins := options.GetInterface(p.AdminsKey).(*adminplugin.Admins)
	Admins.OnAdmin(p, func(e *adminplugin.AdminEvent) {
		e.Admin.RegisterFuncMap("admin_tab", func(v interface{}) *Tab {
			if vi, ok := v.(interfaceGetter); ok {
				return vi.GetInterface(KEY_TAB).(*Tab)
			} else if vi, ok := v.(interfaceStringDefaultsGetter); ok {
				return vi.GetInterface(KEY_TAB).(*Tab)
			}
			return v.(interfaceStringGetter).GetInterface(KEY_TAB).(*Tab)
		})
		e.Admin.RegisterFuncMap("admin_tabs", func(v interface{}) *TabsData {
			if vi, ok := v.(interfaceGetter); ok {
				return vi.GetInterface(KEY_TABS).(*TabsData)
			} else if vi, ok := v.(interfaceStringDefaultsGetter); ok {
				return vi.GetInterface(KEY_TABS).(*TabsData)
			}
			return v.(interfaceStringGetter).GetInterface(KEY_TABS).(*TabsData)
		})
	})
}

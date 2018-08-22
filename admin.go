package admin

import (
	"github.com/moisespsena/go-assetfs"
	"github.com/moisespsena/go-edis"
	"github.com/moisespsena/go-route"
	"github.com/moisespsena/template/html/template"
	"github.com/aghape/assets"
	qorconfig "github.com/aghape/core/config"
	"github.com/aghape/session"
)

type AdminConfig struct {
	*qorconfig.Config
	MountPath  string
	AssetFS    assetfs.Interface
	TemplateFS assetfs.Interface
	StaticFS   assetfs.Interface
	Data       qorconfig.OtherConfig
}

func NewConfig(config *qorconfig.Config) *AdminConfig {
	return &AdminConfig{Config: config}
}

type Router = route.Mux

// Admin is a struct that used to generate admin/api interface
type Admin struct {
	edis.EventDispatcher
	SiteName       string
	SiteTitle      string
	Config         *AdminConfig
	I18n           I18n
	Auth           Auth
	SessionManager session.ManagerInterface
	*Transformer

	TemplateFS          assetfs.Interface
	StaticFS            assetfs.Interface
	menus               []*Menu
	Resources           map[string]*Resource
	ResourcesByParam    map[string]*Resource
	ResourcesByUID      map[string]*Resource
	searchResources     []*Resource
	Router              *route.Mux
	funcMaps            template.FuncMap
	metaConfigorMaps    map[string]func(*Meta)
	NewContextCallbacks []func(context *Context) *Context
	ViewPaths           map[string]bool
	Data                qorconfig.OtherConfig
}

// ResourceNamer is an interface for models that defined method `ResourceName`
type ResourceNamer interface {
	ResourceName() string
}

// New new admin with configuration
func New(config *AdminConfig) *Admin {
	admin := &Admin{
		Config:           config,
		funcMaps:         make(template.FuncMap),
		Router:           route.NewMux(PKG),
		metaConfigorMaps: metaConfigorMaps,
		Transformer:      DefaultTransformer,
		Resources:        make(map[string]*Resource),
		ResourcesByParam: make(map[string]*Resource),
		ResourcesByUID:   make(map[string]*Resource),
		Data:             config.Data,
	}

	admin.SetDispatcher(admin)

	admin.Router.Intersept(&route.Middleware{
		Name:    PKG,
		Handler: admin.routeInterseptor,
	})

	if config.TemplateFS != nil {
		admin.TemplateFS = config.TemplateFS
	} else {
		admin.TemplateFS = assets.TemplateFS(config.AssetFS).NameSpace("admin")
	}
	if config.StaticFS != nil {
		admin.StaticFS = config.StaticFS
	} else {
		admin.StaticFS = assets.StaticFS(config.AssetFS).NameSpace("admin")
	}

	if config.Data == nil {
		config.Data = qorconfig.NewOtherConfig()
	}

	admin.registerCompositePrimaryKeyCallback()
	return admin
}

func (admin *Admin) Init() {
	if admin.Config.MountPath == "" {
		if admin.SiteName == "" {
			admin.Config.MountPath = "/admin"
		} else {
			admin.Config.MountPath = "/admin/" + admin.SiteName
		}
	}
}

func (admin *Admin) GetSiteTitle() string {
	if admin.SiteTitle == "" {
		return admin.SiteName
	}
	return admin.SiteTitle
}

func (admin *Admin) AddNewContextCallback(callback func(context *Context) *Context) *Admin {
	admin.NewContextCallbacks = append(admin.NewContextCallbacks, callback)
	return admin
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (admin *Admin) SetSiteTitle(siteName string) {
	admin.SiteTitle = siteName
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (admin *Admin) SetSiteName(siteName string) {
	admin.SiteName = siteName
}

// SetAuth set admin's authorization gateway
func (admin *Admin) SetAuth(auth Auth) {
	admin.Auth = auth
}

// RegisterMetaConfigor register configor for a kind, it will be called when register those kind of metas
func (admin *Admin) RegisterMetaConfigor(kind string, fc func(*Meta)) {
	admin.metaConfigorMaps[kind] = fc
}

// RegisterFuncMap register view funcs, it could be used in view templates
func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}

func (admin *Admin) InitRoutes() *route.Mux {
	for param, res := range admin.ResourcesByParam {
		pattern := "/" + param
		r := res.InitRoutes()
		admin.Router.Mount(pattern, r)
	}
	return admin.Router
}

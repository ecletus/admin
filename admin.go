package admin

import (
	"github.com/ecletus/assets"
	"github.com/ecletus/core"
	qorconfig "github.com/ecletus/core/config"
	"github.com/ecletus/core/helpers"
	"github.com/ecletus/db"
	"github.com/ecletus/session"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/xroute"
	"github.com/moisespsena/go-assetfs"
	"github.com/moisespsena/go-edis"
	"github.com/moisespsena/go-options"
	"github.com/moisespsena/template/html/template"
)

const ROLE = "Admin"

type AdminConfig struct {
	*qorconfig.Config
	MountPath       string
	AssetFS         assetfs.Interface
	TemplateFS      assetfs.Interface
	StaticFS        assetfs.Interface
	Data            options.Options
	FakeDBDialect   string
	ContextFactory  *core.ContextFactory
	UserResourceUID string
}

func NewConfig(config *qorconfig.Config) *AdminConfig {
	return &AdminConfig{Config: config}
}

type Router = xroute.Mux

// Admin is a struct that used to generate admin/api interface
type Admin struct {
	edis.EventDispatcher
	Paged
	Name           string
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
	Router              *xroute.Mux
	funcMaps            template.FuncMap
	metaConfigorMaps    map[string]func(*Meta)
	NewContextCallbacks []func(context *Context) *Context
	ViewPaths           map[string]bool
	Data                options.Options
	Cache               helpers.SyncMap
	SettingsResource    *Resource
	FakeDB              *aorm.DB
	ContextFactory      *core.ContextFactory

	settings settings

	onRouter []func(r xroute.Router)
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
		metaConfigorMaps: metaConfigorMaps,
		Transformer:      DefaultTransformer,
		Resources:        make(map[string]*Resource),
		ResourcesByParam: make(map[string]*Resource),
		ResourcesByUID:   make(map[string]*Resource),
		Data:             config.Data,
	}

	admin.SetDispatcher(admin)

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
		admin.Data = make(options.Options)
	}

	if admin.Config.FakeDBDialect == "" {
		admin.Config.FakeDBDialect = db.DEFAULT_DIALECT
	}
	admin.FakeDB = aorm.FakeDB(admin.Config.FakeDBDialect)

	cache := make(options.Options)
	admin.Data.Set("cache", &cache)

	admin.OnRouter(admin.registerCompositePrimaryKeyCallback)

	return admin
}

func (admin *Admin) OnRouter(f ...func(r xroute.Router)) {
	admin.onRouter = append(admin.onRouter, f...)
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

package admin

import (
	"reflect"

	"github.com/ecletus/about"
	"github.com/ecletus/assets"
	"github.com/ecletus/db"
	"github.com/ecletus/ecletus"
	"github.com/ecletus/roles"
	"github.com/ecletus/session"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/edis"
	"github.com/moisespsena-go/logging"
	"github.com/moisespsena-go/options"
	path_helpers "github.com/moisespsena-go/path-helpers"
	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/core"
	"github.com/ecletus/core/helpers"
	"github.com/ecletus/sites"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/template/html/template"
)

const ROLE = "Admin"

var (
	dir  = path_helpers.GetCalledDir()
	log  = logging.GetOrCreateLogger(dir)
	rlog = logging.GetOrCreateLogger(dir + "@route")
)

type AdminConfig struct {
	*sites.Config
	Dialect         aorm.Dialector
	MountPath       string
	AssetFS         assetfs.Interface
	TemplateFS      assetfs.Interface
	StaticFS        assetfs.Interface
	Data            options.Options
	FakeDBDialect   string
	ContextFactory  *core.ContextFactory
	UserResourceUID string
	SiteAbouter     func(ctx *Context) about.Abouter
	DefaultDenyMode bool
	Ecletus         *ecletus.Ecletus

	Controller       *AdminController
	Public           bool
	DefaultPageTitle func(ctx *Context) string
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
	ResourcesByType     map[reflect.Type][]*Resource
	searchResources     []*Resource
	Router              *xroute.Mux
	funcMaps            template.FuncMap
	metaConfigorMaps    map[string]func(*Meta)
	NewContextCallbacks []func(context *Context)
	ViewPaths           map[string]bool
	Data                options.Options
	Cache               helpers.SyncMap
	SettingsResource    *Resource
	FakeDB              *aorm.DB
	ContextFactory      *core.ContextFactory

	settings settings

	onRouter                    []func(r xroute.Router)
	onPreInitializeResourceMeta []func(meta *Meta)
	onResourceTypeAdded         map[reflect.Type][]func(res *Resource)
	onMenuAdded                 []func(menu *Menu)
	ContextPermissioners        []Permissioner
	DefaultDenyMode             bool

	BeforeAddResourceCallbacks []func(value interface{}, cfg *Config, afterCallback func(func(res *Resource)))
	BeforeResourceInitializeCallbacks,
	AfterResourceInitializeCallbacks []func(res *Resource)

	RouteTree RouteTree
}

// ResourceNamer is an interface for models that defined method `ResourceName`
type ResourceNamer interface {
	ResourceName() string
}

// New new admin with configuration
func New(config *AdminConfig) *Admin {
	if config.DefaultPageTitle == nil {
		config.DefaultPageTitle = func(ctx *Context) string {
			return ctx.Ts(I18NGROUP+".layout.title", "Admin")
		}
	}
	admin := &Admin{
		Config:           config,
		funcMaps:         make(template.FuncMap),
		metaConfigorMaps: metaConfigorMaps,
		Transformer:      DefaultTransformer,
		Resources:        make(map[string]*Resource),
		ResourcesByParam: make(map[string]*Resource),
		ResourcesByUID:   make(map[string]*Resource),
		ResourcesByType:  make(map[reflect.Type][]*Resource),
		Data:             config.Data,
		menus:            make([]*Menu, 0),
		DefaultDenyMode:  config.DefaultDenyMode,
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

	return admin
}

func (this *Admin) BeforeAddResource(f func(value interface{}, cfg *Config, afterCallback func(func(res *Resource)))) {
	this.BeforeAddResourceCallbacks = append(this.BeforeAddResourceCallbacks, f)
}

func (this *Admin) BeforeResourceInitialize(f ...func(res *Resource)) {
	this.BeforeResourceInitializeCallbacks = append(this.BeforeResourceInitializeCallbacks, f...)
}

func (this *Admin) AfterResourceInitialize(f ...func(res *Resource)) {
	this.AfterResourceInitializeCallbacks = append(this.AfterResourceInitializeCallbacks, f...)
}

func (this *Admin) ContextPermissioner(permissioner ...Permissioner) {
	this.ContextPermissioners = append(this.ContextPermissioners, permissioner...)
}

func (this *Admin) AdminHasContextPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	for _, permissioner := range this.ContextPermissioners {
		if perm = permissioner.AdminHasPermission(mode, context); perm != roles.UNDEF {
			return
		}
	}
	if context.Resource != nil {
		if perm = context.Resource.AdminHasContextPermission(mode, context); perm != roles.UNDEF {
			return
		}
	}
	return
}

func (this *Admin) Init() {
	if this.Config.MountPath == "" {
		this.Config.MountPath = "/admin"
	}
}

func (this *Admin) GetSiteTitle() string {
	if this.SiteTitle == "" {
		return this.SiteName
	}
	return this.SiteTitle
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (this *Admin) SetSiteTitle(siteName string) {
	this.SiteTitle = siteName
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (this *Admin) SetSiteName(siteName string) {
	this.SiteName = siteName
}

// SetAuth set admin's authorization gateway
func (this *Admin) SetAuth(auth Auth) {
	this.Auth = auth
}

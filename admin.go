package admin

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/qor/assetfs"
	"github.com/qor/qor"
	qorconfig "github.com/qor/qor/config"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/session"
	"github.com/moisespsena/template/html/template"
)

type AdminConfig struct {
	*qorconfig.Config
	SetupDB        qor.SetupDB
	AssetFS assetfs.Interface
	RootAssetFS assetfs.Interface
}

func NewConfig(config *qorconfig.Config) *AdminConfig {
	return &AdminConfig{Config:config}
}

// Admin is a struct that used to generate admin/api interface
type Admin struct {
	SiteName       string
	SiteTitle      string
	Config         *AdminConfig
	I18n           I18n
	Auth           Auth
	SessionManager session.ManagerInterface
	*Transformer

	AssetFS          assetfs.Interface
	menus            []*Menu
	resources        []*Resource
	resourcesMap     map[interface{}]*Resource
	searchResources  []*Resource
	router           *Router
	funcMaps         template.FuncMap
	metaConfigorMaps map[string]func(*Meta)
	NewContextCallbacks []func(context *Context) *Context
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
		router:           newRouter(),
		metaConfigorMaps: metaConfigorMaps,
		Transformer:      DefaultTransformer,
		resourcesMap:     make(map[interface{}]*Resource),
	}

	if config.RootAssetFS != nil {
		admin.SetRootAssetFS(config.RootAssetFS)
	} else if config.AssetFS != nil {
		admin.SetAssetFS(config.AssetFS)
	} else {
		admin.SetRootAssetFS(assetfs.AssetFS())
	}

	admin.registerCompositePrimaryKeyCallback()
	return admin
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

// SetAssetFS set AssetFS for admin
func (admin *Admin) SetAssetFS(assetFS assetfs.Interface) {
	admin.AssetFS = assetFS
	globalAssetFSes = append(globalAssetFSes, assetFS)

	admin.AssetFS.RegisterPath(filepath.Join(root, "app/views/qor"))
	admin.RegisterViewPath("github.com/qor/admin/views")

	for _, viewPath := range globalViewPaths {
		admin.RegisterViewPath(viewPath)
	}
}

// SetAssetFS set AssetFS for admin
func (admin *Admin) SetRootAssetFS(assetFS assetfs.Interface) {
	admin.SetAssetFS(assetFS.NameSpace("admin"))
}

// RegisterViewPath register view path for admin
func (admin *Admin) RegisterViewPath(pth string) {
	if admin.AssetFS.RegisterPath(filepath.Join(root, "vendor", pth)) != nil {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			if admin.AssetFS.RegisterPath(filepath.Join(gopath, "src", pth)) == nil {
				break
			}
		}
	}
}

// RegisterMetaConfigor register configor for a kind, it will be called when register those kind of metas
func (admin *Admin) RegisterMetaConfigor(kind string, fc func(*Meta)) {
	admin.metaConfigorMaps[kind] = fc
}

// RegisterFuncMap register view funcs, it could be used in view templates
func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}

// GetRouter get router from admin
func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) newResource(value interface{}, config ...*Config) *Resource {
	var configuration *Config
	if len(config) > 0 {
		configuration = config[0]
	}

	if configuration == nil {
		configuration = &Config{}
	}

	res := &Resource{
		Resource:    resource.New(value),
		Config:      configuration,
		cachedMetas: &map[string][]*Meta{},
		admin:       admin,
		filters:     make(map[string]*Filter),
	}

	res.Permission = configuration.Permission

	if configuration.Name != "" {
		res.Name = configuration.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	// Configure resource when initializing
	modelType := utils.ModelType(res.Value)
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceBeforeInitializeInterface); ok {
				injector.ConfigureQorResourceBeforeInitialize(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceBeforeInitializeInterface); ok {
		injector.ConfigureQorResourceBeforeInitialize(res)
	}

	findOneHandler := res.FindOneHandler
	res.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		if context.ResourceID == "" {
			context.ResourceID = res.GetPrimaryValue(context.Request)
		}
		return findOneHandler(result, metaValues, context)
	}

	res.UseTheme("slideout")

	return res
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (admin *Admin) NewResource(value interface{}, config ...*Config) *Resource {
	res := admin.newResource(value, config...)
	//res.Config.Invisible = true
	res.configure()
	return res
}

// AddResource make a model manageable from admin interface
func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	res := admin.newResource(value, config...)
	admin.resources = append(admin.resources, res)

	res.configure()

	if !res.Config.Invisible {
		admin.AddMenu(res.CreateDefaultMenu())
		res.RegisterDefaultRouters()
	}

	admin.resourcesMap[res.ToParam()] = res

	key := utils.TypeId(value)

	if _, ok := admin.resourcesMap[key]; !ok {
		admin.resourcesMap[key] = res
	}

	return res
}

// GetResources get defined resources from admin
func (admin *Admin) GetResources() []*Resource {
	return admin.resources
}

// GetResource get resource with name
func (admin *Admin) GetResource(key interface{}) (resource *Resource) {
	if resource, ok := admin.resourcesMap[key]; ok {
		return resource
	}

	switch name := key.(type) {
	case string:
		for _, res := range admin.resources {
			modelType := utils.ModelType(res.Value)
			// find with defined name first
			if res.ToParam() == name || res.Name == name || modelType.String() == name {
				return res
			}

			// if failed to find, use its model name
			if modelType.Name() == name {
				resource = res
			}
		}
	}

	return
}

// AddSearchResource make a resource searchable from search center
func (admin *Admin) AddSearchResource(resources ...*Resource) {
	admin.searchResources = append(admin.searchResources, resources...)
}

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

// T call i18n backend to translate
func (admin *Admin) T(context *qor.Context, key string, value string, values ...interface{}) template.HTML {
	if len(values) > 1 {
		panic("Values has many args.")
	}

	t := context.GetI18nContext().T(key).Default(value)

	if len(values) == 1 {
		t.Data(values[0])
	}

	return template.HTML(t.Get())
}

// TT call i18n backend to translate template
func (admin *Admin) TT(context *qor.Context, key string, data interface{}, defaul... string) template.HTML {
	t := context.GetI18nContext().T(key).Data(data)
	if len(defaul) > 0 {
		t = t.Default(defaul[0])
	}
	return template.HTML(t.Get())
}

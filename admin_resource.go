package admin

import (
	"path"
	"reflect"
	"strings"

	errwrap "github.com/moisespsena-go/error-wrap"
	"github.com/moisespsena-go/logging"

	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/core"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/fragment"
	"github.com/jinzhu/inflection"
	"github.com/moisespsena-go/edis"

	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
)

var newResourceLog = logging.WithPrefix(log, "new_resource")

func (this *Admin) newResource(value interface{}, config *Config, onUid func(uid string)) *Resource {
	var log = newResourceLog
	if config == nil {
		config = &Config{}
	}

	if config.Protected {
		config.NotMount = true
	}

	var typ reflect.Type

	if value == nil {
		if config.Sub.Parent == nil {
			log.Fatalf("value is nil")
		}

		if field, ok := reflect.TypeOf(config.Sub.Parent.Value).Elem().FieldByName(config.Sub.FieldName); ok {
			if config.Name == "" {
				config.Name = config.Sub.FieldName
				if config.LabelKey == "" {
					config.LabelKey = config.Sub.Parent.I18nPrefix + ".children." + config.Sub.FieldName
				}
			}
			if config.ID == "" {
				if config.Prefix != "" {
					config.ID = config.Prefix + "."
				}
				config.ID += config.Sub.FieldName
			}

			typ = field.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if typ.Kind() == reflect.Slice {
				if config.Param == "" {
					config.Param = config.Sub.FieldName
				}
				typ = typ.Elem()
			}
		} else {
			log.Fatal("resource field `" + config.Sub.FieldName + "` does not exists")
		}
	} else {
		typ, _, _ = aorm.StructTypeOf(reflect.TypeOf(value))
	}

	if config.Wizard != nil {
		config.Invisible = true
	}

	var (
		uid        = config.UID
		uidDefined = uid != ""
		uidSufix   string
	)

	if config.Sub != nil && config.Sub.Parent != nil {
		if !uidDefined {
			uid = config.Sub.Parent.UID
		}
		if config.Name == "" && config.ID == "" && config.Sub.FieldName != "" {
			config.ID = config.Sub.FieldName
			if !uidDefined {
				uidSufix = config.Sub.FieldName
			}
		}

		if config.Param == "" {
			config.Param = utils.ToParamString(config.Sub.FieldName)
		}
	}

	if config.Sub != nil && config.Sub.FieldName != "" {
		if config.ModelStruct == nil {
			config.ModelStruct = config.Sub.Parent.ModelStruct.FieldsByName[config.Sub.FieldName].Model
		}
	}

	if config.ModelStruct == nil {
		config.ModelStruct = aorm.StructOf(typ)
	}

	if !uidDefined {
		if uidSufix == "" {
			if uid != "" {
				uid += "@"
			}
			uid += utils.TypeId(typ)
		} else {
			uid += "#" + uidSufix
		}
	}

	log = logging.WithPrefix(newResourceLog, uid)
	log.Debug("create")

	if !aorm.AcceptTypeForModelStruct(typ) {
		log.Notice("type excluded")
		return nil
	}

	value = reflect.New(typ).Interface()

	if !config.Protected {
		if res, ok := this.ResourcesByUID[uid]; ok {
			if config.Duplicated != nil {
				config.Duplicated(uid, res)
			}
			return res
		}
	}

	if onUid != nil {
		onUid(uid)
	}

	res := &Resource{
		Resource:            resource.New(value, config.ID, uid, config.ModelStruct),
		Config:              config,
		cachedMetas:         &map[string][]*Meta{},
		Admin:               this,
		Resources:           make(map[string]*Resource),
		ResourcesByParam:    make(map[string]*Resource),
		MetaAliases:         make(map[string]*resource.MetaName),
		MetaLinks:           make(map[string]string),
		MetasByName:         make(map[string]*Meta),
		MetasByFieldName:    make(map[string]*Meta),
		Inherits:            make(map[string]*Child),
		RouteHandlers:       make(map[string]*RouteHandler),
		labelKey:            config.LabelKey,
		Param:               config.Param,
		Tags:                &ResourceTags{},
		AllSectionsProvider: &SchemeSectionsProvider{Name: uid + "#all"},
	}

	res.Resource.Permissioner(core.NewPermissioner(func(mode roles.PermissionMode, ctx *core.Context) (perm roles.Perm) {
		return res.adminHasPermission(mode, ContextFromCoreContext(ctx))
	}))

	res.SetLogger(log)

	res.SetDefaultDenyMode(func() bool {
		return this.DefaultDenyMode
	})

	if config.Controller == nil {
		if config.Wizard != nil {
			config.Controller = NewWizardController()
		} else if len(config.CreateWizards) > 0 {
			config.Controller = &ResourceWithCreateWizardController{}
		} else {
			config.Controller = NewCrudSearchIndexController()
		}
	} else if _, ok := config.Controller.(ControllerUpdater); !ok {
		res.ReadOnly = true
	}

	res.Singleton = res.Config.Singleton
	res.ControllerBuilder = &ResourceControllerBuilder{
		Resource:   res,
		Controller: config.Controller,
	}

	var viewController interface{}
	if config.ViewControllerFactory != nil {
		viewController = config.ViewControllerFactory(config.Controller)
	} else {
		viewController = &Controller{controller: config.Controller}
	}

	res.ViewControllerBuilder = &ResourceViewControllerBuilder{
		ResourceController: res.ControllerBuilder,
		Controller:         viewController,
	}

	res.ControllerBuilder.ViewController = res.ViewControllerBuilder

	if _, res.softDelete = res.Value.(SoftDeleter); !res.softDelete {
		_, res.softDelete = res.ModelStruct.FieldsByName[aorm.SoftDeleteFieldDeletedAt]
	}

	res.Scheme = NewScheme(res, "Default")
	res.Resource.SetDispatcher(res)

	if _, ok := value.(fragment.FragmentedModelInterface); ok {
		res.Fragments = NewFragments()
	}

	res.Children = &Inheritances{resource: res}

	if config.ID != "" {
		res.ID = config.ID
		if base := path.Base(res.ID); base != res.ModelStruct.Type.Name() {
			if typ.Name() == "" {
				res.I18nPrefix = res.Config.Sub.Parent.I18nPrefix + "." + base
			} else {
				res.I18nPrefix += "." + base
			}
		}
	}

	res.Router = xroute.NewMux(res.ID)
	if res.Config.Singleton {
		res.ItemRouter = res.Router
	} else {
		res.ItemRouter = xroute.NewMux(res.ID + ":ItemRouter")
	}

	if config.Prefix != "" {
		res.Router.SetPrefix(strings.Replace(config.Prefix, ".", "/", -1))
	}

	if !config.Alone {
		if config.Sub != nil {
			if config.Sub.Parent == nil {
				log.Fatal("parent is nil.")
			}
			res.ParentResource = config.Sub.Parent
		}
	}

	res.Permission = config.Permission

	if config.Name != "" {
		res.Name = utils.HumanizeString(config.Name)
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	if config.PluralName != "" {
		res.PluralName = config.PluralName
	} else {
		res.PluralName = inflection.Plural(res.Name)
	}

	if !config.Alone && res.Param == "" {
		if res.Config.Singleton {
			res.Param = utils.ToParamString(res.Name)
		} else {
			res.Param = utils.ToParamString(res.PluralName)
		}
		if config.Prefix != "" {
			res.Param = utils.ToParamString(config.Prefix + "." + res.Param)
		}
	}

	if !config.Alone {
		res.Parents = resourceParents(res)
		res.PathLevel = len(res.Parents)
		res.ParamName = resourceParamName(res.Parents, utils.ToParamString(res.Param))
		res.paramIDName = resourceParamIDName(res.PathLevel, res.ParamName)

		if config.Sub != nil {
			if config.Sub.FieldName != "" {
				if field, ok := config.Sub.Parent.ModelStruct.FieldsByName[config.Sub.FieldName]; ok {
					if field.Relationship != nil {
						res.SetParentResource(config.Sub.Parent, &resource.ParentRelationship{
							Relationship: *field.Relationship,
						})
					} else {
						res.SetParentResource(config.Sub.Parent, nil)
					}
				} else {
					log.Fatalf("invalid field name %q", config.Sub.FieldName)
				}
			} else if config.Sub.ParentFieldName != "" && !res.ModelStruct.FieldsByName[config.Sub.ParentFieldName].IsPrimaryKey {
				res.SetParentResource(config.Sub.Parent, &resource.ParentRelationship{
					Relationship: *res.ModelStruct.FieldsByName[config.Sub.ParentFieldName].Relationship,
				})
			} else if rel := config.Sub.Relation; rel != nil {
				if rel.AssociationModel == nil {
					rel.AssociationModel = config.Sub.Parent.ModelStruct
				}
				if rel.Model == nil {
					rel.Model = res.ModelStruct
				}

				if rel.ForeignFieldNames == nil {
					for _, f := range rel.Model.PrimaryFields {
						rel.ForeignFieldNames = append(rel.ForeignFieldNames, f.Name)
						rel.ForeignDBNames = append(rel.AssociationDBNames, f.DBName)
					}
				}
				if rel.AssociationFieldNames == nil {
					for _, f := range rel.AssociationModel.PrimaryFields {
						rel.AssociationFieldNames = append(rel.AssociationFieldNames, f.Name)
						rel.AssociationDBNames = append(rel.AssociationDBNames, f.DBName)
					}
				}
				res.SetParentResource(config.Sub.Parent, config.Sub.Relation)
			}

			if !res.Singleton {
				subResourceConfigureFilters(res)
			}
		}
	}

	for _, cb := range this.BeforeResourceInitializeCallbacks {
		cb(res)
	}

	// Configure resource when initializing
	modelType := utils.ModelType(res.Value)
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceBeforeInitializeInterface); ok {
				injector.ConfigureResourceBeforeInitialize(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceBeforeInitializeInterface); ok {
		injector.ConfigureResourceBeforeInitialize(res)
	}

	if !config.Alone {
		res.OnDBActionE(func(e *resource.DBEvent) (err error) {
			if e.Context.ResourceID == nil {
				if idS := e.Context.URLParam(res.ParamIDName()); idS != "" {
					e.Context.ResourceID, err = res.ParseID(idS)
				}
			}
			return
		}, resource.E_DB_ACTION_FIND_ONE.Before())

		res.UseTheme("slideout")

		if res.ParentRelation != nil {
			if v, ok := res.ParentRelation.Data.Get("hidden_fields_disabled"); !ok || !v.(bool) {
				if res.ParentRelation.FieldName != "" {
					res.MetaDisable(res.ParentRelation.FieldName)
				}
				var ffNames []string
				for _, ff := range res.ParentRelation.ForeignFields() {
					if !ff.IsPrimaryKey {
						ffNames = append(ffNames, ff.Name)
					}
				}
				res.MetaDisable(ffNames...)
			}
		}

		parts := strings.Split(strings.ReplaceAll(res.UID, ".models.", "."), "@")
		for i, p := range parts {
			pos := strings.LastIndexByte(p, '.')
			parts[i] = p[0:pos] + "/" + p[pos+1:]
		}

		if res.ParentResource != nil {
			res.TemplatePath = res.ParentResource.TemplatePath + "/sub/" + utils.ToParamString(res.ID)
		} else {
			res.TemplatePath = utils.ToUri(strings.ReplaceAll(strings.Join(parts, "/sub/"), "#", "/"))
		}

		configureDefaultLayouts(res)
	}
	return res
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (this *Admin) NewResource(value interface{}, config ...*Config) *Resource {
	if len(config) == 0 {
		config = []*Config{nil}
	}
	return this.NewResourceConfig(value, config[0])
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (this *Admin) NewSingletonResource(value interface{}, config ...*Config) *Resource {
	if len(config) == 0 {
		config = []*Config{{}}
	}
	config[0].Singleton = true
	return this.NewResourceConfig(value, config[0])
}

// NewResourceConfig initialize a new qor resource, won't add it to admin, just initialize it
func (this *Admin) NewResourceConfig(value interface{}, cfg *Config) (res *Resource) {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.Alone = true

	if res = this.newResource(value, cfg, nil); res == nil {
		return
	}

	res.configure()

	res.PostInitialize(func() {
		for _, layout := range res.Layouts {
			if l, ok := layout.(*Layout); ok {
				l.Resource = res
				l.SetMetaNames(l.MetaNames)
			}
		}
	})

	if setuper, ok := res.Value.(ResourceSetuper); ok {
		setuper.AdminResourceSetup(res, func() {
			if res.Config.Setup != nil {
				res.Config.Setup(res)
			}
			for _, setup := range res.Config.Setups {
				if setup != nil {
					setup(res)
				}
			}
		})
	} else {
		if res.Config.Setup != nil {
			res.Config.Setup(res)
		}
		for _, setup := range res.Config.Setups {
			if setup != nil {
				setup(res)
			}
		}
	}

	res.initializeLayouts()
	res.initialized = true

	for _, cb := range res.postInitializeCallbacks {
		cb()
	}

	for _, cb := range this.AfterResourceInitializeCallbacks {
		cb(res)
	}

	res.postInitializeCallbacks = nil

	err := this.TriggerResource(&ResourceEvent{edis.NewEvent(E_RESOURCE_ADDED), res, nil, false})
	if err != nil {
		panic(errwrap.Wrap(err, "Trigger Resource Added"))
	}
	return
}

// AddResource make a model manageable from admin interface
func (this *Admin) AddResource(value interface{}, config ...*Config) (res *Resource) {
	var cfg *Config
	for _, cfg = range config {
	}
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Duplicated == nil {
		cfg.Duplicated = func(uid string, res *Resource) {
			panic("Duplicate resource: UID=" + uid)
		}
	}

	var (
		log       logging.Logger
		donea     func()
		callbacks []func(res *Resource)
	)

	defer func() {
		if donea != nil {
			donea()
		}
	}()

	for _, cb := range this.BeforeAddResourceCallbacks {
		cb(value, cfg, func(f func(res *Resource)) {
			callbacks = append(callbacks, f)
		})
	}

	res = this.newResource(value, cfg, func(uid string) {
		log = logging.WithPrefix(newResourceLog, uid)
		donea = func() { log.Debug("done") }
	})

	res.SetDefaultDenyMode(func() bool {
		return this.DefaultDenyMode
	})

	if _, ok := this.ResourcesByUID[res.UID]; ok {
		return res
	}

	if !cfg.Protected {
		this.ResourcesByUID[res.UID] = res
	}

	res.configure()

	if res.ParentResource != nil {
		res.ParentResource.Resources[res.ID] = res
		res.ParentResource.ResourcesByParam[res.Param] = res
		if !res.Config.Invisible {
			if res.IsSingleton() {
				if res.ParentResource.IsSingleton() {
					menu := res.ParentResource.AddMenu(res.DefaultMenu())
					menu.EnabledFunc = func(menu *Menu, context *Context) bool {
						if !context.NotFound {
							if res.Config.MenuEnabled != nil {
								return res.Config.MenuEnabled(menu, context)
							}
							return true
						}
						return false
					}
				} else {
					menu := res.ParentResource.AddItemMenu(res.DefaultMenu())
					menu.EnabledFunc = res.Config.MenuEnabled
				}
			} else {
				menu := res.ParentResource.AddItemMenu(res.DefaultMenu())
				menu.EnabledFunc = func(menu *Menu, context *Context) bool {
					if !context.NotFound {
						if !context.IsResultSlice() {
							if !aorm.IdOf(context.Result).IsZero() {
								if res.Config.MenuEnabled != nil {
									return res.Config.MenuEnabled(menu, context)
								}
								return true
							}
						}
					}
					return false
				}
			}
		}
	} else {
		this.Resources[res.ID] = res
		this.ResourcesByParam[res.Param] = res
		if !res.Config.Invisible {
			this.AddMenu(res.DefaultMenu())
		}
	}

	if !res.Config.NotMount {
		res.RegisterDefaultRouters()
		res.mounted = true

		for _, am := range res.postMountCallbacks {
			am()
		}

		res.postMountCallbacks = nil
	}

	res.initializeLayouts()

	if resources, ok := this.ResourcesByType[res.ModelStruct.Type]; ok {
		this.ResourcesByType[res.ModelStruct.Type] = append(resources, res)
	} else {
		this.ResourcesByType[res.ModelStruct.Type] = []*Resource{res}
	}

	done := func() {
		if setuper, ok := res.Value.(ResourceSetuper); ok {
			log.Debug("setup start ", reflect.TypeOf(res.Config.Setup).PkgPath())
			setuper.AdminResourceSetup(res, func() {
				if res.Config.Setup != nil {
					res.Config.Setup(res)
				}
			})
			log.Debug("setup done ", reflect.TypeOf(res.Config.Setup).PkgPath())
		} else if res.Config.Setup != nil {
			log.Debug("setup start ", reflect.TypeOf(res.Config.Setup).PkgPath())
			res.Config.Setup(res)
			log.Debug("setup done ", reflect.TypeOf(res.Config.Setup).PkgPath())
		}

		for _, f := range res.ModelStruct.Fields {
			if tags := ParseMetaTags(f.Tag); tags.Managed() != nil {
				if res.ParentResource != nil && res.ParentResource.ModelStruct == res.ModelStruct {
					res.Meta(&Meta{Name: f.Name, Disabled: true})
				} else {
					res.Meta(&Meta{Name: f.Name})
				}
			}
		}

		res.initialized = true

		for _, cb := range res.postInitializeCallbacks {
			log.Debug("after register start")
			cb()
			log.Debug("after register done")
		}

		res.postInitializeCallbacks = nil

		for _, cb := range this.AfterResourceInitializeCallbacks {
			cb(res)
		}

		log.Debug("trigger added start")
		if err := this.triggerResourceAdded(res); err != nil {
			panic(err)
		}
		log.Debug("trigger added done")

		if len(cfg.CreateWizards) > 0 {
			for _, wz := range cfg.CreateWizards {
				res.AddCreateWizard(wz.Value, wz.Config)
			}
		}
	}

	if res.ParentResource != nil && !res.ParentResource.initialized {
		res.ParentResource.PostInitialize(done)
	} else {
		done()
	}

	for _, cb := range callbacks {
		cb(res)
	}

	return res
}

// GetResources get defined resources from admin
func (this *Admin) GetResources() (resources []*Resource) {
	for _, r := range this.Resources {
		resources = append(resources, r)
	}
	return
}

func (this *Admin) WalkResources(f func(res *Resource) bool) bool {
	for _, r := range this.Resources {
		if !f(r) {
			break
		}
		if !r.WalkResources(f) {
			break
		}
	}
	return true
}

// GetResourceByID get resource with name
func (this *Admin) GetResourceByID(id string) (resource *Resource) {
	parts := strings.SplitN(id, ".", 2)
	r := this.Resources[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByID(parts[1])
	}
}

// GetResourceByID get resource with name
func (this *Admin) GetResourceByParam(param string) (resource *Resource) {
	parts := strings.SplitN(param, ".", 2)
	r := this.ResourcesByParam[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByParam(parts[1])
	}
}

func (this *Admin) GetParentResourceByID(id string) *Resource {
	return this.GetResourceByID(id)
}

func (this *Admin) GetOrParentResourceByID(id string) *Resource {
	return this.GetParentResourceByID(id)
}

// AddSearchResource make a resource searchable from search center
func (this *Admin) AddSearchResource(resources ...*Resource) {
	for _, res := range resources {
		if _, ok := res.ControllerBuilder.Controller.(ControllerSearcher); ok {
			this.searchResources = append(this.searchResources, res)
		}
	}
}

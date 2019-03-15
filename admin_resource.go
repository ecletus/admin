package admin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/moisespsena-go/aorm"

	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/aghape/fragment"
	"github.com/jinzhu/inflection"
	"github.com/moisespsena-go/xroute"
	"github.com/moisespsena/go-edis"
	"github.com/moisespsena/go-error-wrap"
)

func (admin *Admin) newResource(value interface{}, config *Config, onUid func(uid string)) *Resource {
	if config == nil {
		config = &Config{}
	}

	if value == nil {
		if config.Sub.Parent == nil {
			panic("Resource Value is nil.")
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

			typ := field.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if typ.Kind() == reflect.Slice {
				if config.Param == "" {
					config.Param = config.Sub.FieldName
				}
				typ = typ.Elem()
				if typ.Kind() == reflect.Ptr {
					typ = typ.Elem()
				}
			}
			value = reflect.New(typ).Interface()
		} else {
			panic("Resource field \"" + config.Sub.FieldName + "\" does not exists.")
		}
	}

	var uid string
	if config.Sub != nil && config.Sub.Parent != nil {
		uid = config.Sub.Parent.UID + "@"
		if config.Name == "" && config.Param == "" && config.ID == "" && config.Sub.FieldName != "" {
			config.ID = config.Sub.FieldName
		}
	}

	uid += utils.TypeId(value)

	if onUid != nil {
		onUid(uid)
	}

	if config.Controller == nil {
		config.Controller = NewCrudSearchIndexController()
	}

	res := &Resource{
		Resource:         resource.New(admin.FakeDB.NewScope(value), config.ID, uid),
		Config:           config,
		cachedMetas:      &map[string][]*Meta{},
		admin:            admin,
		Resources:        make(map[string]*Resource),
		ResourcesByParam: make(map[string]*Resource),
		MetaAliases:      make(map[string]*resource.MetaName),
		MetasByName:      make(map[string]*Meta),
		MetasByFieldName: make(map[string]*Meta),
		Inherits:         make(map[string]*Child),
		RouteHandlers:    make(map[string]*RouteHandler),
		labelKey:         config.LabelKey,
	}

	res.Controller.Resource = res

	_, res.softDelete = value.(aorm.SoftDeleter)

	res.Controller.Controller = config.Controller
	res.Controller.ViewController = &ResourceViewController{
		Controller: &Controller{Admin: admin, controller: config.Controller},
	}

	res.Scheme = NewScheme(res, "Default")
	res.Resource.SetDispatcher(res)

	if _, ok := value.(fragment.FragmentedModelInterface); ok {
		res.Fragments = NewFragments()
	}

	res.Children = &Inheritances{resource: res}

	if config.ID != "" {
		res.ID = config.ID
		res.I18nPrefix += "." + res.ID
	}

	if config.Sub != nil {
		if config.Sub.Parent == nil {
			panic("Parent is nil.")
		}
		res.ParentResource = config.Sub.Parent
	}

	res.Router = xroute.NewMux(res.ID)

	if config.Prefix != "" {
		res.Router.SetPrefix(strings.Replace(config.Prefix, ".", "/", -1))
	}

	if res.Config.Singleton {
		res.ObjectRouter = res.Router
	} else {
		res.ObjectRouter = xroute.NewMux(res.ID + ":ObjectRouter")
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

	res.Param = config.Param
	if res.Param == "" {
		if res.Config.Singleton {
			res.Param = res.Name
		} else {
			res.Param = res.PluralName
		}
		if config.Prefix != "" {
			res.Param = config.Prefix + "." + res.Param
		}
	}

	res.Param = utils.ToParamString(res.Param)
	res.Parents = resourceParents(res)
	res.PathLevel = len(res.Parents)
	res.ParamName = resourceParamName(res.Parents, res.Param)
	res.paramIDName = resourceParamIDName(res.PathLevel, res.ParamName)

	if config.Sub != nil {
		if config.Sub.FieldName != "" {
			if field, ok := config.Sub.Parent.FakeScope.FieldByName(config.Sub.FieldName); ok {
				if field.Relationship != nil {
					res.SetParentResource(config.Sub.Parent, field.Relationship.ForeignFieldNames[0])
				} else {
					res.SetParentResource(config.Sub.Parent, "")
				}
				//res.SetPrimaryFields(field.Relationship.ForeignFieldNames...)
			} else {
				panic(fmt.Sprintf("Invalid fieldName %q", config.Sub.FieldName))
			}
		} else if config.Sub.ParentFieldName != "" {
			res.SetParentResource(config.Sub.Parent, config.Sub.ParentFieldName)
		}

		if res.IsParentFieldVirtual() && !config.Invisible {
			panic("Sub resource does not have relation for parent.")
		}

		subResourceConfigureFilters(res)
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

	res.OnDBAction(func(e *resource.DBEvent) {
		if e.Context.ResourceID == "" {
			e.Context.ResourceID = e.Context.URLParam(res.ParamIDName())
		}
	}, resource.E_DB_ACTION_FIND_ONE.Before())

	res.UseTheme("slideout")
	configureDefaultLayouts(res)
	return res
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (admin *Admin) NewResource(value interface{}, config ...*Config) *Resource {
	if len(config) == 0 {
		config = []*Config{nil}
	}
	return admin.NewResourceConfig(value, config[0])
}

// NewResourceConfig initialize a new qor resource, won't add it to admin, just initialize it
func (admin *Admin) NewResourceConfig(value interface{}, cfg *Config) *Resource {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.Invisible = true
	cfg.NotMount = true
	res := admin.newResource(value, cfg, nil)
	res.configure()

	res.AfterRegister(func() {
		for _, layout := range res.Layouts {
			if l, ok := layout.(*Layout); ok {
				l.Resource = res
				l.SetMetaNames(l.MetaNames)
			}
		}
	})

	if res.Config.Setup != nil {
		res.Config.Setup(res)
	}
	res.initializeLayouts()
	for _, cb := range res.afterRegister {
		cb()
	}
	res.afterRegister = nil
	res.registered = true
	err := admin.TriggerResource(&ResourceEvent{edis.NewEvent(E_RESOURCE_ADDED), res, false})
	if err != nil {
		panic(errwrap.Wrap(err, "Trigger Resource Added"))
	}
	return res
}

// AddResource make a model manageable from admin interface
func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	if len(config) == 0 {
		config = []*Config{nil}
	}

	res := admin.newResource(value, config[0], func(uid string) {
		if _, ok := admin.ResourcesByUID[uid]; ok {
			panic("Duplicate resource: UID=" + uid)
		}
	})

	admin.ResourcesByUID[res.UID] = res
	res.configure()

	if res.ParentResource != nil {
		res.ParentResource.Resources[res.ID] = res
		res.ParentResource.ResourcesByParam[res.Param] = res
		if !res.Config.Invisible {
			menu := res.ParentResource.AddMenu(res.DefaultMenu())
			menu.Enabled = func(menu *Menu, context *Context) bool {
				if !context.NotFound && res.GetKey(context.Result) != "" {
					if res.Config.MenuEnabled != nil {
						return res.Config.MenuEnabled(menu, context)
					}
					return true
				}
				return false
			}
		}
	} else {
		admin.Resources[res.ID] = res
		admin.ResourcesByParam[res.Param] = res
		if !res.Config.Invisible {
			admin.AddMenu(res.DefaultMenu())
		}
	}

	if !res.Config.NotMount {
		res.RegisterDefaultRouters()
		res.mounted = true

		for _, am := range res.afterMount {
			am()
		}

		res.afterMount = nil
	}

	res.initializeLayouts()

	if res.Config.Setup != nil {
		res.Config.Setup(res)
	}

	res.registered = true
	for _, cb := range res.afterRegister {
		cb()
	}
	res.afterRegister = nil

	admin.triggerResourceAdded(res)

	return res
}

func (admin *Admin) triggerResourceAdded(res *Resource, cb ...func(e *ResourceEvent)) {
	e := &ResourceEvent{edis.NewEvent(E_RESOURCE_ADDED), res, true}
	if len(cb) > 0 {
		for _, cb := range cb {
			cb(e)
		}
		return
	}
	err := admin.TriggerResource(e)
	if err != nil {
		panic(errwrap.Wrap(err, "Trigger Resource Added"))
	}
}

// GetResources get defined resources from admin
func (admin *Admin) GetResources() (resources []*Resource) {
	for _, r := range admin.Resources {
		resources = append(resources, r)
	}
	return
}

func (admin *Admin) WalkResources(f func(res *Resource) bool) bool {
	for _, r := range admin.Resources {
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
func (admin *Admin) GetResourceByID(id string) (resource *Resource) {
	if resource, ok := admin.ResourcesByUID[id]; ok {
		return resource
	}
	parts := strings.SplitN(id, ".", 2)
	r := admin.Resources[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByID(parts[1])
	}
}

// GetResourceByID get resource with name
func (admin *Admin) GetResourceByParam(param string) (resource *Resource) {
	parts := strings.SplitN(param, ".", 2)
	r := admin.ResourcesByParam[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByParam(parts[1])
	}
}

func (admin *Admin) GetParentResourceByID(id string) *Resource {
	return admin.GetResourceByID(id)
}

func (admin *Admin) GetOrParentResourceByID(id string) *Resource {
	return admin.GetParentResourceByID(id)
}

// AddSearchResource make a resource searchable from search center
func (admin *Admin) AddSearchResource(resources ...*Resource) {
	for _, res := range resources {
		if _, ok := res.Controller.Controller.(ControllerSearcher); ok {
			admin.searchResources = append(admin.searchResources, res)
		}
	}
}

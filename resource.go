package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/moisespsena/go-edis"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
	"github.com/jinzhu/inflection"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/xroute"

	//"github.com/ecletus/responder"
	"strconv"

	"github.com/ecletus/db/inheritance"
	"github.com/moisespsena/template/html/template"
)

const (
	DEFAULT_LAYOUT = resource.DEFAULT_LAYOUT

	// paths
	P_NEW_FORM              = "/new"
	P_NEW                   = "/"
	P_OBJ_READ              = "/"
	P_OBJ_READ_FORM         = "/"
	P_OBJ_UPDATE            = "/"
	P_OBJ_UPDATE_FORM       = "/edit"
	P_OBJ_DELETE            = "/"
	P_SINGLETON_READ        = P_OBJ_READ
	P_SINGLETON_UPDATE      = P_OBJ_UPDATE
	P_SINGLETON_UPDATE_FORM = P_OBJ_UPDATE_FORM
	P_BULK_DELETE           = "/delete"
	P_RESTORE               = "/" + A_RESTORE
	P_DELETED_INDEX         = "/" + A_DELETED_INDEX
	P_INDEX                 = "/"
	P_SEARCH                = "/search"

	// actions
	A_CREATE      = "create"
	A_UPDATE      = "update"
	A_READ        = "read"
	A_DELETE      = "delete"
	A_BULK_DELETE = "bulk_delete"
	A_INDEX       = "index"
	A_SEARCH      = "search"

	A_RESTORE       = "restore"
	A_DELETED_INDEX = "deleted_index"

	META_STRING = "String"

	ActionDelete     = "Delete"
	ActionBulkDelete = "BulkDelete"
)

// Resource is the most important thing for qor admin, every model is defined as a resource, qor admin will genetate management interface based on its definition
type Resource struct {
	*resource.Resource
	*Scheme

	Paged

	ObjectPages Paged

	ParentResource   *Resource
	Config           *Config
	Metas            []*Meta
	MetasByName      map[string]*Meta
	MetasByFieldName map[string]*Meta
	SingleEditMetas  map[string]*Meta
	Actions          []*Action

	admin       *Admin
	mounted     bool
	cachedMetas *map[string][]*Meta

	Controller ResourceController

	Router       *xroute.Mux
	ObjectRouter *xroute.Mux
	Parents      []*Resource
	Param        string
	ParamName    string
	paramIDName  string

	Resources          map[string]*Resource
	ResourcesByParam   map[string]*Resource
	MetaAliases        map[string]*resource.MetaName
	defaultDisplayName string
	Children           *Inheritances
	Inherits           map[string]*Child
	Fragments          *Fragments
	Fragment           *Fragment
	registered         bool
	afterRegister      []func()
	afterMount         []func()
	RouteHandlers      map[string]*RouteHandler

	labelKey   string
	softDelete bool
}

func (res *Resource) IsSoftDelete() bool {
	return res.softDelete
}

func (res *Resource) Top() (top *Resource) {
	top = res
	for top.ParentResource != nil {
		top = top.ParentResource
	}
	return
}

func (res *Resource) OnDBActionE(cb func(e *resource.DBEvent) error, action ...resource.DBActionEvent) (err error) {
	return resource.OnDBActionE(res, cb, action...)
}

func (res *Resource) OnDBAction(cb func(e *resource.DBEvent), action ...resource.DBActionEvent) (err error) {
	return resource.OnDBAction(res, cb, action...)
}

// GetDBKeys Returns the DB Keys values from `value`.
func (res *Resource) GetDBKeys(value interface{}) (keys []string) {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	for _, field := range res.FakeScope.PrimaryFields() {
		keys = append(keys, fmt.Sprint(reflectValue.FieldByIndex(field.Struct.Index).Interface()))
	}
	return
}

func (res *Resource) GetDBKey(value interface{}) string {
	return strings.Join(res.GetDBKeys(value), ",")
}

// GetKeys Returns the Resource Keys values from `value`.
func (res *Resource) GetKeys(value interface{}) []string {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	key := aorm.Key()

	for _, field := range res.PrimaryFields {
		rf := reflectValue.FieldByName(field.Struct.Name).Interface()
		key.Append(rf)
	}

	return key.Strings()
}

func (res *Resource) GetKey(value interface{}) (key string) {
	switch vt := value.(type) {
	case interface{ GetID() string }:
		return vt.GetID()
	case interface{ GetID() int64 }:
		if id := vt.GetID(); id > 0 {
			return strconv.Itoa(int(id))
		}
		return ""
	default:
		return strings.Join(res.GetKeys(value), ",")
	}
}

func (res *Resource) LabelKey() string {
	if res.labelKey != "" {
		return res.labelKey
	}

	return res.I18nPrefix + ".label"
}

func (res *Resource) GetLabelKey(plural bool) string {
	r := res.LabelKey() + "~"
	if plural {
		r += "p"
	} else {
		r += "s"
	}
	return r
}

func (res *Resource) PluralLabelKey() string {
	return res.GetLabelKey(true)
}

func (res *Resource) SingularLabelKey() string {
	return res.GetLabelKey(false)
}

func (res *Resource) GetDefaultLabel(plural bool) string {
	if plural {
		return inflection.Plural(res.Name)
	} else {
		return res.Name
	}
}

func (res *Resource) GetLabel(context *Context, plural bool) string {
	return string(context.t(res.GetLabelKey(plural), res.GetDefaultLabel(plural)))
}

func (res *Resource) GetActionLabelKey(action *Action) string {
	return fmt.Sprintf("resources.%v.actions.%v", res.ToParam(), action.Label)
}

func (res *Resource) GetActionLabel(context *Context, action *Action) template.HTML {
	return context.t(res.GetActionLabelKey(action), action.Label)
}

// GetAdmin get admin from resource
func (res *Resource) GetAdmin() *Admin {
	return res.admin
}

// GetURL
func (res *Resource) GetParentFaultURI(fault func(res *Resource) string, parentkeys ...string) string {
	var p []string
	l := len(parentkeys)
	for i, key := range parentkeys {
		pres := res.Parents[l-i-1]
		p = append(p, pres.ToParam())
		if !pres.Config.Singleton {
			if key == "" {
				key = fault(pres)
			}
			p = append(p, key)
		}
	}
	if len(p) > 0 {
		return "/" + strings.Join(p, "/")
	}
	return ""
}

// GetURL
func (res *Resource) GetParentURI(parentkeys ...string) string {
	return res.GetParentFaultURI(func(res *Resource) string {
		return "{" + res.ParamIDPattern() + "}"
	}, parentkeys...)
}

// GetURL
func (res *Resource) GetIndexURI(parentkeys ...string) string {
	return res.GetParentURI(parentkeys...) + "/" + res.ToParam()
}

// GetURL
func (res *Resource) GetURI(key string, parentkeys ...string) string {
	return res.GetIndexURI(parentkeys...) + "/" + key
}

// GetURL
func (res *Resource) URLFor(recorde interface{}, parentkeys ...string) string {
	return res.GetIndexURI(parentkeys...) + "/" + res.GetKey(recorde)
}

// GetURL
func (res *Resource) GetContextIndexURI(context *core.Context, parentkeys ...string) string {
	if len(parentkeys) == 0 && res.ParentResource != nil {
		parentkeys = context.ParentResourceID
		if len(parentkeys) == 0 {
			parentkeys = make([]string, len(res.Parents))
		}
	}
	return context.GenURL(res.GetParentFaultURI(func(res *Resource) string {
		return context.URLParam(res.ParamIDPattern())
	}, parentkeys...), res.ToParam())
}

// GetURL
func (res *Resource) GetContextURI(context *core.Context, key string, parentkeys ...string) string {
	if key == "" {
		key = context.URLParam(res.ParamIDName())
	}
	return res.GetContextIndexURI(context, parentkeys...) + "/" + key
}

func (res *Resource) GetRecordURI(record interface{}, parentKeys ...string) string {
	if ref, ok := record.(inheritance.ParentModelInterface); ok {
		if child := ref.GetQorChild(); child != nil {
			res = res.Children.Items[child.Index].Resource
			// TODO Fix support for subresources
			return res.GetURI(child.ID)
		}
	}
	uri := res.GetURI(res.GetKey(record), parentKeys...)
	return uri
}

// GetPrimaryValue get priamry value from request
func (res *Resource) GetPrimaryValue(params utils.ReadonlyMapString) string {
	if params != nil {
		return params.Get(res.ParamIDName())
	}
	return ""
}

// ParamIDName return param name for primary key like :product_id
func (res *Resource) ParamIDName() string {
	return res.paramIDName
}

// ParamIDName return param name for primary key like :product_id
func (res *Resource) ParamIDPattern() string {
	return "{" + res.paramIDName + "}"
}

// ToParam used as urls to register routes for resource
func (res *Resource) ToParam() string {
	return res.Param
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (res *Resource) NewResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = res
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}
	return res.admin.NewResource(value, config[0])
}

// AddSubResource register sub-resource
func (res *Resource) AddResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = res
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}
	return res.AddResourceConfig(value, config[0])
}

// AddSubResource register sub-resource
func (res *Resource) AddResourceConfig(value interface{}, cfg *Config) *Resource {
	if cfg.Sub == nil {
		cfg.Sub = &SubConfig{}
	}
	cfg.Sub.Parent = res
	return res.admin.AddResource(value, cfg)
}

func (res *Resource) AddResourceFieldConfig(fieldName string, value interface{}, cfg *Config) *Resource {
	if value == nil {
		field, _ := utils.IndirectType(res.Value).FieldByName(fieldName)
		fieldType := utils.IndirectType(field.Type)

		if fieldType.Kind() == reflect.Slice {
			fieldType = utils.IndirectType(fieldType.Elem())
		}

		value = reflect.New(fieldType).Interface()
	}
	setup := cfg.Setup
	cfg.Setup = func(child *Resource) {
		res.SetMeta(&Meta{Name: fieldName, Resource: child})
		if setup != nil {
			setup(child)
		}
	}
	return res.AddResource(&SubConfig{FieldName: fieldName}, value, cfg)
}

func (res *Resource) AddResourceField(fieldName string, value interface{}, setup ...func(res *Resource)) *Resource {
	return res.AddResourceFieldConfig(fieldName, value, &Config{
		Setup: func(res *Resource) {
			for _, s := range setup {
				s(res)
			}
		},
	})
}

func (res *Resource) DBName(quote bool) (name string, alias string, pkfields []string) {
	if quote {
		name = res.FakeScope.QuotedTableName()
	} else {
		name = res.FakeScope.TableName()
	}

	alias = "parent_" + strconv.Itoa(res.PathLevel)
	for _, f := range res.PrimaryFields {
		pkfields = append(pkfields, f.DBName)
	}

	return
}

func (res *Resource) FilterByParent(db *aorm.DB, parentKey string) *aorm.DB {
	r := res.ParentResource

	res_parent_alias := res.FakeScope.QuotedTableName()
	res_pkfield := res.ParentFieldDBName

	var (
		parent_name, parent_alias string
		pkfield                   []string
	)

	parent_name, parent_alias, pkfield = r.DBName(true)
	parent_alias += "_"
	var fields, wheres []string
	ids := strings.Split(parentKey, ",")

	fields = append(fields, fmt.Sprintf("%[1]v.%[2]v = %[3]v.%[4]v", parent_alias,
		pkfield[0], res_parent_alias, res_pkfield))
	wheres = append(wheres, parent_alias+"."+pkfield[0]+" = ?")

	join := fmt.Sprintf("JOIN %v as %v ON %v", parent_name, parent_alias,
		strings.Join(fields, " AND "))

	idsinterface := make([]interface{}, len(ids), len(ids))
	for i, id := range ids {
		idsinterface[i] = id
	}

	db = db.Joins(join).Where(strings.Join(wheres, " AND "), idsinterface...)

	r = r.ParentResource
	return db
}

func (res *Resource) GetPathLevel() int {
	return res.PathLevel
}

// Decode decode context into a value
func (res *Resource) Decode(context *core.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func DefaultPermission(action string, defaul ...roles.PermissionMode) roles.PermissionMode {
	switch action {
	case "index", "show":
		return roles.Read
	case "edit":
		return roles.Update
	}
	if len(defaul) == 0 {
		return roles.NONE
	}
	return defaul[0]
}

func (res *Resource) BasicValue(ctx *core.Context, recorde interface{}) resource.BasicValue {
	metaLabel, metaIcon := res.MetasByName[BASIC_META_LABEL], res.MetasByName[BASIC_META_ICON]
	id, label, icon := res.GetKey(recorde),
		metaLabel.FormattedValue(ctx, recorde).(string),
		metaIcon.FormattedValue(ctx, recorde).(string)
	return &resource.Basic{id, label, icon}
}

func (res *Resource) DeleteAction() *Action {
	if res.Config.Singleton || !res.Controller.IsDeleter() {
		return nil
	}
	return res.Action(&Action{Name: ActionDelete})
}

func (res *Resource) configure() {
	modelType := utils.ModelType(res.Value)

	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceInterface); ok {
				injector.ConfigureQorResource(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceInterface); ok {
		injector.ConfigureQorResource(res)
	}

	if !res.Config.Singleton && res.Controller.IsDeleter() {
		res.Action(&Action{
			Name:   ActionDelete,
			Method: http.MethodDelete,
			Type:   ActionDanger,
			URL: func(record interface{}, context *Context, args ...interface{}) string {
				return res.GetContextURI(context.Context, res.GetKey(record))
			},
			Modes: []string{"menu_item"},
			Visible: func(recorde interface{}, context *Context) bool {
				if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
					return false
				}
				return !res.IsSoftDeleted(recorde)
			},
			RefreshURL: func(record interface{}, context *Context) string {
				return res.GetContextIndexURI(context.Context)
			},
		})

		if res.Controller.IsBulkDeleter() {
			res.Action(&Action{
				Name:   ActionBulkDelete,
				Method: http.MethodPost,
				Type:   ActionDanger,
				URL: func(record interface{}, context *Context, args ...interface{}) string {
					return res.GetContextIndexURI(context.Context) + P_BULK_DELETE
				},
				Modes: []string{"index"},
				Visible: func(recorde interface{}, context *Context) bool {
					if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
						return false
					}
					return !res.IsSoftDeleted(recorde)
				},
				IndexVisible: func(context *Context) bool {
					if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
						return false
					}
					return true
				},
			})
		}

		if res.softDelete && res.Controller.IsRestorer() {
			res.AfterRegister(res.configureRestorer)
		}
	}

	res.AfterRegister(res.configureAudited)
}
func (res *Resource) configureRestorer() {
	/*
		res.AddDefaultMenuChild(&Menu{
			Name: A_DELETED_INDEX,
			MakeLink: func(context *Context, args ...interface{}) string {
				return res.GetContextIndexURI(context.Context) + "/" + A_DELETED_INDEX
			},
		})
	*/
	res.Action(&Action{
		Name:   A_RESTORE,
		Modes:  []string{"index"},
		Method: http.MethodGet,
		URL: func(record interface{}, context *Context, args ...interface{}) (url string) {
			url = res.GetContextIndexURI(context.Context) + "/" + A_RESTORE
			if reflect.Indirect(reflect.ValueOf(record)).Kind() == reflect.Slice {
				return
			}
			return url + "?key=" + res.GetKey(record)
		},
		IndexVisible: func(context *Context) bool {
			if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
				return true
			}
			return false
		},
		Visible: func(record interface{}, context *Context) bool {
			return false
		},
		Handler: func(argument *ActionArgument) error {
			println()
			return nil
		},
	})
}

func (res *Resource) Restore(ctx *Context, key ...string) {
	DB := ctx.DB.Model(res.Value)
	var where []string
	var args []interface{}

	for _, key := range key {
		pwhere, pargs := resource.StringToPrimaryQuery(res, key)
		where = append(where, pwhere)
		args = append(args, pargs...)
	}

	data := map[string]interface{}{"deleted_at": nil}
	if f, ok := reflect.TypeOf(res.Value).Elem().FieldByName("DeletedByID"); ok {
		switch f.Type.Kind() {
		case reflect.Ptr:
			data["deleted_by_id"] = nil
		case reflect.String:
			data["deleted_by_id"] = ""
		default:
			data["deleted_by_id"] = 0
		}
	}

	DB = DB.Table(res.FakeScope.TableName()).
		Unscoped().
		Where(strings.Join(where, " OR "), args...).
		Set("validations:skip_validations", true)

	err := DB.Updates(data).Error
	ctx.AddError(err)
}

func (res *Resource) IsSoftDeleted(recorde interface{}) bool {
	if res.softDelete {
		typ := reflect.ValueOf(recorde)

		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		f := typ.FieldByName(aorm.SoftDeleteFieldDeletedAt)
		if f.IsValid() {
			v := f.Interface()
			if t, ok := v.(time.Time); ok {
				return !t.IsZero()
			} else if t, ok := v.(*time.Time); ok {
				return t != nil && !t.IsZero()
			}
		}
	}
	return false
}

// GetResources get defined resources from admin
func (res *Resource) GetResources() (resources []*Resource) {
	for _, r := range res.Resources {
		resources = append(resources, r)
	}
	return
}

func (res *Resource) WalkResources(f func(res *Resource) bool) bool {
	for _, r := range res.Resources {
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
func (res *Resource) GetResourceByID(id string) (resource *Resource) {
	parts := strings.SplitN(id, ".", 2)
	r := res.Resources[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByID(parts[1])
	}
}

// GetResourceByParam get resource with name
func (res *Resource) GetResourceByParam(param string) (resource *Resource) {
	parts := strings.SplitN(param, ".", 2)
	r := res.ResourcesByParam[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByParam(parts[1])
	}
}

func (res *Resource) GetParentResourceByID(id string) *Resource {
	for _, p := range res.Parents {
		if p.ID == id {
			return p
		}
	}
	return res.admin.GetResourceByID(id)
}

func (res *Resource) GetOrParentResourceByID(id string) *Resource {
	r := res.GetResourceByID(id)
	if r == nil {
		r = res.GetParentResourceByID(id)
	}
	return r
}

func (res *Resource) SubResources() (items []*Resource) {
	for _, r := range res.Resources {
		if !r.Config.Invisible {
			items = append(items, r)
		}
	}
	return
}

func (res *Resource) ReferencedRecord(record interface{}) interface{} {
	return nil
}

func (res *Resource) CrudScheme(ctx *core.Context, scheme interface{}) *resource.CRUD {
	s := res.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = res.GetSchemeOk(st)
	default:
		if scheme != nil {
			s = scheme.(*Scheme)
		}
	}
	return res.Crud(ctx).Dispatcher(s.EventDispatcher)
}

func (res *Resource) CrudSchemeDB(db *aorm.DB, scheme interface{}) *resource.CRUD {
	s := res.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = res.GetSchemeOk(st)
	default:
		if scheme != nil {
			s = scheme.(*Scheme)
		}
	}
	return res.CrudDB(db).Dispatcher(s.EventDispatcher)
}

func (res *Resource) Crud(ctx *core.Context) *resource.CRUD {
	return resource.NewCrud(res, ctx)
}

func (res *Resource) CrudDB(db *aorm.DB) *resource.CRUD {
	return res.Crud(&core.Context{DB: db})
}

func (res *Resource) SetParentResource(parent *Resource, fieldName string) {
	res.Resource.SetParent(parent, fieldName)
	res.ParentResource = parent
}

func (res *Resource) RegisterScheme(name string, cfg ...*SchemeConfig) *Scheme {
	f := func() *Scheme {
		return res.Scheme.AddChild(name, cfg...)
	}
	if res.registered {
		return f()
	}
	res.AfterRegister(func() {
		f()
	})
	return nil
}

func (res *Resource) AfterRegister(f ...func()) {
	res.afterRegister = append(res.afterRegister, f...)
}

func (res *Resource) AfterMount(f ...func()) {
	if res.mounted {
		for _, f := range f {
			f()
		}
	} else {
		res.afterMount = append(res.afterMount, f...)
	}
}

func (res *Resource) triggerSchemeAdded(s *Scheme) {
	s.Resource.Trigger(&SchemeEvent{edis.NewEvent(E_SCHEME_ADDED), s})
}

func (res *Resource) HasScheme(name string) bool {
	_, ok := res.GetSchemeOk(name)
	return ok
}

func (res *Resource) DefaultFilter(fns ...func(context *core.Context, db *aorm.DB) *aorm.DB) {
	res.Scheme.DefaultFilter(fns...)
}

func (res *Resource) GetAdminLayout(name string, defaul ...string) *Layout {
	return res.GetLayout(name, defaul...).(*Layout)
}

func (res *Resource) ChildrenLabelKey(childrenID string) string {
	return res.I18nPrefix + ".children." + childrenID
}

func (res *Resource) BasicLayout() *Layout {
	return res.GetLayout(resource.BASIC_LAYOUT).(*Layout)
}

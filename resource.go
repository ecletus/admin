package admin

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	// "github.com/ecletus/responder"
	"strconv"
	"strings"
	"time"

	"github.com/moisespsena-go/i18n-modular/i18nmod"

	"github.com/moisespsena-go/edis"

	"github.com/jinzhu/inflection"
	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/roles"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/db/inheritance"

	"github.com/moisespsena/template/html/template"
)

const (
	SectionLayoutDefault = "default"
	SectionLayoutInline  = "inline"

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
	M_DELETED       = "Deleted"

	META_STRING = "String"

	ActionDelete     = "Delete"
	ActionBulkDelete = "BulkDelete"

	PrintMenu = "PrintMenu"
)

// Resource is the most important thing for qor admin, every model is defined as a resource, qor admin will genetate management interface based on its definition
type Resource struct {
	*resource.Resource
	*Scheme

	Paged

	ObjectPages Paged

	ParentResource *Resource

	// BaseResource resource with contains field as CHILD for this Resource
	// see model:
	//
	// type User struct {
	//      ID bid.BID
	//		Name string
	//		Address Address `aorm:"child"`
	// }
	//
	// type Address struct {
	//		ID bid.BID
	//		Country string
	// }
	// UserRes := Admin.AddResource(&User{})
	// UserAddressRes := UserRes.Meta(&Meta{Name:"Address"}).Resource
	// fmt.Println(UserAddressRes.BaseResource == UserRes)
	BaseResource         *Resource
	Config               *Config
	Metas                []*Meta
	MetasByName          map[string]*Meta
	MetaLinks            map[string]string
	MetasByFieldName     map[string]*Meta
	SingleEditMetas      map[string]*Meta
	Actions              []*Action
	ActionAddedCallbacks []func(action *Action)

	Admin       *Admin
	mounted     bool
	cachedMetas *map[string][]*Meta

	ControllerBuilder     *ResourceControllerBuilder
	ViewControllerBuilder *ResourceViewControllerBuilder

	Router      *xroute.Mux
	ItemRouter  *xroute.Mux
	Parents     []*Resource
	Param       string
	ParamName   string
	paramIDName string

	Resources               map[string]*Resource
	ResourcesByParam        map[string]*Resource
	MetaAliases             map[string]*resource.MetaName
	defaultDisplayName      string
	Children                *Inheritances
	Inherits                map[string]*Child
	Fragments               *Fragments
	Fragment                *Fragment
	initialized             bool
	postInitializeCallbacks []func()
	postMountCallbacks      []func()
	postMetasSetupCallbacks []func()
	RouteHandlers           map[string]*RouteHandler
	ForeignMetas            []*Meta

	labelKey string
	softDelete,
	ReadOnly,
	Virtual bool

	Help, HelpKey,
	PluralHelp, PluralHelpKey string

	NewCrudFunc           func(res *Resource, ctx *core.Context) *resource.CRUD
	MetaContextGetterFunc func(ctx *Context, name string) *Meta
	ContextSetuper        ContextSetuper
	GetContextAttrsFunc   func(ctx *Context) []string

	TemplatePath string

	RecordPermissionFunc func(mode roles.PermissionMode, ctx *Context, record interface{}) (perm roles.Perm)

	DescriptionValuer func(ctx *core.Context, r interface{}) string

	Tags   *ResourceTags
	UITags Tags

	setupMetasCalled bool

	metaUpdateCallbacks MetaUpdateCallbacks

	createWizards      []*Wizard
	CreateWizardByName map[string]*Wizard

	metaEmbedded bool

	ContextPermissioners,
	Permissioners []Permissioner

	ItemRoutes RouteNode
	Routes     RouteNode

	AllSectionsProvider *SchemeSectionsProvider

	UcStates
}

func (this *Resource) GetTemplatePaths() []string {
	if this.TemplatePath != "" {
		return []string{this.TemplatePath}
	}
	return nil
}

func (this *Resource) PostMetasSetup(f ...func()) {
	this.postMetasSetupCallbacks = append(this.postMetasSetupCallbacks, f...)
}

func (this *Resource) HasPermission(mode roles.PermissionMode, context *core.Context) (perm roles.Perm) {
	return this.AdminHasPermission(mode, ContextFromCoreContext(context))
}

func (res *Resource) Permissioner(p Permissioner, pN ...Permissioner) {
	res.Permissioners = append(append(res.Permissioners, p), pN...)
}

func (res *Resource) AdminHasContextPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	if perm = res.adminHasContextPermission(mode, context); perm != roles.UNDEF {
		return
	}
	return res.Resource.HasPermission(mode, context.Context)
}

func (res *Resource) adminHasContextPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	for _, permissioner := range res.ContextPermissioners {
		if perm = permissioner.AdminHasPermission(mode, context); perm != roles.UNDEF {
			return
		}
	}
	return
}

func (this *Resource) autoAddIdMeta(metas []*Meta) (res []resource.Metaor) {
	var (
		metaId *Meta
	)
	if metaId = this.MetasByName["id"]; metaId != nil {
		res = append(res, metaId)
	}
	for _, m := range metas {
		res = append(res, m)
	}
	return
}

func (this *Resource) GetContextMetas(context *core.Context) (metas []resource.Metaor) {
	ctx := ContextFromCoreContext(context)

	if this.GetContextAttrsFunc != nil {
		if ctx.Type.Has(EDIT) {
			metas = this.autoAddIdMeta(this.SectionsList(this.GetContextAttrsFunc(ctx)).ToMetas())
		} else {
			metas = this.SectionsList(this.GetContextAttrsFunc(ctx)).ToMetaors()
		}
		return
	}

	if ctx.Type.Has(EDIT) {
		metas = this.autoAddIdMeta(this.EditAttrs().ToMetas())
	} else if ctx.Type.Has(NEW) {
		metas = this.NewAttrs().ToMetaors()
	} else if ctx.Type.Has(SHOW) {
		metas = this.ShowAttrs().ToMetaors()
	} else if ctx.Type.Has(INDEX) {
		metas = this.IndexAttrs().ToMetaors()
	}
	return metas
}

func (this *Resource) AdminHasRecordPermission(mode roles.PermissionMode, ctx *Context, record interface{}) (perm roles.Perm) {
	if mode == roles.Delete && this.Config.Deletable != nil {
		if !this.Config.Deletable(ctx, record) {
			return roles.DENY
		}
	}
	if this.RecordPermissionFunc != nil {
		perm = this.RecordPermissionFunc(mode, ctx, record)
	}
	return
}

func (this *Resource) HasRecordPermission(mode roles.PermissionMode, ctx *core.Context, record interface{}) (perm roles.Perm) {
	return this.AdminHasRecordPermission(mode, ContextFromCoreContext(ctx), record)
}

func (this *Resource) AdminHasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	if perm = this.HasLocalPermission(mode, ctx); perm != roles.UNDEF {
		return
	}
	if this.DefaultDenyMode() {
		return roles.DENY
	}
	return
}

func (this *Resource) HasLocalPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	if perm = this.adminHasPermission(mode, ctx); perm != roles.UNDEF {
		return
	}
	if perm = this.Resource.HasPermission(mode, ctx.Context); perm != roles.UNDEF {
		return
	}
	return
}

func (this *Resource) adminHasPermission(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
	if ctx.Roles.Has(roles.Anyone) || (mode == roles.Read && ctx.Roles.Has(roles.Viewer)) {
		return roles.ALLOW
	}

	for _, permissioner := range this.Permissioners {
		if perm = permissioner.AdminHasPermission(mode, ctx); perm != roles.UNDEF {
			if _, ok := permissioner.(*Resource); ok {
				if perm.Allow() && this.Permission != nil {
					perm = roles.UNDEF
				}
			}
			if perm != roles.UNDEF {
				return
			}
		}
	}
	return
}

func (this *Resource) TableName(ctx context.Context) string {
	return this.ModelStruct.TableName(ctx, this.Admin.Config.SingularTableName)
}

func (this *Resource) QuotedTableName(DB *aorm.DB) string {
	return aorm.Quote(DB.Dialect(), this.TableName(DB.Context))
}

func (this *Resource) IsSoftDelete() bool {
	return this.softDelete
}

func (this *Resource) Top() (top *Resource) {
	top = this
	for top.ParentResource != nil {
		top = top.ParentResource
	}
	return
}

func (this *Resource) TopAt(parent *Resource) (top *Resource) {
	top = this
	for top.ParentResource != nil && top.ParentResource != parent {
		top = top.ParentResource
	}
	return
}

func (this *Resource) OnDBActionE(cb func(e *resource.DBEvent) error, action ...resource.DBActionEvent) (err error) {
	return resource.OnDBActionE(this, cb, action...)
}

func (this *Resource) OnDBAction(cb func(e *resource.DBEvent), action ...resource.DBActionEvent) (err error) {
	return resource.OnDBAction(this, cb, action...)
}

func (this *Resource) LabelKey() string {
	if this.labelKey != "" {
		return this.labelKey
	}

	return this.I18nPrefix + ".label"
}

func (this *Resource) GetLabelKey(plural bool) string {
	r := this.LabelKey() + "~"
	if plural {
		r += "p"
	} else {
		r += "s"
	}
	return r
}

func (this *Resource) TranslateLabel(ctx i18nmod.Context) string {
	return ctx.T(this.SingularLabelKey()).Default(this.GetDefaultLabel(false)).String()
}

func (this *Resource) PluralLabelKey() string {
	return this.GetLabelKey(true)
}

func (this *Resource) SingularLabelKey() string {
	return this.GetLabelKey(false)
}

func (this *Resource) GetDefaultLabel(plural bool) string {
	if plural {
		return inflection.Plural(this.Name)
	} else {
		return this.Name
	}
}

func (this *Resource) GetLabel(context *Context, plural bool) string {
	return string(context.t(this.GetLabelKey(plural), this.GetDefaultLabel(plural)))
}

func (this *Resource) GetActionLabelKey(action *Action) string {
	return fmt.Sprintf("resources.%v.actions.%v", this.ToParam(), action.Label)
}

func (this *Resource) GetActionLabel(context *Context, action *Action) template.HTML {
	return context.t(this.GetActionLabelKey(action), action.Label)
}

// GetAdmin get admin from resource
func (this *Resource) GetAdmin() *Admin {
	return this.Admin
}

func (this *Resource) FullID() string {
	if this.ParentResource != nil {
		return this.ParentResource.FullID() + "." + this.ID
	}
	return this.ID
}

func (this *Resource) FullPkgPathName() []string {
	if this.ParentResource != nil {
		return append(this.ParentResource.FullPkgPathName(), this.ModelStruct.PkgName())
	}
	return []string{this.ModelStruct.PkgName()}
}

// GetURL
func (this *Resource) GetParentFaultURI(fault func(res *Resource) aorm.ID, parentkeys ...aorm.ID) string {
	child := this
	var p []string
	l := len(parentkeys)
	for i, key := range parentkeys {
		var ix = l - i - 1
		if ix >= len(this.Parents) {
			return ""
		}
		pres := this.Parents[ix]
		p = append(p, pres.ToParam())
		if !pres.Config.Singleton {
			if child.Config.Sub != nil && !child.Config.Sub.MountAsItemDisabled {
				if key == nil {
					key = fault(pres)
				}
				p = append(p, url.PathEscape(key.String()))
			} else if child.Config.Wizard == nil {
				p = append(p, url.PathEscape(key.String()))
			}
		}
		child = pres
	}
	if len(p) > 0 {
		return "/" + strings.Join(p, "/")
	}
	return ""
}

// GetURL
func (this *Resource) FindResource(typ reflect.Type) *Resource {
	typ = IndirectRealType(typ)
	p := this
	for p != nil {
		if p.ModelStruct.Type == typ {
			return p
		}
		for _, sub := range p.Resources {
			if sub.ModelStruct.Type == typ {
				return sub
			}
		}
		p = p.ParentResource
	}
	return nil
}

// GetURL
func (this *Resource) GetParentURI(parentkeys ...aorm.ID) string {
	return this.GetParentFaultURI(func(res *Resource) aorm.ID {
		return aorm.FakeID("{" + res.ParamIDPattern() + "}")
	}, parentkeys...)
}

// GetURL
func (this *Resource) GetIndexURI(ctx *Context, parentkeys ...aorm.ID) string {
	if this.Config.IndexUriHandler != nil {
		return this.Config.IndexUriHandler(ctx, parentkeys...)
	}
	if uri := this.GetParentURI(parentkeys...); len(this.Parents) == 0 || uri != "" {
		return uri + "/" + this.ToParam()
	}
	return ""
}

// GetURL
func (this *Resource) GetURI(ctx *Context, key aorm.ID, parentkeys ...aorm.ID) string {
	if key == nil {
		return ""
	}
	if uri := this.GetIndexURI(ctx, parentkeys...); uri != "" {
		return uri + "/" + url.PathEscape(key.String())
	}
	return ""
}

// GetURL
func (this *Resource) URLFor(ctx *Context, recorde interface{}, parentkeys ...aorm.ID) string {
	return this.GetIndexURI(ctx, parentkeys...) + "/" + url.PathEscape(this.GetKey(recorde).String())
}

// GetURL
func (this *Resource) GetContextIndexURI(context *Context, parentkeys ...aorm.ID) string {
	if parentkeys == nil && this.ParentResource != nil {
		parentkeys = context.ParentResourceID
		if len(parentkeys) == 0 {
			parentkeys = make([]aorm.ID, len(this.Parents))
		}
	}
	return context.Path(this.GetParentFaultURI(func(res *Resource) aorm.ID {
		return resource.MustParseID(res, context.URLParam(res.paramIDName))
	}, parentkeys...), this.ToParam())
}

// GetURL
func (this *Resource) GetContextURI(context *Context, key aorm.ID, parentkeys ...aorm.ID) string {
	base := this.GetContextIndexURI(context, parentkeys...)
	if this.Fragment != nil {
		return base
	} else if key == nil {
		if s := context.URLParam(this.ParamIDName()); s != "" {
			key = resource.MustParseID(this, s)
		}
	}
	if key != nil {
		base += "/" + url.PathEscape(key.String())
	}
	return base
}

func (this *Resource) GetRecordURI(ctx *Context, record interface{}, parentKeys ...aorm.ID) string {
	if ref, ok := record.(inheritance.ParentModelInterface); ok {
		if child := ref.GetQorChild(); child != nil {
			this = this.Children.Items[child.Index].Resource
			// TODO Fix support for subresources
			return this.GetURI(ctx, child.ID)
		}
	}
	if this.Config.RecordUriHandler != nil {
		return this.Config.RecordUriHandler(ctx, record, parentKeys...)
	}
	return this.GetURI(ctx, this.GetKey(record), parentKeys...)
}

func (this *Resource) GetContextRecordURI(ctx *Context, record interface{}, parentKeys ...aorm.ID) string {
	return this.GetRecordURI(ctx, record, parentKeys...)
}

// GetPrimaryValue get priamry value from request
func (this *Resource) GetPrimaryValue(params utils.ReadonlyMapString) string {
	if params != nil {
		return params.Get(this.ParamIDName())
	}
	return ""
}

// ParamIDName return param name for primary key like :product_id
func (this *Resource) ParamIDName() string {
	return this.paramIDName
}

// ParamIDName return param name for primary key like :product_id
func (this *Resource) ParamIDPattern() string {
	return "{" + this.paramIDName + "}"
}

// ToParam used as urls to register routes for resource
func (this *Resource) ToParam() string {
	return this.Param
}

func (this *Resource) DBName(DB *aorm.DB, quote bool) (name string, alias string, pkfields []string) {
	if quote {
		name = this.QuotedTableName(DB)
	} else {
		name = this.TableName(DB)
	}

	alias = "_p" + strconv.Itoa(this.PathLevel)
	for _, f := range this.PrimaryFields {
		pkfields = append(pkfields, f.DBName)
	}

	return
}

func (this *Resource) FilterByParent(ctx *core.Context, db *aorm.DB, parentKey aorm.ID) (_ *aorm.DB, err error) {
	if this.Config.DisableParentJoin {
		return db, nil
	}
	var (
		r                         = this.ParentResource
		res_pkfield               = this.ParentRelation.ForeignDBNames[0]
		parent_name, parent_alias string
		pkfield, fields, wheres   []string
		ids                       []interface{}
	)

	parent_name, parent_alias, pkfield = r.DBName(db, true)
	fields = append(fields, fmt.Sprintf("%[1]v.%[2]v = _.%[3]v", parent_alias,
		pkfield[0], res_pkfield))
	wheres = append(wheres, parent_alias+"."+pkfield[0]+" = ?")
	join := fmt.Sprintf("JOIN %v as %v ON %v", parent_name, parent_alias,
		strings.Join(fields, " AND "))
	ids = r.PrimaryValues(parentKey)
	db = db.Joins(join).Where(strings.Join(wheres, " AND "), ids...)
	r = r.ParentResource
	return db, nil
}

func (this *Resource) GetPathLevel() int {
	return this.PathLevel
}

// Decode decode context into a value
func (this *Resource) Decode(context *core.Context, value interface{}, f ...resource.ProcessorFlag) error {
	return resource.Decode(context, value, this, f...)
}

// NewDecoder return new decoder
func (this *Resource) NewDecoder(context *core.Context, value interface{}) *resource.Decoder {
	return resource.NewDecoder(this, context)
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

func (this *Resource) BasicValue(ctx *core.Context, recorde interface{}) resource.BasicValuer {
	metaLabel, metaIcon := this.MetasByName[BASIC_META_LABEL], this.MetasByName[BASIC_META_ICON]
	id, label, icon := this.GetKey(recorde),
		metaLabel.FormattedValue(ctx, recorde).Value,
		metaIcon.FormattedValue(ctx, recorde).Value
	return &resource.Basic{id, label, icon}
}

func (this *Resource) BasicDescriptableValue(ctx *core.Context, recorde interface{}) resource.BasicDescriptableValuer {
	basic := this.BasicValue(ctx, recorde).(*resource.Basic)
	if this.DescriptionValuer != nil {
		return &resource.BasicDescriptableValue{*basic, this.DescriptionValuer(ctx, recorde)}
	}
	metaHelp := this.MetasByName[META_DESCRIPTIFY]
	help := metaHelp.FormattedValue(ctx, recorde).Value
	return &resource.BasicDescriptableValue{*basic, help}
}

func (this *Resource) DeleteAction() *Action {
	if this.Config.Singleton || !this.ControllerBuilder.IsDeleter() {
		return nil
	}
	return this.Action(&Action{Name: ActionDelete})
}

func (this *Resource) Restore(ctx *Context, key ...aorm.ID) error {
	DB := ctx.DB().
		Table(this.TableName(ctx)).
		ModelStruct(this.ModelStruct).
		Opt(aorm.OptSingleUpdateDisabled()).
		Where(aorm.InID(key...))

	return DB.Restore(this.NewSlicePtr()).Error
}

func (this *Resource) RestoreRecord(ctx *Context, record interface{}, key ...aorm.ID) error {
	DB := ctx.DB().
		Table(this.TableName(ctx)).
		ModelStruct(this.ModelStruct, record)

	return DB.Restore(record).ExpectRowAffected().Error
}

func (this *Resource) IsSoftDeleted(recorde interface{}) bool {
	if this.softDelete {
		if del, ok := recorde.(SoftDeleter); ok {
			return del.IsSoftDeleted()
		}

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
func (this *Resource) GetResources() (resources []*Resource) {
	for _, r := range this.Resources {
		resources = append(resources, r)
	}
	return
}

func (this *Resource) WalkResources(f func(res *Resource) bool) bool {
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
func (this *Resource) GetResourceByID(id string) (resource *Resource) {
	parts := strings.SplitN(id, ".", 2)
	r := this.Resources[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByID(parts[1])
	}
}

// GetResourceByParam get resource with name
func (this *Resource) GetResourceByParam(param string) (resource *Resource) {
	parts := strings.SplitN(param, ".", 2)
	r := this.ResourcesByParam[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByParam(parts[1])
	}
}

func (this *Resource) GetParentResourceByID(id string) *Resource {
	for _, p := range this.Parents {
		if p.ID == id {
			return p
		}
	}
	return this.Admin.GetResourceByID(id)
}

func (this *Resource) GetOrParentResourceByID(id string) *Resource {
	r := this.GetResourceByID(id)
	if r == nil {
		r = this.GetParentResourceByID(id)
	}
	return r
}

func (this *Resource) SubResources() (items []*Resource) {
	for _, r := range this.Resources {
		if !r.Config.Invisible {
			items = append(items, r)
		}
	}
	return
}

func (this *Resource) ReferencedRecord(record interface{}) interface{} {
	return nil
}

func (this *Resource) CrudScheme(ctx *core.Context, scheme interface{}) *resource.CRUD {
	s := this.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = this.GetSchemeOk(st)
	case *Scheme:
		s = st
	}
	return this.Crud(ctx).Dispatcher(s.EventDispatcher)
}

func (this *Resource) CrudSchemeDB(db *aorm.DB, scheme interface{}) *resource.CRUD {
	s := this.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = this.GetSchemeOk(st)
	case *Scheme:
		s = st
	}
	return this.CrudDB(db).Dispatcher(s.EventDispatcher)
}

func (this *Resource) Crud(ctx *core.Context) *resource.CRUD {
	if this.NewCrudFunc == nil {
		return resource.NewCrud(this, ctx)
	}
	return this.NewCrudFunc(this, ctx)
}

func (this *Resource) CrudDB(db *aorm.DB) *resource.CRUD {
	return this.Crud((&core.Context{}).SetDB(db))
}

func (this *Resource) SetParentResource(parent *Resource, relationship *resource.ParentRelationship) {
	this.Resource.SetParent(parent, relationship)
	this.ParentResource = parent
}

func (this *Resource) RegisterScheme(name string, cfg ...*SchemeConfig) *Scheme {
	f := func() *Scheme {
		return this.Scheme.AddChild(name, cfg...)
	}
	if this.initialized {
		return f()
	}
	this.PostInitialize(func() {
		f()
	})
	return nil
}

func (this *Resource) PostInitialize(f ...func()) {
	if this.initialized {
		for _, f := range f {
			f()
		}
		return
	}
	this.postInitializeCallbacks = append(this.postInitializeCallbacks, f...)
}

func (this *Resource) PostMount(f ...func()) {
	if this.mounted {
		for _, f := range f {
			f()
		}
	} else {
		this.postMountCallbacks = append(this.postMountCallbacks, f...)
	}
}

func (this *Resource) triggerSchemeAdded(s *Scheme) {
	if err := s.Resource.Trigger(&SchemeEvent{edis.NewEvent(E_SCHEME_ADDED), s}); err != nil {
		panic(err)
	}
}

func (this *Resource) HasScheme(name string) bool {
	_, ok := this.GetSchemeOk(name)
	return ok
}

func (this *Resource) GetAdminLayout(name string, defaul ...string) *Layout {
	return this.GetLayout(name, defaul...).(*Layout)
}

func (this *Resource) ChildrenLabelKey(childrenID string) string {
	return this.I18nPrefix + ".children." + childrenID
}

func (this *Resource) BasicLayout() *Layout {
	return this.GetLayout(resource.BASIC_LAYOUT).(*Layout)
}

func (this *Resource) AddForeignMeta(meta *Meta) {
	for _, m := range this.ForeignMetas {
		if m == meta {
			return
		}
	}
	this.ForeignMetas = append(this.ForeignMetas, meta)
	this.triggerForeignMetaAdded(meta)
}

func (this *Resource) GetLayout(name string, defaul ...string) resource.LayoutInterface {
	l := this.Resource.GetLayout(name, defaul...)
	if l == nil {
		var typ = ParseContextType(name)
		if typ.String() == name {
			if typ.Has(INDEX, NEW, SHOW, EDIT) {
				return this.Resource.GetLayout(resource.DEFAULT_LAYOUT)
			}
		}
	}
	return l
}

type ForeignMetaEvent struct {
	edis.EventInterface
	Resource *Resource
	Meta     *Meta
}

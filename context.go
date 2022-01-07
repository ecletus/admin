package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/moisespsena-go/getters"
	"github.com/moisespsena-go/maps"
	"unapu.com/lib"

	"github.com/ecletus/validations"

	"github.com/moisespsena-go/i18n-modular/i18nmod"

	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/render"

	"github.com/ecletus/core/resource"

	"github.com/ecletus/responder"
	"github.com/ecletus/roles"
	"github.com/ecletus/session"

	"github.com/moisespsena/template/html/template"

	"github.com/moisespsena-go/assetfs/assetfsapi"

	"github.com/ecletus/core"
)

type ContextType uint16

func (b ContextType) Set(flag ContextType) ContextType    { return b | flag }
func (b ContextType) Clear(flag ContextType) ContextType  { return b &^ flag }
func (b ContextType) Toggle(flag ContextType) ContextType { return b ^ flag }
func (b ContextType) Has(flag ...ContextType) bool {
	for _, flag := range flag {
		if (b & flag) != 0 {
			return true
		}
	}
	return false
}

func (b ContextType) ClearBasic() ContextType {
	return b &^ INDEX &^ SHOW &^ EDIT &^ NEW &^ ACTION
}

func (b ContextType) ClearCrud() ContextType {
	return b &^ INDEX &^ SHOW &^ EDIT &^ NEW
}

func (b *ContextType) SetBasic(typ ContextType) {
	*b = b.ClearBasic()
	*b |= typ
}

func (b *ContextType) SetBasicS(typ string) {
	var basic ContextType
	basic.Parse(typ)
	*b = b.ClearBasic()
	*b |= basic
}

func (b ContextType) HasAll(flag ...ContextType) bool {
	if len(flag) == 0 {
		return false
	}

	for _, flag := range flag {
		if (b & flag) == 0 {
			return false
		}
	}
	return true
}

func (b ContextType) String() string {
	var s []string
	if b.Has(INDEX) {
		s = append(s, "index")
	}
	if b.Has(NEW) {
		s = append(s, "new")
	}
	if b.Has(SHOW) {
		s = append(s, "show")
	}
	if b.Has(EDIT) {
		s = append(s, "edit")
	}
	if b.Has(DELETED) {
		s = append(s, "deleted")
	}
	if b.Has(ACTION) {
		s = append(s, "action")
	}
	if b.Has(META_INDEX) {
		s = append(s, "metaindex")
	}
	if b.Has(PRINT) {
		s = append(s, "print")
	}
	if b.Has(INLINE) {
		s = append(s, "row")
	}
	if b.Has(RE_RENDER) {
		s = append(s, "re_render")
	}
	return strings.Join(s, "|")
}

func (b ContextType) HasS(s ...string) bool {
	for _, s := range s {
		switch s {
		case "index":
			if b.Has(INDEX) {
				return true
			}
		case "new":
			if b.Has(NEW) {
				return true
			}
		case "show":
			if b.Has(SHOW) {
				return true
			}
		case "edit":
			if b.Has(EDIT) {
				return true
			}
		case "deleted":
			if b.Has(DELETED) {
				return true
			}
		case "action":
			if b.Has(ACTION) {
				return true
			}
		case "meta_index":
			if b.Has(META_INDEX) {
				return true
			}
		case "print":
			if b.Has(PRINT) {
				return true
			}
		case "inline":
			if b.Has(INLINE) {
				return true
			}
		case "re_render":
			if b.Has(RE_RENDER) {
				return true
			}
		}
	}
	return false
}

func ParseContextType(s string) (b ContextType) {
	b.ParseMerge(s)
	return
}

func (b *ContextType) ParseMerge(s string) {
	for _, s := range strings.Split(strings.ToLower(s), "|") {
		switch s {
		case "index":
			*b = b.ClearCrud() | INDEX
		case "new":
			*b = b.ClearCrud() | NEW
		case "show":
			*b = b.ClearCrud() | SHOW
		case "edit":
			*b = b.ClearCrud() | EDIT
		case "deleted":
			*b |= DELETED
		case "action":
			*b |= ACTION
		case "meta_index":
			*b |= META_INDEX
		case "print":
			*b |= PRINT
		case "inline":
			*b |= INLINE
		case "re_render":
			*b |= RE_RENDER
		}
	}
}

func (b *ContextType) Parse(s string) {
	*b = 0
	b.ParseMerge(s)
}

func (ct ContextType) S() string {
	return ct.String()
}

const (
	_ ContextType = 1 << iota
	INDEX
	NEW
	SHOW
	EDIT
	DELETE
	DELETED
	ACTION
	META_INDEX
	PRINT
	INLINE
	RE_RENDER // re-render posted form
)

// Context admin context, which is used for admin controller
type Context struct {
	*core.Context
	*Searcher
	Scheme         *Scheme
	SectionLayout  string
	ParentRecord   []interface{}
	ParentResource []*Resource
	Resource       *Resource
	ResourceType   string
	Admin          *Admin
	Content        template.HTML
	TemplateName   string
	Action         string
	Settings       map[string]interface{}
	// Result is a generic result
	Result,
	// ResourceRecord is a resource record
	ResourceRecord interface{}
	ResourceItems interface{}

	PageTitle string

	usedThemes     []string
	funcMaps       []template.FuncMap
	funcValues     template.FuncValues
	PermissionMode roles.PermissionMode
	Display        string
	Type           ContextType
	NotFound       bool
	RouteHandler   *RouteHandler
	ReadOnly       bool

	Alerts []template.HTML

	ParentResults                             []interface{}
	nestedForm                                int
	encodes                                   []EncoderInterface
	RequestLayout                             string
	Yield                                     func(w io.Writer, results ...interface{})
	SiteAssetFS, SiteTemplateFS, SiteStaticFS assetfsapi.Interface

	templatesStack *PathStack

	Layout string
	jsLibs *lib.Libs

	IDParser func(ctx *Context, res *Resource, value string) (ID aorm.ID, err error)

	config       map[interface{}]interface{}
	PageHandlers *render.PageHandlers
	Parent       *Context

	WizardCompleteConfig *WizardContextCompleteConfig

	MetaStack      *MetaStack
	RenderFlags    ContextRenderFlag
	ResourceAction *Action

	Warnings core.Errors
}

func NewContext(arg ...*Context) (ctx *Context) {
	for _, arg := range arg {
		ctx = arg
	}
	if ctx == nil {
		ctx = &Context{}
	}
	if ctx.MetaStack == nil {
		ctx.MetaStack = &MetaStack{}
	}
	return
}

func (this *Context) WithType(typ ContextType, do func()) {
	oldType := this.Type
	this.Type = typ
	defer func() {
		this.Type = oldType
	}()
	do()
}

func (this *Context) WithTypeE(typ ContextType, do func() error) error {
	oldType := this.Type
	this.Type = typ
	defer func() {
		this.Type = oldType
	}()
	return do()
}

func (this *Context) SetBasicType(t ContextType) *Context {
	this.Type = this.Type.ClearBasic() | t
	return this
}

func (this *Context) GetParentRecordContext() *Context {
	if this.Parent != nil {
		if this.Parent.ResourceRecord != nil {
			return this.Parent
		} else if this.Parent.ResourceItems != nil {
			return this.Parent.Parent
		}
	}
	return nil
}

func (this *Context) GetParentRecord() interface{} {
	if len(this.ParentRecord) > 0 {
		return this.ParentRecord[len(this.ParentRecord)-1]
	}
	if ctx := this.GetParentRecordContext(); ctx != nil {
		return ctx.ResourceRecord
	}
	return nil
}

func (this *Context) LoadResourceRecord(record interface{}, id ...aorm.ID) (err error) {
	var id_ aorm.ID
	for _, id_ = range id {
	}
	if id_ == nil {
		id_ = this.ResourceID
	}
	return this.DB().ModelStruct(this.Resource.ModelStruct).First(record, id_).Error
}

func (this *Context) ConfigSet(key, value interface{}) {
	if this.config == nil {
		this.config = map[interface{}]interface{}{}
	}
	this.config[key] = value
}

func (this *Context) ConfigGet(key interface{}) (value interface{}, ok bool) {
	if this.config == nil {
		return
	}
	value, ok = this.config[key]
	return
}

const (
	P_LAYOUT  = "_layout"
	P_DISPLAY = "_display"
)

func (this *Context) IsSuperUser() bool {
	return this.Roles.Has(roles.Anyone)
}

func (this *Context) WithDB(f func(context *Context), db ...*aorm.DB) *Context {
	this.Context.WithDB(func(ctx *core.Context) {
		this.Context = ctx
		f(this)
	}, db...)
	return this
}

func (this *Context) IsResultSlice() bool {
	if this.Result != nil {
		value := reflect.ValueOf(this.Result)
		return value.Kind() == reflect.Slice || value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Slice
	}
	return false
}

func (this *Context) WithResource(res *Resource, value interface{}) func() {
	id, parentResource, parentResourceID, parentRecord, resource, searcher, DB, result, record, scheme := this.ResourceID,
		this.ParentResource, this.ParentResourceID, this.ParentRecord, this.Resource, this.Searcher, this.DB(), this.Result,
		this.ResourceRecord, this.Scheme

	newDB := DB
	if this.Context.Parent != nil {
		newDB = this.Context.Parent.DB()
	}

	this.ResourceID,
		this.ParentResource, this.ParentResourceID, this.ParentRecord, this.Resource, this.Searcher, this.Result,
		this.ResourceRecord = nil, []*Resource{}, []aorm.ID{}, []interface{}{}, res, this.NewSearcher(), value, value

	this.SetRawDB(newDB)

	if scheme == nil {
		this.Scheme = res.Scheme
	} else if scheme.Resource != res {
		this.Scheme = res.Scheme
	}

	if value != nil {
		this.ResourceID = res.GetKey(value)
	}
	return func() {
		this.ResourceID,
			this.ParentResource, this.ParentResourceID, this.ParentRecord, this.Resource, this.Searcher, this.Result, this.ResourceRecord,
			this.Scheme = id, parentResource, parentResourceID, parentRecord, resource, searcher, result, record, scheme
		this.SetRawDB(DB)
	}
}

func (this *Context) Stringify(value interface{}, opt ...getters.Getter) string {
	var o getters.Getter
	for _, o = range opt {
	}
	if o == nil {
		var m maps.Map
		o = m
	}
	if value == nil {
		return ""
	}
	switch t := value.(type) {
	case Stringer:
		return t.AdminString(this, o)
	case core.ContextStringer:
		return t.ContextString(this.Context)
	case i18nmod.Translater:
		return t.Translate(this.GetI18nContext())
	default:
		return utils.Stringify(value)
	}
}

func (this *Context) HtmlifyRecordsMeta(res *Resource, metaName string, records ...interface{}) (result []template.HTML) {
	if len(records) == 0 {
		return
	}
	defer this.WithResource(res, nil)()
	result = make([]template.HTML, len(records))
	qorContext := this.CloneBasic()
	valuer := res.GetDefinedMeta(metaName).GetFormattedValuer()
	var v interface{}
	for i, item := range records {
		if item == nil {
			continue
		}
		this.ResourceID = res.GetKey(item)
		v = valuer(item, qorContext)
		if v != nil {
			switch vt := v.(type) {
			case template.HTML:
				result[i] = vt
			case string:
				result[i] = template.HTML(vt)
			default:
				result[i] = this.HtmlifyInterfaces(v)[0]
			}
		}
	}
	return
}

func (this *Context) StringifyRecordsMeta(res *Resource, metaName string, records ...interface{}) (result []string) {
	if len(records) == 0 {
		return
	}
	defer this.WithResource(res, nil)()
	result = make([]string, len(records))
	qorContext := this.CloneBasic()
	valuer := res.GetDefinedMeta(metaName).GetFormattedValuer()
	var v interface{}
	for i, item := range records {
		if item == nil {
			continue
		}
		this.ResourceID = res.GetKey(item)
		v = valuer(item, qorContext)
		if v != nil {
			switch vt := v.(type) {
			case string:
				result[i] = vt
			default:
				result[i] = utils.ToString(v)
			}
		}
	}
	return
}

func (this *Context) StringifyRecords(res *Resource, records ...interface{}) (result []string) {
	return this.StringifyRecordsMeta(res, META_STRINGIFY, records...)
}

func (this *Context) StringifyRecord(res *Resource, record interface{}) string {
	return this.StringifyRecordsMeta(res, META_STRINGIFY, record)[0]
}

func (this *Context) HtmlifyRecordMeta(res *Resource, metaName string, record interface{}) template.HTML {
	return this.HtmlifyRecordsMeta(res, metaName, record)[0]
}

func (this *Context) HtmlifyRecords(res *Resource, records ...interface{}) (result []template.HTML) {
	return this.HtmlifyRecordsMeta(res, BASIC_META_HTML, records...)
}

func (this *Context) HtmlifyRecord(res *Resource, record interface{}) template.HTML {
	return this.HtmlifyRecordMeta(res, BASIC_META_HTML, record)
}

func (this *Context) HtmlifyInterfaces(values ...interface{}) (result []template.HTML) {
	result = make([]template.HTML, len(values))
	for i, value := range values {
		if value == nil {
			continue
		}
		switch vt := value.(type) {
		case interface{ Htmlify(*Context) template.HTML }:
			result[i] = vt.Htmlify(this)
		case *FormattedValue:
			if vt.SafeValue != "" {
				result[i] = template.HTML(vt.SafeValue)
			} else if vt.Value != "" {
				result[i] = template.HTML(vt.Value)
			} else {
				result[i] = this.Context.Htmlify(vt.Raw)
			}
		default:
			result[i] = this.Context.Htmlify(value)
		}
	}
	return
}

func (this *Context) HtmlifyInterface(value interface{}) (result template.HTML) {
	if value == nil {
		return
	}
	switch vt := value.(type) {
	case interface{ Htmlify(*Context) template.HTML }:
		return vt.Htmlify(this)
	default:
		return this.Context.Htmlify(value)
	}
}

func (this *Context) HtmlifyItems(values ...interface{}) (result []template.HTML) {
	if l := len(values); l > 0 {
		if res, ok := values[0].(*Resource); ok {
			if l == 1 {
				return
			}
			return this.HtmlifyRecords(res, values[1:]...)
		}
	}
	return this.HtmlifyInterfaces(values...)
}

func (this *Context) Htmlify(value interface{}, res ...*Resource) (result template.HTML) {
	if len(res) > 0 {
		return this.HtmlifyRecords(res[0], value)[0]
	}
	return this.HtmlifyInterfaces(value)[0]
}

func (this *Context) ValidateLayout() bool {
	l := this.Resource.GetLayout(this.Layout)
	if l == nil {
		this.AddError(fmt.Errorf("Layout %q is not valid.", this.Layout))
		return false
	}
	return true
}

func (this *Context) ValidateLayoutOrError() bool {
	if !this.ValidateLayout() {
		this.SendError()
		return false
	}
	return true
}

func (this *Context) LoadDisplay(displayType string) bool {
	if this.HasError() {
		return false
	}

	if this.Display != "" {
		display := this.Resource.GetDisplay(displayType + "/" + this.Display)
		if display == nil {
			this.AddError(fmt.Errorf("Display %q does not exists.", this.Display))
		} else {
			this.Layout = display.GetLayoutName()
		}
		this.ValidateLayout()
	} else {
		this.Display = this.Resource.GetDefaultDisplayName()
	}
	return !this.HasError()
}

func (this *Context) TypeS() string {
	return this.Type.S()
}

func (this *Context) Is(values ...interface{}) bool {
	for _, v := range values {
		switch vt := v.(type) {
		case ContextType:
			if this.Type.Has(vt) {
				return true
			}
		case string:
			if this.Type.HasS(vt) {
				return true
			}
		}
	}
	return false
}

func (this *Context) LoadDisplayOrError(displayType ...string) bool {
	if len(displayType) == 0 || displayType[0] == "" {
		displayType = []string{this.Type.S()}
	}
	if !this.LoadDisplay(displayType[0]) {
		this.Writer.WriteHeader(http.StatusPreconditionFailed)
		return false
	}
	return true
}

func (parent *Context) CreateReadOnlerTypedChild(ro bool, typ interface{}, res *Resource, record ...interface{}) *Context {
	child := parent.CreateTypedChild(typ, res, record...)
	child.ReadOnly = ro
	return child
}

func (parent *Context) CreateTypedChild(typ interface{}, res *Resource, record ...interface{}) *Context {
	var T ContextType
	switch t := typ.(type) {
	case string:
		T = ParseContextType(t)
	default:
		T = t.(ContextType)
	}
	child := parent.CreateChild(res, record...)
	child.Type |= T
	return child
}

func (parent *Context) CreateChild(res *Resource, record ...interface{}) *Context {
	this := parent.Clone()
	_, this.Context = this.Context.NewChild(nil)
	this.Context.LocalContext = *this.Context.LocalContext.Copy()
	this.Context.SetValue(CONTEXT_KEY, this)
	this.Parent = parent
	this.Resource = res
	this.ParentRecord = nil
	this.ParentResourceID = nil
	this.ParentResource = nil
	if this.ResourceID != nil {
		this.ParentResourceID = []aorm.ID{this.ResourceID}
	}
	if res != parent.Resource && parent.Scheme != nil {
		this.Scheme = res.Scheme
	}
	if len(record) == 1 && record[0] != nil {
		this.Result = record[0]

		if t := IndirectRealType(reflect.TypeOf(record[0])); t == res.ModelStruct.Type {
			this.ResourceID = res.GetKey(record[0])
			this.ResourceRecord = record[0]
		} else {
			this.ResourceRecord = nil
			this.ResourceID = nil
			switch t.Kind() {
			case reflect.Slice, reflect.Chan:
				if IndirectRealType(t.Elem()) == res.ModelStruct.Type {
					this.ResourceItems = record[0]
				}
			}
		}
	} else {
		this.Result = nil
		this.ResourceID = nil
		this.ResourceRecord = nil
	}
	return this
}

// Funcs set FuncMap for templates
func (this *Context) Funcs(funcMaps ...template.FuncMap) *Context {
	this.funcMaps = append(this.funcMaps, funcMaps...)
	return this
}

// Flash set flash message
func (this *Context) Flash(message string, typ string) {
	this.SessionManager().Flash(session.Message{
		Message: template.HTML(message),
		Type:    typ,
	})
}

// FlashS set flash messages
func (this *Context) FlashS(typ string, message ...string) {
	for _, message := range message {
		this.SessionManager().Flash(session.Message{
			Message: template.HTML(message),
			Type:    typ,
		})
	}
}

// NewResourceContext new resource context
func (this *Context) NewResourceContext(name ...interface{}) *Context {
	clone := &Context{
		Context:        this.Context.Clone(),
		Admin:          this.Admin,
		Result:         this.Result,
		ResourceRecord: this.ResourceRecord,
		Action:         this.Action,
		Type:           this.Type,
		MetaStack:      this.MetaStack,
	}

	if len(name) > 0 {
		if str, ok := name[0].(string); ok {
			clone.setResource(this.Admin.GetResourceByID(str))
		} else if res, ok := name[0].(*Resource); ok {
			clone.setResource(res)
		}
	} else {
		clone.setResource(this.Resource)
	}
	return clone
}

func (this *Context) WithResult(result interface{}) func() {
	l := len(this.ParentResults)
	this.ParentResults = append(this.ParentResults, this.Result)
	this.Result = result
	return func() {
		this.Result = this.ParentResults[l]
		this.ParentResults = this.ParentResults[0:l]
	}
}

func (this *Context) IsAction(name string, names ...string) bool {
	if this.Action == name {
		return true
	}

	for _, name = range names {
		if this.Action == name {
			return true
		}
	}

	return false
}

// Get get context's Settings
func (this *Context) GetSettings(key string) interface{} {
	return this.Settings[key]
}

// Set set context's Settings
func (this *Context) SetSettings(key string, value interface{}) {
	this.Settings[key] = value
}

func (this *Context) resourcePath() string {
	if this.Resource == nil {
		return ""
	}
	return this.Resource.TemplatePath
}

func (this *Context) NewSearcher() *Searcher {
	s := &Searcher{Context: this}
	if this.Request != nil {
		s.Layout = this.Request.URL.Query().Get(P_LAYOUT)
	}
	return s
}

func (this *Context) setResource(res *Resource, recorde ...interface{}) *Context {
	if res != nil {
		this.Resource = res
		if this.Scheme == nil || (this.Scheme != nil && this.Scheme.Resource != res) {
			this.Scheme = res.Scheme
		}
		if this.ResourceID == nil {
			if len(recorde) == 1 && recorde[1] != nil {
				this.ResourceID = res.GetKey(recorde)
			} else if idS := this.URLParam(res.ParamIDName()); idS != "" {
				var err error
				this.ResourceID, err = res.ParseID(idS)
				if this.AddError(err) != nil {
					return nil
				}
			} else if this.IDParser != nil {
				var err error
				this.ResourceID, err = this.IDParser(this, res, idS)
				if this.AddError(err) != nil {
					return nil
				}
			}
		}
	}
	this.Searcher = this.NewSearcher()
	return this
}

func (this *Context) setResourceFromCrumber(crumber *ResourceCrumber) *Context {
	this.Resource = crumber.Resource
	this.ResourceID = crumber.ID
	this.ParentResourceID = crumber.ParentID
	this.ParentResource = crumber.Parent
	this.ParentRecord = make([]interface{}, len(crumber.Parent))
	this.Scheme = crumber.Resource.Scheme
	this.Searcher = this.NewSearcher()
	return this
}

func (this *Context) SetResource(res *Resource, recorde ...interface{}) *Context {
	return this.setResource(res)
}

func (this *Context) SetResourceWithDB(res *Resource) *Context {
	ctx := this.setResource(res)
	ctx.SetRawDB(ctx.DB().NewScope(res.Value).DB())
	return ctx
}

func (this *Context) getFlashes() []session.Message {
	return this.SessionManager().Flashes()
}

// JSON generate json outputs for action
func (this *Context) JSON(result interface{}, action ...string) {
	if this.Encode(result, action...) == nil {
		this.Writer.Header().Set("Content-Type", "application/json")
	}
}

func (this *Context) Encode(result interface{}, layout ...string) error {
	if len(layout) == 0 {
		layout = []string{this.Layout}
	}

	if layout[0] == "show" && !this.Scheme.Sections.Default.Screen.Show.IsSetI() {
		layout[0] = "edit"
	}

	encoder := &Encoder{
		Layout:   layout[0],
		Resource: this.Resource,
		Context:  this,
		Result:   result,
	}

	oldRenderFlags := this.RenderFlags

	this.RenderFlags |= CtxRenderEncode
	defer func() {
		this.RenderFlags = oldRenderFlags
	}()

	return this.Admin.Encode(func(e EncoderInterface) (done func()) {
		i := len(this.encodes)
		this.encodes = append(this.encodes, e)
		return func() {
			this.encodes = this.encodes[0:i]
		}
	}, this.Writer, encoder)
}

func (this *Context) Encodes() bool {
	return len(this.encodes) > 0
}

func (this *Context) Encoder() EncoderInterface {
	if len(this.encodes) == 0 {
		return nil
	}
	return this.encodes[len(this.encodes)-1]
}

func (this *Context) SendError() bool {
	if this.HasError() {
		responder.With("html", func() {
			this.Flash(this.Error(), "error")
		}).With("json", func() {
			this.Encode(map[string]interface{}{"errors": this.GetErrors()})
		}).Respond(this.Request)
		return true
	}
	return false
}

// GetSearchableResources get defined searchable resources has performance
func (this *Context) GetSearchableResources() (resources []*Resource) {
	if admin := this.Admin; admin != nil {
		for _, res := range admin.searchResources {
			if this.HasPermission(res, roles.Read) {
				resources = append(resources, res)
			}
		}
	}
	return
}

func (this *Context) GetActionLabel() string {
	var defaul string
	key := I18NGROUP + ".action." + this.Action

	switch this.Type {
	case NEW:
		defaul = "Add {{.}}"
	case EDIT:
		defaul = "Edit {{.}}"
	case SHOW:
		defaul = "{{.}} Details"
	default:
		return ""
	}

	return string(this.t(key, defaul))
}

func (this *Context) Crud(ctx ...*core.Context) *resource.CRUD {
	if len(ctx) == 0 {
		ctx = append(ctx, this.Context)
	}
	return this.Resource.CrudScheme(ctx[0], this.Scheme)
}

func (this *Context) WithTransaction(f func() (err error)) (err error) {
	oldDB := this.DB()
	return oldDB.WithTransaction(func(db *aorm.DB) (err error) {
		this.SetRawDB(db)
		defer this.SetRawDB(oldDB)
		return f()
	})
}

func (this *Context) LogErrors() {
	if errors := this.Errors.Filter(func(err error) error {
		if validations.IsError(err) {
			return nil
		}
		return err
	}); errors.HasError() {
		panic(errors)
	}
}

// AddError add error to Errors struct
func (this *Context) AddError(errors ...error) error {
	for i, err := range errors {
		if err != nil {
			if d, ok := err.(resource.DuplicateUniqueIndexError); ok {
				var labels []string
				for _, f := range d.Index().Fields {
					meta := d.Resource().(*Resource).Meta(&Meta{Name: f.Name})
					labels = append(labels, meta.GetRecordLabel(this, d.Record()))
				}
				var msg = this.GetI18nContext().TT(I18NGROUP + ".errors.duplicate_unique_index").Data(map[string]interface{}{
					"Field":  labels[0],
					"Fields": labels,
				}).Count(len(labels)).Get()
				errors[i] = validations.FieldFailed(d.Record(), d.Index().Fields[0].Name, msg)
			}
		}
	}
	return this.Context.AddError(errors...)
}

func GetContext(ctx context.Context) *Context {
	return ctx.Value(CONTEXT_KEY).(*Context)
}

func GetOrNewContext(ctx context.Context, admin *Admin) *Context {
	if c := ctx.Value(CONTEXT_KEY); c != nil {
		return c.(*Context)
	}
	return admin.NewContext(ctx)
}

func (this *Admin) RenderContext(s *template.State) *Context {
	return GetOrNewContext(render.Context(s), this)
}

const CONTEXT_KEY = contextKey("admin.context")

func ContextFromCoreContext(ctx *core.Context) *Context {
	if i := ctx.GetValue(CONTEXT_KEY); i != nil {
		return i.(*Context)
	}
	return nil
}

func ContextFromRequest(r *http.Request) *Context {
	return core.ContextFromRequest(r).Value(CONTEXT_KEY).(*Context)
}

func ContextFromContext(ctx context.Context) *Context {
	switch t := ctx.(type) {
	case *Context:
		return t
	case *core.Context:
		return ContextFromCoreContext(t)
	case AdminContextGetter:
		return t.Context()
	default:
		if i := ctx.Value(CONTEXT_KEY); i != nil {
			return i.(*Context)
		}
		return nil
	}
}

type contextKey string

type AdminContextGetter interface {
	Context() *Context
}

type ContextSetuper interface {
	ContextSetup(ctx *Context) (err error)
}

type ContextSetupFunc = func(ctx *Context) (err error)

type contextSetupFunc ContextSetupFunc

func (f contextSetupFunc) ContextSetup(ctx *Context) (err error) {
	return f(ctx)
}

func NewContextSetup(f ContextSetupFunc) ContextSetuper {
	return contextSetupFunc(f)
}

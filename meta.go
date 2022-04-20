package admin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ecletus/fragment"
	"github.com/moisespsena-go/i18n-modular/i18nmod"

	"github.com/moisespsena-go/maps"

	"github.com/pkg/errors"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/edis"

	"github.com/ecletus/roles"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"

	"github.com/go-aorm/aorm"
)

const (
	BASIC_LAYOUT                            = "basic"
	BASIC_LAYOUT_DESCRIPTION                = "basic_help"
	BASIC_LAYOUT_WITH_ICON                  = "basic_with_icon"
	BASIC_LAYOUT_DESCRIPTION_WITH_ICON      = "basic_help_with_icon"
	BASIC_LAYOUT_HTML                       = "basic_html"
	BASIC_LAYOUT_HTML_WITH_ICON             = "basic_html_with_icon"
	BASIC_LAYOUT_HTML_DESCRIPTION           = "basic_html_help"
	BASIC_LAYOUT_HTML_DESCRIPTION_WITH_ICON = "basic_html_help_with_icon"

	BASIC_META_ID    = "_BasicID"
	BASIC_META_LABEL = "_BasicLabel"
	BASIC_META_HTML  = "_BasicHTML"
	BASIC_META_ICON  = "_BasicIcon"

	META_DESCRIPTIFY = "Descriptify"
	META_STRINGIFY   = "Stringify"
)

type DependencyParent struct {
	Meta  *Meta
	Value aorm.ID
}

type DependencyPath struct {
	Meta *Meta
}

type DependencyQuery struct {
	Meta  *Meta
	Param string
}

type DependencyValue struct {
	Param string
	Value interface{}
}

type MetaOutputValuer func(context *core.Context, record, value interface{}) interface{}
type MetaValuer func(record interface{}, context *core.Context) interface{}
type MetaFValuer func(record interface{}, context *core.Context) *FormattedValue
type MetaSetter func(record interface{}, metaValue *resource.MetaValue, context *core.Context) error
type MetaEnabled func(record interface{}, context *Context, meta *Meta, readOnly bool) bool

type MetaPostFormatted interface {
	MetaPostFormatted(meta *Meta, ctx *core.Context, record, value interface{}) interface{}
}

// Meta meta struct definition
type Meta struct {
	edis.EventDispatcher

	*resource.Meta

	Name             string
	DB               *aorm.Alias
	Type             string
	ReadOnlyType     string
	ReadOnlyStringer bool
	TypeHandler      func(meta *Meta, recorde interface{}, ctx *Context) string
	Enabled          MetaEnabled

	SkipDefaultLabel     bool
	DefaultLabel         string
	Label                string
	HiddenLabel          bool
	recordLabelPairFuncs []func(meta *Meta, ctx *Context, record interface{}) (key string, defaul string, ok bool)

	Help        string
	HelpKey     string
	ShowHelp    string
	ShowHelpKey string

	recordHelpPairFuncs,
	recordShowHelpPairFuncs []func(meta *Meta, ctx *Context, record interface{}) (key string, defaul string, ok bool)

	FieldName         string
	FieldLabel        bool
	EncodedName       string
	Setter            MetaSetter
	Valuer            MetaValuer
	FormattedValuer   MetaFValuer
	DefaultFormat     string
	DefaultFormatF    func(value interface{}) string
	ContextResourcer  func(meta resource.Metaor, ctx *core.Context) resource.Resourcer
	ContextMetas      func(record interface{}, context *Context) []*Meta
	SkipResourceModel bool
	Resource          *Resource
	Permission        *roles.Permission
	Config            MetaConfigInterface
	Icon              bool

	Metas        []resource.Metaor
	GetMetasFunc func() []resource.Metaor
	Collection   interface{}

	BaseResource          *Resource
	TemplateData          map[string]interface{}
	i18nGroup             string
	Dependency            []interface{}
	ProxyTo               *Meta
	Proxies               []*Meta
	Include               bool
	ForceShowZero         bool
	IsZeroFunc            func(meta *Meta, record, value interface{}) bool
	Fragment              *Fragment
	Options               maps.Map
	OutputFormattedValuer MetaOutputValuer
	DefaultValueFunc      MetaValuer
	proxyPath             []ProxyPath
	Typ                   reflect.Type
	DefaultDeny           bool
	Virtual               bool

	SectionNotAllowed bool
	ReadOnly          bool
	ReadOnlyFunc      func(meta *Meta, ctx *Context, record interface{}) bool
	ReadOnlyItemFunc  func(meta *Meta, ctx *Context, record, item interface{}) bool
	URIForFunc        func(meta *Meta, ctx *Context, record interface{}) string
	URLForFunc        func(meta *Meta, ctx *Context, record interface{}) string

	Required   bool
	mustValuer bool

	ForceEmptyFormattedRender bool

	// if require specify to show
	DefaultInvisible bool

	Tags            MetaTags
	tagsInitialized bool

	NilAsZero bool

	afterUpdate []func()

	UITags Tags

	AdminData maps.SyncedMap

	ReadOnlyFormattedValuerFunc func(meta *Meta, ctx *Context, record interface{}) *FormattedValue

	RecordLabelFormatter,
	RecordHelpFormatter,
	RecordShowHelpFormatter func(meta *Meta, ctx *Context, record interface{}, s string) string

	LabelFormatter,
	HelpFormatter func(meta *Meta, ctx *Context, s string) string

	LockedFunc func(meta *Meta, ctx *Context, record interface{}) bool

	RecordRequirerFunc func(ctx *core.Context, record interface{}) bool

	CanCreateItemFunc func(ctx *Context, record interface{}) bool
	CanDeleteItemFunc func(ctx *Context, record, item interface{}) bool
	CanUpdateItemFunc func(ctx *Context, record, item interface{}) bool

	Disabled       bool
	ItemUrlFunc    func(ctx *Context, record, item interface{}) string
	ParentMetaName string

	Permissioners []Permissioner

	updateCallbacks []func(m *Meta)

	SortHandler func(s *Searcher, m *Meta, order aorm.Order) aorm.Clauser

	beforeRenderHandlers   []func(ctx *MetaContext, record interface{})
	beforeDoRenderHandlers []func(ctx *MetaContext, record interface{})
	prepareContextHandlers []func(ctx *MetaContext, record interface{})
}

func MetaAliases(tuples ...[]string) map[string]*resource.MetaName {
	m := make(map[string]*resource.MetaName)
	for _, t := range tuples {
		if len(t) == 2 {
			m[t[0]] = &resource.MetaName{Name: t[1]}
		} else if len(t) == 3 {
			m[t[0]] = &resource.MetaName{t[1], t[2]}
		}
	}
	return m
}

// GetName get meta's name
func (this *Meta) GetName() string {
	return this.Name
}

func (this *Meta) BeforeRender(f func(ctx *MetaContext, record interface{})) *Meta {
	this.beforeRenderHandlers = append(this.beforeRenderHandlers, f)
	return this
}

func (this *Meta) BeforeDoRender(f func(ctx *MetaContext, record interface{})) *Meta {
	this.beforeDoRenderHandlers = append(this.beforeDoRenderHandlers, f)
	return this
}

func (this *Meta) PrepareContext(f func(ctx *MetaContext, record interface{})) *Meta {
	this.prepareContextHandlers = append(this.prepareContextHandlers, f)
	return this
}

func (this *Meta) RecordLookup(ctx *core.Context, record interface{}) interface{} {
	if this.ProxyTo != nil {
		if len(this.proxyPath) > 0 {
			for _, p := range this.proxyPath {
				if pm, ok := p.(proxyPathMeta); ok {
					if record = pm.Meta.Value(ctx, record); record == nil {
						return record
					}
				}
			}
		}
	}
	return record
}

func (this *Meta) CallBeforeRenderHandlers(ctx *MetaContext, record interface{}) {
	if br, ok := this.Config.(MetaConfigBeforeRenderer1); ok {
		br.BeforeRender1(ctx, record)
	}
	for _, h := range this.beforeRenderHandlers {
		h(ctx, record)
	}
	if br, ok := this.Config.(MetaConfigBeforeRenderer2); ok {
		br.BeforeRender2(ctx, record)
	}
	if this.ProxyTo != nil {
		record = this.RecordLookup(ctx.Context.Context, record)
		this.ProxyTo.CallBeforeRenderHandlers(ctx, record)
	}
	if br, ok := this.Config.(MetaConfigBeforeRenderer3); ok {
		br.BeforeRender3(ctx, record)
	}
}

func (this *Meta) CallBeforeDoRenderHandler(ctx *MetaContext, record interface{}) {
	for _, h := range this.beforeDoRenderHandlers {
		h(ctx, record)
	}
	if this.ProxyTo != nil {
		record = this.RecordLookup(ctx.Context.Context, record)
		this.ProxyTo.CallBeforeDoRenderHandler(ctx, record)
	}
}

func (this *Meta) CallPrepareContextHandlers(ctx *MetaContext, record interface{}) {
	if br, ok := this.Config.(MetaConfigContextPreparer); ok {
		br.PrepareMetaContext(ctx, record)
	}
	for _, h := range this.prepareContextHandlers {
		h(ctx, record)
	}
	if this.ProxyTo != nil {
		record = this.RecordLookup(ctx.Context.Context, record)
		this.ProxyTo.CallPrepareContextHandlers(ctx, record)
	}
}

func (this *Meta) MakeError(ctx *core.Context, record interface{}, err error) error {
	return resource.ErrField(ctx, record, this.GetFieldName(), this.GetRecordLabelC(ctx, record))(err)
}

func (this *Meta) Sortable() bool {
	if this.SortHandler != nil {
		return true
	}
	return this.FieldStruct != nil && this.FieldStruct.IsNormal && this.FieldStruct.DBName != ""
}

func (this *Meta) Sort(s *Searcher, order aorm.Order) aorm.Clauser {
	if this.SortHandler != nil {
		return this.SortHandler(s, this, order)
	}
	return nil
}

func (this *Meta) GetRecordValueOrLoad(ctx *Context, record interface{}) (rec interface{}, err error) {
	if rec = this.Value(ctx.Context, record); rec != nil || this.ProxyTo != nil || this.FieldStruct == nil {
		return
	}

	if id := this.FieldStruct.Relationship.GetRelatedID(record); !id.IsZero() {
		rec = this.Resource.New()
		id.SetTo(rec)
		DB := ctx.DB().ModelStruct(this.Resource.ModelStruct).First(rec)
		if DB.Error == nil {
			this.BaseResource.FieldOf(record, this.Name).Set(rec)
		}
		return rec, DB.Error
	}
	return rec, nil
}

func (this *Meta) AsReadOnly() *Meta {
	this.ReadOnly = true
	this.ReadOnlyFunc = func(meta *Meta, ctx *Context, record interface{}) bool {
		return true
	}
	return this
}

func (this *Meta) OnUpdate(f func(m *Meta)) {
	this.updateCallbacks = append(this.updateCallbacks, f)
}

func (this *Meta) Permissioner(p ...Permissioner) {
	this.Permissioners = append(this.Permissioners, p...)
}

func (this *Meta) GetItemUrl(ctx *Context, record, item interface{}) string {
	if this.ItemUrlFunc != nil {
		return this.ItemUrlFunc(ctx, record, item)
	}
	return ctx.URLFor(item, this.Resource)
}

// GetItemUrlC get item url from item context
func (this *Meta) GetItemUrlC(ctx *Context) string {
	if this.ItemUrlFunc != nil {
		return this.ItemUrlFunc(ctx.Parent, ctx.Parent.ResourceRecord, ctx.ResourceRecord)
	}
	if this.Resource.ParentResource == ctx.Parent.Resource {
		var keys []aorm.ID
		var p = ctx.Parent
		for p != nil {
			if p.ResourceID != nil {
				keys = append(keys, p.ResourceID)
			}
			p = p.Parent
		}
		uri := this.Resource.URLFor(ctx.Parent, ctx.ResourceRecord, keys...)
		return ctx.URL(uri)
	}
	return ctx.URLFor(ctx.ResourceRecord, this.Resource)
}

func (this *Meta) SetRequired(v bool) *Meta {
	this.Required = v
	this.Meta.Required = v
	return this
}

func (this *Meta) ReadOnlyFormattedValue(ctx *Context, record interface{}) *FormattedValue {
	if this.ReadOnlyFormattedValuerFunc != nil {
		return this.Severitify(this.ReadOnlyFormattedValuerFunc(this, ctx, record))
	}
	return nil
}

func (this *Meta) AfterUpdate(f ...func()) {
	this.afterUpdate = append(this.afterUpdate, f...)
}

func (this *Meta) IsEnabled(record interface{}, context *Context, meta *Meta, readOnly bool) bool {
	if len(this.proxyPath) > 0 {
		rec := record
		for _, p := range this.proxyPath {
			if pm, ok := p.(proxyPathMeta); ok {
				if !pm.Meta.IsEnabled(rec, context, pm.Meta, readOnly) {
					return false
				}
				if rec != nil {
					rec = pm.Meta.Value(context.Context, rec)
				}
			}
		}
	}

	if this.GetRecordHandler != nil {
		record = this.GetRecordHandler(nil, record)
	}
	if this.Enabled != nil {
		return this.Enabled(record, context, meta, readOnly)
	}
	if readOnly && context.Type.Has(NEW) {
		return false
	}
	return true
}

func (this *Meta) IsReadOnly(ctx *Context, recorde interface{}) bool {
	var rec = recorde
	if this.ReadOnly {
		return true
	}
	if this.ReadOnlyFunc != nil {
		return this.ReadOnlyFunc(this, ctx, recorde)
	}
	if len(this.proxyPath) > 0 {
		for _, p := range this.proxyPath {
			if pm, ok := p.(proxyPathMeta); ok {
				if pm.Meta.IsReadOnly(ctx, rec) {
					return true
				}
				rec = pm.Meta.Value(ctx.Context, rec)
			}
		}
	}
	if this.ProxyTo != nil {
		if this.ProxyTo.ReadOnly {
			return true
		}
		if this.ProxyTo.ReadOnlyFunc != nil {
			return this.ProxyTo.IsReadOnly(ctx, rec)
		}
	}
	switch this.Type {
	case "single_edit", "collection_edit", "select_one", "select_many":
		return false
	}
	return this.GetSetter() == nil
}

func (this *Meta) IsReadOnlyItem(ctx *Context, recorde, item interface{}) bool {
	if this.ReadOnlyItemFunc != nil {
		return this.ReadOnlyItemFunc(this, ctx, recorde, item)
	}
	return false
}

func (this *Meta) RecordRequirer() func(ctx *core.Context, record interface{}) bool {
	if this.ReadOnlyFunc != nil {
		return func(ctx *core.Context, record interface{}) bool {
			return this.RecordRequired(ContextFromContext(ctx), record)
		}
	}
	return this.RecordRequirerFunc
}

func (this *Meta) RecordRequired(ctx *Context, record interface{}) bool {
	if this.ReadOnlyFunc != nil {
		if !this.IsReadOnly(ctx, record) {
			if this.RecordRequirerFunc != nil {
				return this.RecordRequirerFunc(ctx.Context, record)
			}
			return this.Required
		}
		return true
	}
	if this.RecordRequirerFunc != nil {
		return this.RecordRequirerFunc(ctx.Context, record)
	}
	return this.Required
}

func (this *Meta) I18nGroup(defaul ...bool) string {
	if len(defaul) > 0 && this.i18nGroup == "" {
		return I18NGROUP
	}
	return this.i18nGroup
}

func (this *Meta) SetI18nGroup(group string) *Meta {
	this.i18nGroup = group
	return this
}

func (this *Meta) TKey(key string) string {
	return this.I18nGroup(true) + ":meta." + this.Type + "." + key
}

func (this *Meta) Namer() *resource.MetaName {
	if name, ok := this.BaseResource.MetaAliases[this.Name]; ok {
		return name
	}
	return this.Meta.Namer()
}

func (this *Meta) NewSetter(f func(meta *Meta, old MetaSetter, recorde interface{}, metaValue *resource.MetaValue, ctx *core.Context) error) *Meta {
	old := this.Meta.Setter
	this.Setter = func(recorde interface{}, metaValue *resource.MetaValue, ctx *core.Context) error {
		return f(this, old, recorde, metaValue, ctx)
	}
	this.Meta.Setter = this.Setter
	return this
}

func (this *Meta) NewValuer(f func(meta *Meta, old MetaValuer, recorde interface{}, ctx *core.Context) interface{}) *Meta {
	old := this.Meta.Valuer
	this.Valuer = func(recorde interface{}, context *core.Context) interface{} {
		return f(this, old, recorde, context)
	}
	this.Meta.Valuer = this.Valuer
	return this
}

func (this *Meta) NewFormattedValuer(f func(meta *Meta, old MetaFValuer, recorde interface{}, ctx *core.Context) *FormattedValue) *Meta {
	old := this.Meta.FormattedValuer
	if old == nil {
		old = func(record interface{}, context *core.Context) *resource.FormattedValue {
			v := this.Value(context, record)
			if v == nil {
				return nil
			}
			return &FormattedValue{Record: record, Raw: v}
		}
	}
	this.FormattedValuer = func(record interface{}, ctx *core.Context) *FormattedValue {
		return f(this, old, record, ctx)
	}
	this.Meta.FormattedValuer = this.FormattedValuer
	return this
}

func (this *Meta) NewOutputFormattedValuer(f func(meta *Meta, old MetaOutputValuer, ctx *core.Context, record, value interface{}) interface{}) *Meta {
	old := this.OutputFormattedValuer
	this.OutputFormattedValuer = func(ctx *core.Context, recorde, value interface{}) interface{} {
		return f(this, old, ctx, recorde, value)
	}
	return this
}

func (this *Meta) SetValuer(f func(recorde interface{}, ctx *core.Context) interface{}) {
	this.Valuer = f
	this.Meta.SetValuer(f)
}

func (this *Meta) SetFormattedValuer(f MetaFValuer) {
	this.FormattedValuer = f
	this.Meta.SetFormattedValuer(f)
}

func (this *Meta) NewEnabled(f func(old MetaEnabled, record interface{}, ctx *Context, meta *Meta, readOnly bool) bool) *Meta {
	old := this.Enabled
	if old == nil {
		old = func(record interface{}, context *Context, meta *Meta, readOnly bool) bool {
			return true
		}
	}
	this.Enabled = func(recorde interface{}, ctx *Context, meta *Meta, readOnly bool) bool {
		return f(old, recorde, ctx, meta, readOnly)
	}
	return this
}

// Locked returns if this meta was locked for input changes
func (this *Meta) Locked(ctx *Context, record interface{}) bool {
	if this.LockedFunc != nil {
		return this.LockedFunc(this, ctx, record)
	}
	return false
}

func (this *Meta) GetType(record interface{}, context *Context, readOnly ...bool) string {
	if len(readOnly) > 0 && readOnly[0] && this.ReadOnlyType != "" {
		return this.ReadOnlyType
	}
	if this.GetRecordHandler != nil {
		record = this.GetRecordHandler(nil, record)
	}
	if this.TypeHandler != nil {
		return this.TypeHandler(this, record, context)
	}
	return this.Type
}

func (this *Meta) GetHelpPair() (key string, defaul string) {
	if this.Help == "-" {
		return "", ""
	}
	key = this.HelpKey

	if key == "" {
		key, _ = this.GetLabelPair()
		key += "_help"
		defaul = this.Help
	}

	if defaul == "" && this.Resource != nil {
		defaul = this.Resource.Tags.GetString("HELP")
	}

	return key, defaul
}

func (this *Meta) GetHelp(ctx *Context) (s string) {
	if key, defaul := this.GetHelpPair(); key != "" {
		s = ctx.Ts(key, defaul)
	} else {
		s = defaul
	}
	if s != "" {
		if this.HelpFormatter != nil {
			s = this.HelpFormatter(this, ctx, s)
		}
		return this.formatTemplateString(ctx, "help", s)
	}
	return
}

func (this *Meta) GetShowHelpPair() (key string, defaul string) {
	if this.ShowHelp == "-" {
		return "", ""
	}
	key = this.ShowHelpKey

	if key == "" {
		key, _ = this.GetLabelPair()
		key += "_show_help"
		defaul = this.ShowHelp
	}

	if defaul == "" && this.Resource != nil {
		defaul = this.Resource.Tags.GetString("RO_HELP")
	}

	return key, defaul
}

func (this *Meta) RecordHelpPair(f ...func(meta *Meta, ctx *Context, record interface{}) (key string, defaul string, ok bool)) *Meta {
	this.recordHelpPairFuncs = append(this.recordHelpPairFuncs, f...)
	return this
}

func (this *Meta) GetRecordHelpPair(ctx *Context, record interface{}) (key string, defaul string) {
	var ok bool
	for _, f := range this.recordHelpPairFuncs {
		if key, defaul, ok = f(this, ctx, record); ok {
			return
		}
	}
	return this.GetHelpPair()
}

func (this *Meta) GetRecordHelp(ctx *Context, record interface{}) (s string) {
	if key, defaul := this.GetRecordHelpPair(ctx, record); key != "" {
		s = ctx.Ts(key, defaul)
	} else {
		s = defaul
	}
	if s != "" {
		if this.RecordHelpFormatter != nil {
			s = this.RecordHelpFormatter(this, ctx, record, s)
		}
		return this.formatTemplateString(ctx, "record_help", s)
	}
	return
}

func (this *Meta) RecordShowHelpPair(f ...func(meta *Meta, ctx *Context, record interface{}) (key string, defaul string, ok bool)) *Meta {
	this.recordShowHelpPairFuncs = append(this.recordShowHelpPairFuncs, f...)
	return this
}

func (this *Meta) GetRecordShowHelpPair(ctx *Context, record interface{}) (key string, defaul string) {
	var ok bool
	for _, f := range this.recordShowHelpPairFuncs {
		if key, defaul, ok = f(this, ctx, record); ok {
			return
		}
	}
	return this.GetShowHelpPair()
}

func (this *Meta) GetRecordShowHelp(ctx *Context, record interface{}) (s string) {
	if key, defaul := this.GetRecordShowHelpPair(ctx, record); key != "" {
		s = ctx.Ts(key, defaul)
	} else {
		s = defaul
	}
	if s != "" {
		if this.RecordShowHelpFormatter != nil {
			s = this.RecordShowHelpFormatter(this, ctx, record, s)
		}
		return this.formatTemplateString(ctx, "record_show_help", s)
	}
	return
}

func (this *Meta) I18nKey(sub ...string) string {
	return this.GetLabelKey() + "_" + strings.Join(sub, ".")
}

func (this *Meta) TranslateLabel(ctx i18nmod.Context) string {
	key, defaul := this.GetLabelPair()
	return ctx.T(key).Default(defaul).Get()
}

func (this *Meta) GetLabelPair() (string, string) {
	name := this.Name

	if this.FieldLabel && this.FieldName != "" {
		name = this.FieldName
	}

	var (
		key    = name
		defaul = this.DefaultLabel
	)

	if this.Label != "" {
		key = this.Label
		if !strings.ContainsRune(key, '.') {
			defaul = this.Label
		}
	} else if !strings.ContainsRune(key, '.') {
		key = this.getLabelKey(key)
	}

	if this.SkipDefaultLabel {
		defaul = ""
	}

	return key, defaul
}

func (this *Meta) RecordLabelPair(f ...func(meta *Meta, ctx *Context, record interface{}) (key string, defaul string, ok bool)) *Meta {
	this.recordLabelPairFuncs = append(this.recordLabelPairFuncs, f...)
	return this
}

func (this *Meta) GetRecordLabelPair(ctx *Context, record interface{}) (key string, defaul string) {
	var ok bool
	for _, f := range this.recordLabelPairFuncs {
		if key, defaul, ok = f(this, ctx, record); ok {
			return
		}
	}
	return this.GetLabelPair()
}

func (this *Meta) GetRecordLabel(ctx *Context, record interface{}) (label string) {
	if key, defaul := this.GetRecordLabelPair(ctx, record); key != "" {
		label = ctx.Ts(key, defaul)
	} else {
		label = defaul
	}
	if this.RecordLabelFormatter != nil {
		label = this.RecordLabelFormatter(this, ctx, record, label)
	}
	return
}

func (this *Meta) GetRecordLabelC(ctx *core.Context, record interface{}) string {
	return this.GetRecordLabel(ContextFromCoreContext(ctx), record)
}

func (this *Meta) GetLabelC(ctx *core.Context) (s string) {
	if key, defaul := this.GetLabelPair(); key != "" {
		s = ctx.Ts(key, defaul)
	} else {
		s = defaul
	}
	if s != "" && this.LabelFormatter != nil {
		return this.LabelFormatter(this, ContextFromContext(ctx), s)
	}
	return
}

func (this *Meta) GetLabel(ctx *Context) string {
	return this.GetLabelC(ctx.Context)
}

func (this *Meta) getLabelKey(key string) string {
	if key == "" {
		key = this.Name

		if this.FieldLabel && this.FieldName != "" {
			key = this.FieldName
		}
	}

	return this.BaseResource.I18nPrefix + ".attributes." + key
}

func (this *Meta) GetLabelKey() string {
	return this.getLabelKey("")
}

func (this *Meta) URIFor(ctx *Context, record interface{}) string {
	if this.URIForFunc != nil {
		return this.URIForFunc(this, ctx, record)
	}
	return this.Resource.GetContextURI(ctx, this.Resource.GetKey(record))
}

func (this *Meta) URLFor(ctx *Context, record interface{}) string {
	if this.URLForFunc != nil {
		return this.URLForFunc(this, ctx, record)
	}
	return this.URIFor(ctx, record)
}

// metaConfig meta config
type metaConfig struct {
}

// GetTemplate get customized template for meta
func (metaConfig) GetTemplate(context *Context, metaType string) (*assetfs.AssetInterface, error) {
	return nil, errors.New("not implemented")
}

// MetaConfigInterface meta config interface
type MetaConfigInterface interface {
	resource.MetaConfigInterface
}

type MetaConfigContextPreparer interface {
	PrepareMetaContext(ctx *MetaContext, record interface{})
}

type MetaConfigBeforeRenderer1 interface {
	BeforeRender1(ctx *MetaContext, record interface{})
}

type MetaConfigBeforeRenderer2 interface {
	BeforeRender2(ctx *MetaContext, record interface{})
}

type MetaConfigBeforeRenderer3 interface {
	BeforeRender3(ctx *MetaContext, record interface{})
}

// GetMetas get sub metas
func (this *Meta) GetMetas() []resource.Metaor {
	if len(this.Metas) > 0 {
		return this.Metas
	} else if this.Resource == nil {
		return []resource.Metaor{}
	} else if this.GetMetasFunc != nil {
		return this.GetMetasFunc()
	} else {
		return this.Resource.GetMetas([]string{})
	}
}

func (this *Meta) GetContextMetas(recorde interface{}, context *core.Context) (result []resource.Metaor) {
	if this.ContextMetas != nil {
		metas := this.ContextMetas(recorde, GetContext(context))
		r := make([]resource.Metaor, len(metas))
		for i, m := range metas {
			r[i] = m
		}
		if this.Type == "single_edit" {
			if this.Typ.Kind() == reflect.Ptr {
				// TODO: add meta for '@enabled'
				// r = append(r, )
			}
		}
		return r
	}
	if this.ContextResourcer != nil {
		res := this.ContextResourcer(this, context)
		if res != nil {
			return res.GetMetas([]string{})
		}
	}
	if this.Resource != nil {
		return this.Resource.GetContextMetas(context)
	}
	return this.GetMetas()
}

// GetResourceByID get resource from meta
func (this *Meta) GetResource() resource.Resourcer {
	return this.Resource
}

// GetResourceByID get resource from meta
func (this *Meta) GetBaseResource() resource.Resourcer {
	return this.BaseResource
}

// GetContextResource get resource from meta
func (this *Meta) GetContextResourcer() func(meta resource.Metaor, context *core.Context) resource.Resourcer {
	if this.ContextResourcer != nil {
		return this.ContextResourcer
	}
	return this.Meta.ContextResourcer
}

func (this *Meta) GetContextResource(context *core.Context) resource.Resourcer {
	getter := this.GetContextResourcer()
	if getter != nil {
		return getter(this, context)
	}
	return this.GetResource()
}

// SetPermission set meta's permission
func (this *Meta) SetPermission(permission *roles.Permission) {
	this.Permission = permission
	this.Meta.Permission = permission
	if this.Resource != nil && this.Resource.metaEmbedded {
		this.Resource.Permission = permission
	}
}

func (this *Meta) HasPermission(mode roles.PermissionMode, context *core.Context) (perm roles.Perm) {
	return this.AdminHasPermission(mode, ContextFromCoreContext(context))
}

// AdminHasContextPermission check has permission or not
func (this *Meta) AdminHasPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	if perm = this.Meta.HasPermission(mode, context.Context); perm != roles.UNDEF {
		return
	}

	if mode == roles.Read && this.DefaultPermissionDeny() {
		return roles.DENY
	}

	if perm = this.adminHasPermission(mode, context); perm != roles.UNDEF {
		return
	}
	return
}

// AdminHasContextPermission check has permission or not
func (this *Meta) adminHasPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	if this.BaseResource != nil {
		return this.BaseResource.HasLocalPermission(mode, context)
	}
	return
}

func (this *Meta) TriggerValuerEvent(ename string, recorde interface{}, ctx *core.Context, valuer MetaValuer, value ...interface{}) (result interface{}, err error) {
	var v interface{}
	if len(value) > 0 {
		v = value[0]
	}
	e := &MetaValuerEvent{
		MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				this,
				this.BaseResource,
				ctx,
			},
			recorde,
		}, valuer, v, v, false}
	if err = this.Trigger(e); err != nil {
		return
	}
	if valuer != nil {
		if e.Value == nil && !e.originalValueCalled {
			return valuer(recorde, ctx), nil
		}
	}
	return e.Value, nil
}

func (this *Meta) MustTriggerValuerEvent(ename string, recorde interface{}, ctx *core.Context, valuer MetaValuer, value ...interface{}) (result interface{}) {
	if i, err := this.TriggerValuerEvent(ename, recorde, ctx, valuer, value...); err == nil {
		return i.(string)
	} else {
		panic(MetaEventError{this, ename, err})
	}
}

func (this *Meta) TriggerSetEvent(ename string, recorde interface{}, ctx *core.Context, setter MetaSetter, metaValue *resource.MetaValue) (err error) {
	e := &MetaSetEvent{
		MetaRecordeEvent: MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				this,
				this.BaseResource,
				ctx,
			},
			recorde,
		},
		Setter:    setter,
		MetaValue: metaValue,
	}
	if err = this.Trigger(e); err != nil {
		return
	}
	return setter(recorde, metaValue, ctx)
}

func (this *Meta) TriggerValueChangedEvent(ename string, recorde interface{}, ctx *core.Context, oldValue interface{}, valuer MetaValuer) error {
	e := &MetaValueChangedEvent{
		MetaRecordeEvent: MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				this,
				this.BaseResource,
				ctx,
			},
			recorde,
		},
		Old:    oldValue,
		valuer: valuer,
	}
	return this.Trigger(e)
}

// GetValuer get valuer from meta
func (this *Meta) GetValuer() func(interface{}, *core.Context) interface{} {
	var valuer func(interface{}, *core.Context) interface{}
	if this.ProxyTo != nil {
		valuer = func(i interface{}, context *core.Context) interface{} {
			if i = this.GetRecordHandler(context, i); i != nil {
				vlr := this.ProxyTo.GetValuer()
				return vlr(i, context)
			}
			return nil
		}
	} else {
		valuer = this.Meta.GetValuer()
	}
	if valuer != nil {
		if this.mustValuer {
			return valuer
		}
		return func(i interface{}, context *core.Context) (v interface{}) {
			var err error
			if v, err = this.TriggerValuerEvent(E_META_VALUE, i, context, valuer); err != nil {
				panic(MetaEventError{this, E_META_VALUE, err})
			}
			return
		}
	}
	return nil
}

func (this *Meta) FormatValue(value interface{}) string {
	if this.DefaultFormatF != nil {
		return this.DefaultFormatF(value)
	}
	if this.DefaultFormat != "" {
		return fmt.Sprintf(this.DefaultFormat, value)
	}
	return ""
}

// GetFormattedValuer get formatted valuer from meta
func (this *Meta) GetFormattedValuer() func(interface{}, *core.Context) *FormattedValue {
	var valuer MetaFValuer
	if this.FormattedValuer != nil {
		if this.ProxyTo != nil {
			valuer = func(record interface{}, context *core.Context) *FormattedValue {
				if record = this.GetRecordHandler(context, record); record == nil {
					return nil
				}
				if this.ProxyTo.FormattedValuer != nil {
					return this.ProxyTo.FormattedValuer(record, context)
				}
				return this.ProxyTo.FormattedValue(context, record)
			}
		} else {
			valuer = this.FormattedValuer
		}
	} else if this.ProxyTo != nil {
		valuer = func(r interface{}, ctx *core.Context) *FormattedValue {
			if r = this.GetRecordHandler(ctx, r); r != nil {
				vlr := this.ProxyTo.GetFormattedValuer()
				return vlr(r, ctx)
			}
			return nil
		}
	} else {
		valuer = func(record interface{}, context *core.Context) *FormattedValue {
			v := this.Value(context, record)
			if v == nil {
				return nil
			}
			if _, ok := v.(*resource.SliceValue); ok {
				return &FormattedValue{Record: record, Raw: v}
			}
			return &FormattedValue{Record: record, Raw: v, IsZeroF: this.IsZero}
		}
	}

	if this.mustValuer {
		return valuer
	}

	return func(record interface{}, context *core.Context) *FormattedValue {
		var (
			err error
			v   interface{}
		)
		if v, err = this.TriggerValuerEvent(E_META_FORMATTED_VALUE, record, context, func(record interface{}, context *core.Context) interface{} {
			return valuer(record, context)
		}); err != nil {
			panic(MetaEventError{this, E_META_FORMATTED_VALUE, err})
		}
		if v, err = this.TriggerValuerEvent(E_META_POST_FORMATTED_VALUE, record, context, nil, v); err != nil {
			panic(MetaEventError{this, E_META_POST_FORMATTED_VALUE, err})
		}
		if v == nil {
			return nil
		}
		if fv, ok := v.(*FormattedValue); ok {
			if fv == nil {
				return fv
			}
			if fv.Value == "" {
				fv.Value = this.FormatValue(fv.Raw)
			}
			return fv
		}
		return &FormattedValue{Record: record, Raw: v, IsZeroF: this.IsZero, Value: this.FormatValue(v)}
	}
}

// FormattedValue get formatted valuer from meta
func (this *Meta) Value(ctx *core.Context, record interface{}) interface{} {
	if this.Meta == nil {
		if this.Valuer != nil {
			return this.Valuer(record, ctx)
		}
	} else if valuer := this.GetValuer(); valuer != nil {
		return valuer(record, ctx)
	}
	return nil
}

// FormattedValue get formatted valuer from meta
func (this *Meta) FormattedValue(ctx *core.Context, record interface{}) (fv *FormattedValue) {
	if formattedValuer := this.GetFormattedValuer(); formattedValuer != nil {
		fv := formattedValuer(record, ctx)

		if fv == nil {
			return nil
		}

		var toString = func(raw interface{}) string {
			switch t := raw.(type) {
			case MetaFormattedValuer:
				return t.FormattedValue(ContextFromCoreContext(ctx), this, record)
			case MetaorFormattedValuer:
				return t.FormattedValue(ctx, this, record)
			case MetaSingleFormattedValuer:
				return t.FormattedValue()
			case Stringer:
				return t.AdminString(ContextFromCoreContext(ctx), maps.Map{})
			case core.ContextStringer:
				return t.ContextString(ctx)
			case fmt.Stringer:
				return t.String()
			default:
				return fmt.Sprint(fv.Raw)
			}
		}

		if fv.Raw != nil && (!fv.IsZero() || this.ForceShowZero) {
			if fv.Slice {
				if len(fv.Values) == 0 && len(fv.SafeValues) == 0 {
					var (
						slice = reflect.ValueOf(fv.Raws)
						l     = slice.Len()
					)
					fv.Values = make([]string, l)
					for i := 0; i < l; i++ {
						fv.Values[i] = toString(slice.Index(i).Interface())
					}
				}
				if fv.Value == "" {
					fv.Value = strings.Join(fv.Values, ", ")
				}
				if fv.SafeValue == "" {
					fv.SafeValue = strings.Join(fv.SafeValues, ", ")
				}
			} else {
				if fv.Value == "" && fv.SafeValue == "" {
					fv.Value = toString(fv.Raw)
				}
			}
		}
		return this.Severitify(fv)
	}
	return
}

func (this *Meta) GetFormattedValue(ctx *Context, record interface{}, ro bool) (v *FormattedValue) {
	if ro && this.Resource == nil {
		if v = this.ReadOnlyFormattedValue(ctx, record); v != nil {
			return
		}
	}
	return this.FormattedValue(ctx.Context, record)
}

// FormattedValue get formatted valuer from meta
func (this *Meta) GetDefaultValue(ctx *core.Context, recorde interface{}) (v interface{}) {
	var zero interface{}
	if this.DefaultValueFunc != nil {
		zero = this.DefaultValueFunc(recorde, ctx)
	} else if this.FieldStruct != nil {
		z := reflect.New(this.FieldStruct.Struct.Type).Elem()
		if this.FieldStruct.Struct.Type.Kind() == reflect.Struct {
			zero = z.Addr().Interface()
		} else {
			zero = z.Interface()
		}
	}
	var err error
	if v, err = this.TriggerValuerEvent(E_META_DEFAULT_VALUE, recorde, ctx, nil, zero); err != nil {
		panic(MetaEventError{this, E_META_DEFAULT_VALUE, err})
	}
	return
}

func (this *Meta) IsZeroInput(record interface{}, ctx *core.Context) (zero bool) {
	return this.IsZero(record, this.Value(ctx, record))
}

func (this *Meta) IsZero(record, value interface{}) (zero bool) {
	if this.Meta == nil {
		goto checkVal
	}
	if value == nil {
		if this.FieldStruct != nil && this.FieldStruct.Relationship != nil && indirectType(this.FieldStruct.Struct.Type).Kind() == reflect.Struct {
			return this.FieldStruct.Relationship.GetRelatedID(record).IsZero()
		}
		return true
	}
	if this.GetRecordHandler != nil {
		record = this.GetRecordHandler(nil, record)
		if record == nil {
			return true
		}
	}

checkVal:
	if this.NilAsZero {
		return value == nil
	}
	if this.IsZeroFunc != nil {
		return this.IsZeroFunc(this, record, value)
	}
	return aorm.IsZero(value)
}

// GetSetter get setter from meta
func (this *Meta) GetSetter() func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
	if this.Meta == nil {
		return nil
	}
	if setter := this.Meta.GetSetter(); setter != nil {
		return func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) (err error) {
			valuer := this.Meta.GetValuer()
			var old interface{}
			if valuer != nil {
				if old = valuer(recorde, context); old == nil && this.Typ != nil {
					typ := indirectType(this.Typ)
					switch this.Typ.Kind() {
					case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
					default:
						old = reflect.New(typ).Interface()
					}
				}
				if old != nil {
					value := reflect.ValueOf(old)
					if value.Kind() != reflect.Ptr {
						newValue := reflect.New(value.Type())
						newValue.Elem().Set(value)
						old = newValue.Interface()
					}
				}
			}
			if err = this.TriggerSetEvent(E_META_SET, recorde, context, setter, metaValue); err == nil {
				if err = this.TriggerValueChangedEvent(E_META_CHANGED, recorde, context, old, valuer); err != nil {
					err = MetaEventError{this, E_META_CHANGED, err}
				}
			} else {
				err = MetaEventError{this, E_META_SET, err}
			}
			return
		}
	}
	return nil
}

func (this *Meta) Set(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
	if setter := this.GetSetter(); setter != nil {
		return setter(recorde, metaValue, context)
	}
	return nil
}

func (this *Meta) Proxier() bool {
	return this.ProxyTo != nil
}

func (this *Meta) IsAlone() bool {
	return this.SectionNotAllowed
}

func (this *Meta) ID() string {
	return this.BaseResource.FullID() + "#" + this.Name
}

func (this *Meta) CanCreateItem(ctx *Context, record interface{}) bool {
	if this.CanCreateItemFunc != nil {
		return this.CanCreateItemFunc(ctx, record)
	}
	if this.Tags.Flag("ITEM_CREATE_DISABLED") {
		return false
	}
	return true
}

func (this *Meta) CanDeleteItem(ctx *Context, record, item interface{}) bool {
	if this.CanDeleteItemFunc != nil {
		return this.CanDeleteItemFunc(ctx, record, item)
	}
	if this.Tags.Flag("ITEM_DELETE_DISABLED") {
		return false
	}
	return true
}

func (this *Meta) CanUpdateItem(ctx *Context, record, item interface{}) bool {
	if this.CanUpdateItemFunc != nil {
		return this.CanUpdateItemFunc(ctx, record, item)
	}
	if this.Tags.Flag("ITEM_UPDATE_DISABLED") {
		return false
	}
	return true
}

func (this *Meta) GetReflectStructValueOrInstantiate(record reflect.Value) reflect.Value {
	if this.Resource != nil && this.Resource.Fragment != nil {
		f := record.Addr().Interface().(fragment.FragmentedModelInterface)
		frag := this.Resource.Fragment
		if frag.IsForm {
			v := f.GetFormFragment(frag.ID)
			if v == nil {
				v = this.Resource.New().(fragment.FormFragmentModelInterface)
				f.SetFormFragment(f, frag.ID, v)
			}
			return reflect.ValueOf(v).Elem()
		} else {
			v := f.GetFragment(frag.ID)
			if v == nil {
				v = this.Resource.New().(fragment.FragmentModelInterface)
				f.SetFragment(f, frag.ID, v)
			}
			return reflect.ValueOf(v).Elem()
		}
	}
	return this.Meta.GetReflectStructValueOrInstantiate(record)
}

type MetaEventError struct {
	Meta  *Meta
	Event string
	Err   error
}

func (this MetaEventError) Translate(ctx i18nmod.Context) string {
	var errMsg string
	if et, ok := this.Err.(i18nmod.Translater); ok {
		errMsg = et.Translate(ctx)
	} else {
		errMsg = this.Err.Error()
	}
	return fmt.Sprintf("%s: %s: %s",
		this.Meta.BaseResource.TranslateLabel(ctx),
		this.Meta.TranslateLabel(ctx),
		errMsg)
}

func (this MetaEventError) Error() string {
	return fmt.Sprintf("%s: %s [event %s]: %s", this.Meta.BaseResource.ID, this.Meta.Name, this.Event, this.Err)
}

func (this MetaEventError) Cause() error {
	return this.Err
}

type ReadOnlyMeta struct {
	*Meta
	ro bool
}

func (this *ReadOnlyMeta) CanReadOnly() bool {
	return this.ro
}

func (this *ReadOnlyMeta) Metaor() resource.Metaor {
	return this.Meta
}

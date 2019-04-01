package admin

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	errors2 "github.com/pkg/errors"

	"github.com/ecletus/core/helpers"

	"github.com/moisespsena/go-options"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-assetfs"
	"github.com/moisespsena/go-edis"
)

const (
	BASIC_LAYOUT                = "basic"
	BASIC_LAYOUT_WITH_ICON      = "basic_with_icon"
	BASIC_LAYOUT_HTML           = "basic_html"
	BASIC_LAYOUT_HTML_WITH_ICON = "basic_html_with_icon"
	BASIC_META_ID               = "_BasicID"
	BASIC_META_LABEL            = "_BasicLabel"
	BASIC_META_HTML             = "_BasicHTML"
	BASIC_META_ICON             = "_BasicIcon"
)

type DependencyParent struct {
	Meta *Meta
}

type DependencyPath struct {
	Meta *Meta
}

type DependencyQuery struct {
	Meta  *Meta
	Param string
}

type MetaOutputValuer func(context *core.Context, recorde, value interface{}) interface{}
type MetaValuer func(recorde interface{}, context *core.Context) interface{}
type MetaSetter func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error
type MetaEnabled func(recorde interface{}, context *Context, meta *Meta) bool

type MetaPostFormatted interface {
	MetaPostFormatted(meta *Meta, ctx *core.Context, recorde, value interface{}) interface{}
}

// Meta meta struct definition
type Meta struct {
	edis.EventDispatcher
	Name              string
	DB                *aorm.Alias
	Type              string
	TypeHandler       func(recorde interface{}, context *Context, meta *Meta) string
	Enabled           MetaEnabled
	DefaultLabel      string
	Label             string
	SkipDefaultLabel  bool
	FieldName         string
	FieldLabel        bool
	EncodedName       string
	Setter            MetaSetter
	Valuer            MetaValuer
	FormattedValuer   MetaValuer
	ContextResourcer  func(meta resource.Metaor, context *core.Context) resource.Resourcer
	ContextMetas      func(recorde interface{}, context *Context) []*Meta
	SkipResourceModel bool
	Resource          *Resource
	Permission        *roles.Permission
	Config            MetaConfigInterface

	Description    string
	DescriptionKey string

	Metas        []resource.Metaor
	GetMetasFunc func() []resource.Metaor
	Collection   interface{}
	*resource.Meta
	BaseResource          *Resource
	TemplateData          map[string]interface{}
	i18nGroup             string
	Dependency            []interface{}
	ProxyTo               *Meta
	Include               bool
	ForceShowZero         bool
	IsZeroFunc            func(recorde, value interface{}) bool
	Fragment              *Fragment
	Options               options.Options
	OutputFormattedValuer MetaOutputValuer
	DefaultValueFunc      MetaValuer
	proxyPath             []ProxyPath
	Virtual               bool

	SectionNotAllowed bool
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

func (meta *Meta) I18nGroup(defaul ...bool) string {
	if len(defaul) > 0 && meta.i18nGroup == "" {
		return I18NGROUP
	}
	return meta.i18nGroup
}

func (meta *Meta) SetI18nGroup(group string) *Meta {
	meta.i18nGroup = group
	return meta
}

func (meta *Meta) TKey(key string) string {
	return meta.I18nGroup(true) + ":meta." + meta.Type + "." + key
}

func (meta *Meta) Namer() *resource.MetaName {
	if name, ok := meta.BaseResource.MetaAliases[meta.Name]; ok {
		return name
	}
	return meta.Meta.Namer()
}

func (meta *Meta) NewSetter(f func(meta *Meta, old MetaSetter, recorde interface{}, metaValue *resource.MetaValue, ctx *core.Context) error) *Meta {
	old := meta.Setter
	meta.Setter = func(recorde interface{}, metaValue *resource.MetaValue, ctx *core.Context) error {
		return f(meta, old, recorde, metaValue, ctx)
	}
	return meta
}

func (meta *Meta) NewValuer(f func(meta *Meta, old MetaValuer, recorde interface{}, ctx *core.Context) interface{}) *Meta {
	old := meta.Valuer
	meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
		return f(meta, old, recorde, context)
	}
	return meta
}

func (meta *Meta) NewFormattedValuer(f func(meta *Meta, old MetaValuer, recorde interface{}, ctx *core.Context) interface{}) *Meta {
	old := meta.FormattedValuer
	meta.FormattedValuer = func(recorde interface{}, ctx *core.Context) interface{} {
		return f(meta, old, recorde, ctx)
	}
	return meta
}

func (meta *Meta) NewOutputFormattedValuer(f func(meta *Meta, old MetaOutputValuer, ctx *core.Context, recorde, value interface{}) interface{}) *Meta {
	old := meta.OutputFormattedValuer
	meta.OutputFormattedValuer = func(ctx *core.Context, recorde, value interface{}) interface{} {
		return f(meta, old, ctx, recorde, value)
	}
	return meta
}

func (meta *Meta) SetValuer(f func(recorde interface{}, ctx *core.Context) interface{}) {
	meta.Valuer = f
	meta.Meta.SetValuer(f)
}

func (meta *Meta) SetFormattedValuer(f func(recorde interface{}, ctx *core.Context) interface{}) {
	meta.FormattedValuer = f
	meta.Meta.SetFormattedValuer(f)
}

func (meta *Meta) NewEnabled(f func(old MetaEnabled, recorde interface{}, ctx *Context, meta *Meta) bool) *Meta {
	old := meta.Enabled
	meta.Enabled = func(recorde interface{}, ctx *Context, meta *Meta) bool {
		return f(old, recorde, ctx, meta)
	}
	return meta
}

func (meta *Meta) GetType(record interface{}, context *Context) string {
	if meta.TypeHandler != nil {
		return meta.TypeHandler(record, context, meta)
	}
	return meta.Type
}

func (meta *Meta) GetDescriptionPair() (string, string) {
	var (
		key    = meta.DescriptionKey
		defaul string
	)

	if key == "" && meta.Description != "" {
		if meta.Description == "@" {
			key, _ = meta.GetLabelPair()
			key += "_description"
		} else if strings.ContainsRune(key, '.') {
			key = meta.Description
		} else {
			defaul = meta.Description
		}
	}

	return key, defaul
}

func (meta *Meta) GetLabelPair() (string, string) {
	name := meta.Name

	if meta.FieldLabel && meta.FieldName != "" {
		name = meta.FieldName
	}

	var (
		key    = name
		defaul = meta.DefaultLabel
	)

	if meta.Label != "" {
		key = meta.Label
		if !strings.ContainsRune(key, '.') {
			defaul = meta.Label
		}
	} else if !strings.ContainsRune(key, '.') {
		key = meta.BaseResource.I18nPrefix + ".attributes." + key
	}

	if meta.SkipDefaultLabel {
		defaul = ""
	}

	return key, defaul
}

func (meta *Meta) GetLabelKey() (string) {
	key := meta.Name

	if meta.FieldLabel && meta.FieldName != "" {
		key = meta.FieldName
	}

	return meta.BaseResource.I18nPrefix + ".attributes." + key
}

// metaConfig meta config
type metaConfig struct {
}

// GetTemplate get customized template for meta
func (metaConfig) GetTemplate(context *Context, metaType string) (*assetfs.Asset, error) {
	return nil, errors.New("not implemented")
}

// MetaConfigInterface meta config interface
type MetaConfigInterface interface {
	resource.MetaConfigInterface
}

// GetMetas get sub metas
func (meta *Meta) GetMetas() []resource.Metaor {
	if len(meta.Metas) > 0 {
		return meta.Metas
	} else if meta.Resource == nil {
		return []resource.Metaor{}
	} else if meta.GetMetasFunc != nil {
		return meta.GetMetasFunc()
	} else {
		return meta.Resource.GetMetas([]string{})
	}
}

func (meta *Meta) GetContextMetas(recorde interface{}, context *core.Context) []resource.Metaor {
	if meta.ContextMetas != nil {
		metas := meta.ContextMetas(recorde, context.Data().Get(CONTEXT_KEY).(*Context))
		r := make([]resource.Metaor, len(metas))
		for i, m := range metas {
			r[i] = m
		}
		return r
	}
	if meta.ContextResourcer != nil {
		res := meta.ContextResourcer(meta, context)
		if res != nil {
			return res.GetMetas([]string{})
		}
	}
	return meta.GetMetas()
}

// GetResourceByID get resource from meta
func (meta *Meta) GetResource() resource.Resourcer {
	if meta.Resource == nil {
		return nil
	}
	return meta.Resource
}

// GetContextResource get resource from meta
func (meta *Meta) GetContextResourcer() func(meta resource.Metaor, context *core.Context) resource.Resourcer {
	if meta.ContextResourcer != nil {
		return meta.ContextResourcer
	}
	return meta.Meta.ContextResourcer
}

func (meta *Meta) GetContextResource(context *core.Context) resource.Resourcer {
	getter := meta.GetContextResourcer()
	if getter != nil {
		return getter(meta, context)
	}
	return meta.GetResource()
}

// DBName get meta's db name
func (meta *Meta) DBName() string {
	if meta.FieldStruct != nil {
		return meta.FieldStruct.DBName
	}
	return ""
}

// SetPermission set meta's permission
func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
	meta.Meta.Permission = permission
	if meta.Resource != nil {
		meta.Resource.Permission = permission
	}
}

// HasPermission check has permission or not
func (meta Meta) HasPermissionE(mode roles.PermissionMode, context *core.Context) (ok bool, err error) {
	var roles_ = []interface{}{}
	for _, role := range context.Roles {
		roles_ = append(roles_, role)
	}
	if meta.Permission != nil && !roles.HasPermission(meta.Permission, mode, roles_...) {
		return false, nil
	}

	if meta.BaseResource != nil {
		return core.HasPermissionDefaultE(true, meta.BaseResource, mode, context)
	}

	return true, roles.ErrDefaultPermission
}

func (meta *Meta) TriggerValuerEvent(ename string, recorde interface{}, ctx *core.Context, valuer MetaValuer, value ...interface{}) interface{} {
	var v interface{}
	if len(value) > 0 {
		v = value[0]
	}
	e := &MetaValuerEvent{
		MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				meta,
				meta.BaseResource,
				ctx,
			},
			recorde,
		}, valuer, v, v, false}
	meta.Trigger(e)
	if valuer != nil {
		if e.Value == nil && !e.originalValueCalled {
			return valuer(recorde, ctx)
		}
	}
	return e.Value
}

func (meta *Meta) TriggerSetEvent(ename string, recorde interface{}, ctx *core.Context, setter MetaSetter, metaValue *resource.MetaValue) error {
	e := &MetaSetEvent{
		MetaRecordeEvent: MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				meta,
				meta.BaseResource,
				ctx,
			},
			recorde,
		},
		Setter:    setter,
		MetaValue: metaValue,
	}
	meta.Trigger(e)
	return setter(recorde, metaValue, ctx)
}

func (meta *Meta) TriggerValueChangedEvent(ename string, recorde interface{}, ctx *core.Context, oldValue interface{}, valuer MetaValuer) error {
	e := &MetaValueChangedEvent{
		MetaRecordeEvent: MetaRecordeEvent{
			MetaEvent{
				edis.NewEvent(ename),
				meta,
				meta.BaseResource,
				ctx,
			},
			recorde,
		},
		Old:    oldValue,
		valuer: valuer,
	}
	return meta.Trigger(e)
}

// GetValuer get valuer from meta
func (meta *Meta) GetValuer() func(interface{}, *core.Context) interface{} {
	if valuer := meta.Meta.GetValuer(); valuer != nil {
		return func(i interface{}, context *core.Context) interface{} {
			return meta.TriggerValuerEvent(E_META_VALUE, i, context, valuer)
		}
	}
	return nil
}

// GetFormattedValuer get formatted valuer from meta
func (meta *Meta) GetFormattedValuer() func(interface{}, *core.Context) interface{} {
	var valuer MetaValuer
	if meta.FormattedValuer != nil {
		valuer = meta.FormattedValuer
	} else {
		valuer = meta.GetValuer()
	}
	return func(i interface{}, context *core.Context) interface{} {
		v := meta.TriggerValuerEvent(E_META_FORMATTED_VALUE, i, context, valuer)
		v = meta.TriggerValuerEvent(E_META_POST_FORMATTED_VALUE, i, context, nil, v)
		return v
	}
}

// FormattedValue get formatted valuer from meta
func (meta *Meta) Value(ctx *core.Context, recorde interface{}) interface{} {
	if valuer := meta.GetValuer(); valuer != nil {
		return valuer(recorde, ctx)
	}
	return nil
}

// FormattedValue get formatted valuer from meta
func (meta *Meta) FormattedValue(ctx *core.Context, recorde interface{}) interface{} {
	if formattedValuer := meta.GetFormattedValuer(); formattedValuer != nil {
		return formattedValuer(recorde, ctx)
	}
	return ""
}

// FormattedValue get formatted valuer from meta
func (meta *Meta) GetDefaultValue(ctx *core.Context, recorde interface{}) interface{} {
	var zero interface{}
	if meta.DefaultValueFunc != nil {
		zero = meta.DefaultValueFunc(recorde, ctx)
	} else if meta.FieldStruct != nil {
		z := reflect.New(meta.FieldStruct.Struct.Type).Elem()
		if meta.FieldStruct.Struct.Type.Kind() == reflect.Struct {
			zero = z.Addr().Interface()
		} else {
			zero = z.Interface()
		}
	}
	return meta.TriggerValuerEvent(E_META_DEFAULT_VALUE, recorde, ctx, nil, zero)
}

func (meta *Meta) updateMeta() {
	if meta.Meta == nil {
		meta.Meta = &resource.Meta{
			MetaName:         &resource.MetaName{meta.Name, meta.EncodedName},
			FieldName:        meta.FieldName,
			Setter:           meta.Setter,
			Valuer:           meta.Valuer,
			FormattedValuer:  meta.FormattedValuer,
			BaseResource:     meta.BaseResource,
			ContextResourcer: meta.ContextResourcer,
			Resource:         meta.Resource,
			Permission:       meta.Permission,
			Config:           meta.Config,
		}
	} else {
		meta.Meta.Alias = meta.Alias
		meta.Meta.Name = meta.Name
		meta.Meta.FieldName = meta.FieldName
		meta.Meta.EncodedName = meta.EncodedName
		meta.Meta.Setter = meta.Setter
		meta.Meta.Valuer = meta.Valuer
		meta.Meta.FormattedValuer = meta.FormattedValuer
		meta.Meta.BaseResource = meta.BaseResource
		meta.Meta.Resource = meta.Resource
		meta.Meta.Permission = meta.Permission
		meta.Meta.Config = meta.Config
		meta.Meta.ContextResourcer = meta.ContextResourcer
	}

	if meta.Options == nil {
		meta.Options = make(options.Options)
	}

	if meta.EventDispatcher.GetDefinedDispatcher() == nil {
		meta.EventDispatcher.SetDispatcher(meta)
	}

	meta.PreInitialize()

	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(meta)
		}
	}

	if meta.Virtual {
		if meta.Valuer == nil {
			meta.Valuer = func(i interface{}, context *core.Context) interface{} {
				return nil
			}
		}
		if meta.Setter == nil {
			meta.Setter = func(interface{}, *resource.MetaValue, *core.Context) error {
				return nil
			}
		}
	}

	meta.Initialize()

	if meta.Label != "" && meta.DefaultLabel == "" && !strings.ContainsRune(meta.Label, '.') {
		meta.DefaultLabel = meta.Label
	} else if meta.DefaultLabel == "" {
		meta.DefaultLabel = utils.HumanizeString(meta.Name)
	}

	var fieldType reflect.Type
	var hasColumn = meta.FieldStruct != nil
	var isPtr bool

	if hasColumn {
		fieldType = meta.FieldStruct.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			isPtr = true
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Type
	if hasColumn {
		if meta.Type == "" {
			if _, ok := reflect.New(fieldType).Interface().(sql.Scanner); ok {
				if fieldType.Kind() == reflect.Struct {
					fieldType = reflect.Indirect(reflect.New(fieldType)).Field(0).Type()
				}
			}

			if relationship := meta.FieldStruct.Relationship; relationship != nil {
				if relationship.Kind == "has_one" {
					meta.Type = "single_edit"
				} else if relationship.Kind == "has_many" {
					meta.Type = "collection_edit"
				} else if relationship.Kind == "belongs_to" {
					meta.Type = "select_one"
				} else if relationship.Kind == "many_to_many" {
					meta.Type = "select_many"
				}
			} else {
				switch fieldType.Kind() {
				case reflect.String:
					var tags = meta.FieldStruct.TagSettings
					if size, ok := tags["SIZE"]; ok {
						if i, _ := strconv.Atoi(size); i > 255 {
							meta.Type = "text"
						} else {
							meta.Type = "string"
						}
					} else if text, ok := tags["TYPE"]; ok && text == "text" {
						meta.Type = "text"
					} else {
						meta.Type = "string"
					}
				case reflect.Bool:
					if isPtr {
						meta.Type = "select_one"
						meta.Config = &SelectOneConfig{
							AllowBlank: true,
							Collection: func(ctx *core.Context) [][]string {
								p := I18NGROUP + ".form.bool."
								return [][]string{
									{"true", ctx.Ts(p+"true", "Yes")},
									{"false", ctx.Ts(p+"false", "No")},
								}
							},
						}
						meta.SetFormattedValuer(func(recorde interface{}, ctx *core.Context) interface{} {
							value := meta.Value(ctx, recorde)
							if value == nil {
								return ""
							}
							p := I18NGROUP + ".form.bool."
							b := value.(*bool)
							if b == nil {
								return ""
							}
							if *b {
								return ctx.Ts(p+"true", "Yes")
							}
							return ctx.Ts(p+"false", "No")
						})
						meta.IsZeroFunc = func(recorde, value interface{}) bool {
							if value == nil {
								return true
							}
							return false
						}
					} else {
						meta.Type = "switch"
					}
				default:
					if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(fieldType.Kind().String()) {
						meta.Type = "number"
					} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(fieldType.Kind().String()) {
						meta.Type = "float"
					} else if _, ok := reflect.New(fieldType).Interface().(*time.Time); ok {
						meta.Type = "datetime"
					} else {
						if fieldType.Kind() == reflect.Struct {
							meta.Type = "single_edit"
						} else if fieldType.Kind() == reflect.Slice {
							refelectType := fieldType.Elem()
							for refelectType.Kind() == reflect.Ptr {
								refelectType = refelectType.Elem()
							}
							if refelectType.Kind() == reflect.Struct {
								meta.Type = "collection_edit"
							}
						}
					}
				}
			}
		} else {
			if relationship := meta.FieldStruct.Relationship; relationship != nil {
				if (relationship.Kind == "has_one" || relationship.Kind == "has_many") && meta.Meta.Setter == nil && (meta.Type == "select_one" || meta.Type == "select_many") {
					meta.SetSetter(func(resource interface{}, metaValue *resource.MetaValue, context *core.Context) error {
						scope := &aorm.Scope{Value: resource}
						reflectValue := reflect.Indirect(reflect.ValueOf(resource))
						field := reflectValue.FieldByName(meta.FieldName)

						if field.Kind() == reflect.Ptr {
							if field.IsNil() {
								field.Set(utils.NewValue(field.Type()).Elem())
							}

							for field.Kind() == reflect.Ptr {
								field = field.Elem()
							}
						}

						primaryKeys := utils.ToArray(metaValue.Value)
						if len(primaryKeys) > 0 {
							// set current field value to blank and replace it with new value
							field.Set(reflect.Zero(field.Type()))
							context.DB.Where(primaryKeys).Find(field.Addr().Interface())
						}

						if !scope.PrimaryKeyZero() {
							context.DB.Model(resource).Association(meta.FieldName).Replace(field.Interface())
							field.Set(reflect.Zero(field.Type()))
						}
						return nil
					})
				}
			}
		}
	}

	{ // Set Meta Resource
		if hasColumn {
			if meta.Resource == nil {
				var result interface{}

				if fieldType.Kind() == reflect.Struct {
					result = reflect.New(fieldType).Interface()
				} else if fieldType.Kind() == reflect.Slice {
					refelectType := fieldType.Elem()
					for refelectType.Kind() == reflect.Ptr {
						refelectType = refelectType.Elem()
					}
					if refelectType.Kind() == reflect.Struct {
						result = reflect.New(refelectType).Interface()
					}
				}

				if result != nil {
					res := meta.BaseResource.NewResource(&SubConfig{FieldName: meta.FieldStruct.Name}, result)
					meta.Resource = res
					meta.Meta.Permission = meta.Meta.Permission.Concat(res.Config.Permission)
				}
			} else if meta.Config == nil && meta.Resource.mounted {
				switch meta.Type {
				case "select_one", "select_many":
					cfg := &SelectOneConfig{RemoteDataResource: &DataResource{}}
					cfg.Layout = BASIC_LAYOUT
					meta.Config = cfg
				}
			}

			if meta.Resource != nil && meta.Resource != meta.BaseResource {
				permission := meta.Resource.Permission.Concat(meta.Meta.Permission)
				meta.Meta.Resource = meta.Resource
				meta.SetPermission(permission)
			}
		}
	}

	meta.FieldName = meta.GetFieldName()

	if meta.BaseResource.SingleEditMetas == nil {
		meta.BaseResource.SingleEditMetas = make(map[string]*Meta)
	}

	if _, ok := meta.BaseResource.SingleEditMetas[meta.Name]; ok {
		if meta.Type != "single_edit" {
			delete(meta.BaseResource.SingleEditMetas, meta.Name)
			meta.Inline = false
		}
	} else if meta.Type == "single_edit" {
		meta.BaseResource.SingleEditMetas[meta.Name] = meta
		meta.Inline = true
	}

	// call meta config's ConfigureMetaInterface
	if meta.Config != nil {
		meta.Config.ConfigureQorMeta(meta)
	}

	// call field's ConfigureMetaInterface
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaInterface); ok {
			injector.ConfigureQorMeta(meta)
		}
	}

	// run meta configors
	if baseResource := meta.BaseResource; baseResource != nil {
		for key, fc := range baseResource.GetAdmin().metaConfigorMaps {
			if key == meta.Type {
				fc(meta)
			}
		}
	}
}

func (meta *Meta) IsZero(recorde, value interface{}) bool {
	if value == nil {
		return true
	}
	if meta.IsZeroFunc != nil {
		return meta.IsZeroFunc(recorde, value)
	}
	switch vt := value.(type) {
	case helpers.Zeroer:
		return vt.IsZero()
	case string:
		if vt == "" {
			return true
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if vt == 0 {
			return true
		}
	case float32:
		if vt == 0.0 {
			return true
		}
	case float64:
		if vt == 0.0 {
			return true
		}
	case time.Time:
		return vt.IsZero()
	case *time.Time:
		return vt == nil || vt.IsZero()
	}
	return false
}

// GetSetter get setter from meta
func (meta *Meta) GetSetter() func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
	if setter := meta.Meta.GetSetter(); setter != nil {
		return func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) (err error) {
			valuer := meta.Meta.GetValuer()
			var old interface{}
			if valuer != nil {
				old = valuer(recorde, context)
				if old != nil {
					value := reflect.ValueOf(old)
					if value.Kind() != reflect.Ptr {
						newValue := reflect.New(value.Type())
						newValue.Elem().Set(value)
						old = newValue.Interface()
					}
				}
			}
			if err = meta.TriggerSetEvent(E_META_SET, recorde, context, setter, metaValue); err == nil {
				err = errors2.Wrap(meta.TriggerValueChangedEvent(E_META_CHANGED, recorde, context, old, valuer), "Trigger POST_SET")
			}
			return
		}
	}
	return nil
}

func (meta *Meta) Set(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
	if setter := meta.GetSetter(); setter != nil {
		return setter(recorde, metaValue, context)
	}
	return nil
}

package admin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
)

func (this *Resource) GetDefinedMeta(name string) *Meta {
	meta := this.MetasByName[name]
	if meta == nil {
		meta = this.MetasByFieldName[name]
	}
	return meta
}

func (this *Resource) GetMetaOrSet(name string) (meta *Meta) {
	if meta = this.GetDefinedMeta(name); meta == nil {
		meta = this.Meta(&Meta{Name: name})
	}
	return
}

// GetMeta get meta with name
func (this *Resource) GetMeta(name string, notUpdate ...bool) *Meta {
	return this.getMeta(&Meta{Name: name}, notUpdate...)
}

func (this *Resource) getMeta(meta *Meta, notUpdate ...bool) *Meta {
	fallbackMeta := this.MetasByName[meta.Name]

	if meta.Type == "-" {
		meta.Enabled = func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}
		meta.Type = ""
	}

	if fallbackMeta == nil {
		if meta.Name[0] == '@' {
			// meta for getter function
			if meta.Valuer == nil {
				if method, ok := this.ModelStruct.Type.MethodByName(meta.Name[1:]); ok {
					meta.Typ = method.Type.Out(0)
					if method.Type.NumOut() != 1 {
						log.Fatalf("meta method %q getter: expected 1 output values", meta.Name[1:])
					}
					switch method.Type.NumIn() {
					case 1:
						meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
							return reflect.Indirect(reflect.ValueOf(recorde)).Method(method.Index).Call(nil)[0].Interface()
						}
					case 2:
						if expected, got := reflect.TypeOf(&core.Context{}), method.Type.In(0); expected != got {
							log.Fatalf("meta method %q getter: expected %s argument type, but got %s", meta.Name[1:], expected, got)
						}
						meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
							return reflect.Indirect(reflect.ValueOf(recorde)).Method(method.Index).Call([]reflect.Value{reflect.ValueOf(context)})[0].Interface()
						}
					default:
						log.Fatalf("meta method %q getter: expected 1 or 2 input arguments", meta.Name[1:])
					}
				}
			}

			if meta.Setter == nil {
				methodName := "Set" + meta.Name[1:] + "MetaValue"
				if method, ok := this.ModelStruct.Type.MethodByName(methodName); ok {
					meta.Typ = method.Type.Out(0)
					if method.Type.NumOut() != 1 || method.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
						log.Fatalf("meta method %q: expected error output", methodName)
					}
					switch method.Type.NumIn() {
					case 1:
						if expected, got := reflect.TypeOf(&resource.MetaValue{}), method.Type.In(0); expected != got {
							log.Fatalf("meta method %q setter: expected %s argument type, but got %s", methodName, expected, got)
						}
						meta.Setter = func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
							return reflect.Indirect(reflect.ValueOf(recorde)).Method(method.Index).Call([]reflect.Value{
								reflect.ValueOf(metaValue),
							})[0].Interface().(error)
						}
					case 2:
						if expected, got := reflect.TypeOf(&core.Context{}), method.Type.In(0); expected != got {
							log.Fatalf("meta method %q setter: expected %s first argument type, but got %s", methodName, expected, got)
						}
						if expected, got := reflect.TypeOf(&resource.MetaValue{}), method.Type.In(1); expected != got {
							log.Fatalf("meta method %q setter: expected %s second argument type, but got %s", methodName, expected, got)
						}
						meta.Setter = func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
							return reflect.Indirect(reflect.ValueOf(recorde)).Method(method.Index).Call([]reflect.Value{
								reflect.ValueOf(context),
								reflect.ValueOf(metaValue),
							})[0].Interface().(error)
						}
					default:
						log.Fatalf("meta method %q setter: expected 1 or 2 input arguments", methodName)
					}
				}
			}
		} else if field, ok := this.ModelStruct.FieldsByName[meta.Name]; ok {
			if meta.BaseResource == nil {
				meta.BaseResource = this
			}
			if field.IsPrimaryKey && meta.Type == "" {
				meta.Type = "hidden_primary_key"
			}
			this.MetasByName[meta.Name] = meta
			this.MetasByFieldName[meta.Name] = meta
			this.Metas = append(this.Metas, meta)
			meta.updateMeta()
			return meta
		} else if field := this.ModelStruct.GetVirtualField(meta.Name); field != nil {
			if meta.BaseResource == nil {
				meta.BaseResource = this
			}
			this.MetasByName[meta.Name] = meta
			this.MetasByFieldName[meta.Name] = meta
			this.Metas = append(this.Metas, meta)
			if meta.Valuer == nil {
				meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
					if value, ok := field.Get(recorde); ok {
						return value
					}
					return nil
				}
			}
			if meta.Setter == nil {
				meta.Setter = func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
					field.Set(recorde, metaValue.Value)
					return nil
				}
			}
			meta.updateMeta()
			return meta
		} else if meta.Name == META_STRINGIFY {
			if meta.Label == "" {
				meta.Label = this.SingularLabelKey()
			}
			if meta.Type == "" {
				meta.Type = "string"
			}
			if meta.Valuer == nil {
				meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
					return utils.StringifyContext(recorde, context)
				}
			}
			this.MetasByName[meta.Name] = meta
			meta.BaseResource = this
			meta.updateMeta()
			return meta
		} else {
			parts := strings.Split(meta.Name, ".")
			if len(parts) > 1 {
				r := this
				var pth []interface{}
				for _, p := range parts[0 : len(parts)-1] {
					if r.Fragments != nil && r.Fragments.Get(p) != nil {
						r = r.Fragments.Get(p).Resource
						pth = append(pth, ProxyVirtualFieldPath{r.Fragment.ID, r.Value})
					} else if meta := r.GetMeta(p); meta != nil {
						if meta.Resource != nil {
							r = meta.Resource
						}
						pth = append(pth, ProxyMetaPath{meta})
					}
				}

				if pth != nil {
					to := r.GetMeta(parts[len(parts)-1])
					if to == nil {
						panic(fmt.Errorf("meta %q: destination does not exists", meta.Name))
					}
					meta = NewMetaFieldProxy(meta.Name, pth, this.Value, to)
					this.MetasByName[meta.Name] = meta
					this.Metas = append(this.Metas, meta)
					meta.updateMeta()
					return meta
				}

				return nil
			}
		}
	}

	return fallbackMeta
}

// Meta register meta for admin resource
func (this *Resource) SetMeta(meta *Meta, notUpdate ...bool) *Meta {
	return this.Meta(meta, true)
}

// MetaDisable disable metas by name
func (this *Resource) MetaDisable(names ...string) {
	for _, name := range names {
		this.Meta(&Meta{Name: name, Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}})
	}
}

// MetaRequired set metas as required
func (this *Resource) MetaRequired(names ...string) {
	for _, name := range names {
		this.Meta(&Meta{Name: name, Required: true})
	}
}

// MetaOptional set metas to optional
func (this *Resource) MetaOptional(names ...string) {
	for _, name := range names {
		m := this.Meta(&Meta{Name: name, Required: false})
		m.Meta.Required = false
		m.updateMeta()
	}
}

// MetaR register meta for admin resource and return this resource
func (this *Resource) MetaR(meta *Meta, notUpdate ...bool) *Resource {
	this.Meta(meta, notUpdate...)
	return this
}

// Meta register meta for admin resource
func (this *Resource) Meta(meta *Meta, notUpdate ...bool) *Meta {
	if oldMeta := this.getMeta(meta, notUpdate...); oldMeta != nil {
		if meta != oldMeta {
			if meta.Type != "" {
				oldMeta.Type = meta.Type
				oldMeta.Config = nil
			}

			if meta.TypeHandler != nil {
				oldMeta.TypeHandler = meta.TypeHandler
			}

			if meta.Enabled != nil {
				oldMeta.Enabled = meta.Enabled
			}

			if meta.SkipDefaultLabel {
				oldMeta.SkipDefaultLabel = true
			}

			if meta.DefaultLabel != "" {
				oldMeta.DefaultLabel = meta.DefaultLabel
			}

			if meta.Label != "" {
				oldMeta.Label = meta.Label
			}

			if meta.FieldName != "" {
				oldMeta.FieldName = meta.FieldName
			}

			if meta.Setter != nil {
				oldMeta.Setter = meta.Setter
			}

			if meta.Valuer != nil {
				oldMeta.Valuer = meta.Valuer
			}

			if meta.FormattedValuer != nil {
				oldMeta.FormattedValuer = meta.FormattedValuer
			}

			if meta.Resource != nil {
				oldMeta.Resource = meta.Resource
			}

			if meta.Permission != nil {
				oldMeta.Permission = meta.Permission
			}

			if meta.Config != nil {
				oldMeta.Config = meta.Config
			}

			if meta.Collection != nil {
				oldMeta.Collection = meta.Collection
			}

			if len(meta.Dependency) > 0 {
				oldMeta.Dependency = meta.Dependency
			}

			if meta.Fragment != nil {
				oldMeta.Fragment = meta.Fragment
			}

			if meta.Options != nil {
				oldMeta.Options.Update(meta.Options)
			}

			if meta.ReadOnlyFunc != nil {
				oldMeta.ReadOnlyFunc = meta.ReadOnlyFunc
			}

			oldMeta.updateMeta()
			meta = oldMeta
		}
	} else {
		this.MetasByName[meta.Name] = meta
		this.Metas = append(this.Metas, meta)
		meta.BaseResource = this
		meta.updateMeta()
	}

	return meta
}

// GetMetas get metas with give attrs
func (this *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = this.allAttrs()
	}
	var showSections, ignoredAttrs []string
	for _, attr := range attrs {
		if strings.HasPrefix(attr, "-") {
			ignoredAttrs = append(ignoredAttrs, strings.TrimLeft(attr, "-"))
		} else {
			showSections = append(showSections, attr)
		}
	}

	metas := []resource.Metaor{}

Attrs:
	for _, attr := range showSections {
		for _, a := range ignoredAttrs {
			if attr == a {
				continue Attrs
			}
		}

		meta := this.GetMetaOrSet(attr)
		metas = append(metas, meta)
	}

	return metas
}

func (this *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if this.cachedMetas == nil {
		this.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*this.cachedMetas)[cacheKey]; ok {
		return values
	}

	values := fc()
	var metas []*Meta
	for _, value := range values {
		metas = append(metas, value.(*Meta))
	}
	(*this.cachedMetas)[cacheKey] = metas
	return metas
}

func (this *Resource) MetasFromLayoutContext(l *Layout, context *Context, value interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
	if l.MetasFunc != nil {
		metas, names = l.MetasFunc(this, context, value, roles...)
	} else if l.MetaNamesFunc != nil {
		namess := l.MetaNamesFunc(this, context, value, roles...)
		if len(namess) > 0 {
			metas = this.ConvertSectionToMetas(this.allowedSections(value, this.generateSections(namess), context, roles...))
		}
	} else if len(l.Metas) > 0 {
		for _, metaName := range l.Metas {
			metas = append(metas, this.MetasByName[metaName])
		}

		names = l.MetaNames
	}

	if len(metas) > 0 && len(names) == 0 {
		names = make([]*resource.MetaName, len(metas), len(metas))

		if l.MetaAliases == nil {
			for i, meta := range metas {
				names[i] = meta.Namer()
			}
		} else {
			for i, meta := range metas {
				if alias, ok := l.MetaAliases[meta.Name]; ok {
					names[i] = alias
				} else {
					names[i] = meta.Namer()
				}
			}
		}
	}

	if context.Encodes() && len(this.PrimaryFields) > 0 && this.Fragment == nil {
		for _, name := range names {
			if name.Name == this.PrimaryFields[0].Name {
				return
			}
		}
		names = append(names, nil)
		copy(names[1:], names)
		names[0] = &resource.MetaName{"", "ID"}

		metas = append(metas, nil)
		copy(metas[1:], metas)
		metas[0] = &Meta{
			Meta: &resource.Meta{
				Typ: reflect.TypeOf((*string)(nil)).Elem(),
			},
			FormattedValuer: func(recorde interface{}, context *core.Context) interface{} {
				return this.GetKey(recorde).String()
			},
			mustValuer: true,
		}
	}
	return
}

func (this *Resource) MetasFromLayoutNameContext(layout string, context *Context, value interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
	if l := this.GetLayout(layout); l != nil {
		if len(roles) == 0 {
			defaultRole := DefaultPermission(layout)
			roles = append(roles, defaultRole)
		}
		return this.MetasFromLayoutContext(l.(*Layout), context, value, roles...)
	}
	return
}

func (this *Resource) MetaContextGetter(ctx *Context) func(name string) *Meta {
	if this.MetaContextGetterFunc != nil {
		return this.MetaContextGetterFunc(ctx)
	}
	return func(name string) *Meta {
		return this.GetMeta(name)
	}
}

func NewMeta(meta *Meta) *Meta {
	if meta.BaseResource == nil {
		panic(fmt.Errorf("meta.BaseResource is nil"))
	}
	meta.updateMeta()
	return meta
}

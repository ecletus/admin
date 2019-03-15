package admin

import (
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
)

func (res *Resource) GetDefinedMeta(name string) *Meta {
	meta := res.MetasByName[name]
	if meta == nil {
		meta = res.MetasByFieldName[name]
	}
	return meta
}

func (res *Resource) GetMetaOrSet(name string) (meta *Meta) {
	if meta = res.GetDefinedMeta(name); meta == nil {
		meta = res.Meta(&Meta{Name: name})
	}
	return
}

// GetMeta get meta with name
func (res *Resource) GetMeta(name string, notUpdate ...bool) *Meta {
	return res.getMeta(&Meta{Name: name}, notUpdate...)
}

func (res *Resource) getMeta(meta *Meta, notUpdate ...bool) *Meta {
	fallbackMeta := res.MetasByName[meta.Name]

	if fallbackMeta == nil {
		fallbackMeta = res.MetasByFieldName[meta.Name]
	}

	if meta.Type == "-" {
		meta.Enabled = func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}
		meta.Type = ""
	}

	if fallbackMeta == nil {
		if field, ok := res.FakeScope.FieldByName(meta.Name); ok {
			if meta.BaseResource == nil {
				meta.BaseResource = res
			}
			if field.IsPrimaryKey {
				meta.Type = "hidden_primary_key"
			}
			res.MetasByName[meta.Name] = meta
			res.MetasByFieldName[meta.Name] = meta
			res.Metas = append(res.Metas, meta)
			meta.updateMeta()
			return meta
		} else if field := res.FakeScope.GetVirtualField(meta.Name); field != nil {
			if meta.BaseResource == nil {
				meta.BaseResource = res
			}
			res.MetasByName[meta.Name] = meta
			res.MetasByFieldName[meta.Name] = meta
			res.Metas = append(res.Metas, meta)
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
		} else if meta.Name == META_STRING {
			if meta.Label == "" {
				meta.Label = res.SingularLabelKey()
			}
			if meta.Type == "" {
				meta.Type = "string"
			}
			if meta.Valuer == nil {
				meta.Valuer = func(recorde interface{}, context *core.Context) interface{} {
					return utils.StringifyContext(recorde, context)
				}
			}
			res.MetasByName[meta.Name] = meta
			meta.BaseResource = res
			meta.updateMeta()
			return meta
		} else {
			parts := strings.Split(meta.Name, ".")
			if len(parts) > 1 {
				r := res
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
					meta = NewMetaFieldProxy(meta.Name, pth, res.Value, to)
					res.MetasByName[meta.Name] = meta
					res.Metas = append(res.Metas, meta)
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
func (res *Resource) SetMeta(meta *Meta, notUpdate ...bool) *Meta {
	return res.Meta(meta, true)
}

// MetaDisable disable metas by name
func (res *Resource) MetaDisable(names ...string) {
	for _, name := range names {
		res.Meta(&Meta{Name: name, Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}})
	}
}

// Meta register meta for admin resource
func (res *Resource) Meta(meta *Meta, notUpdate ...bool) *Meta {
	if oldMeta := res.getMeta(meta, notUpdate...); oldMeta != nil {
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

			meta = oldMeta
			meta.updateMeta()
		}
	} else {
		res.MetasByName[meta.Name] = meta
		res.Metas = append(res.Metas, meta)
		meta.BaseResource = res
		meta.updateMeta()
	}

	return meta
}

// GetMetas get metas with give attrs
func (res *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = res.allAttrs()
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

		meta := res.GetMetaOrSet(attr)
		metas = append(metas, meta)
	}

	return metas
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	}

	values := fc()
	var metas []*Meta
	for _, value := range values {
		metas = append(metas, value.(*Meta))
	}
	(*res.cachedMetas)[cacheKey] = metas
	return metas
}

func (res *Resource) MetasFromLayoutContext(layout string, context *Context, value interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
	if len(roles) == 0 {
		defaultRole := DefaultPermission(layout)
		roles = append(roles, defaultRole)
	}
	l := res.GetLayout(layout).(*Layout)
	if l != nil {
		if l.MetasFunc != nil {
			metas, names = l.MetasFunc(res, context, value, roles...)
		} else if l.MetaNamesFunc != nil {
			namess := l.MetaNamesFunc(res, context, value, roles...)
			if len(namess) > 0 {
				metas = res.ConvertSectionToMetas(res.allowedSections(value, res.generateSections(namess), context, roles...))
			}
		} else if len(l.Metas) > 0 {
			for _, metaName := range l.Metas {
				metas = append(metas, res.MetasByName[metaName])
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
	}
	return
}

package admin

import (
	"reflect"
	"strings"

	"github.com/ecletus/helpers"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/maps"
	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

const (
	SelectConfigOptionNotIcon SelectConfigOption = 1 << iota
	SelectConfigOptionAllowBlank
	SelectConfigOptionBottonSheet
)

type SelectConfigOption uint8

func (b SelectConfigOption) Set(flag SelectConfigOption) SelectConfigOption    { return b | flag }
func (b SelectConfigOption) Clear(flag SelectConfigOption) SelectConfigOption  { return b &^ flag }
func (b SelectConfigOption) Toggle(flag SelectConfigOption) SelectConfigOption { return b ^ flag }
func (b SelectConfigOption) Has(flag SelectConfigOption) bool                  { return (b & flag) != 0 }

func (b SelectConfigOption) String() string {
	var s []string
	if b.Has(SelectConfigOptionNotIcon) {
		s = append(s, "not_icon")
	}
	if b.Has(SelectConfigOptionAllowBlank) {
		s = append(s, "allow_blank")
	}
	if b.Has(SelectConfigOptionBottonSheet) {
		s = append(s, "bottom_sheet")
	}
	return strings.Join(s, "|")
}

func (b *SelectConfigOption) Parse(s string) {
	for _, r := range s {
		switch r {
		case 'I':
			*b |= SelectConfigOptionNotIcon
		case 'b':
			*b |= SelectConfigOptionAllowBlank
		case 'S':
			*b |= SelectConfigOptionBottonSheet
		}
	}
}

func (ct SelectConfigOption) S() string {
	return ct.String()
}

func SelectOne(r *Resource, names ...NameCallback) {
	SelectOneOption(0, r, names...)
}

func SelectOneBS(r *Resource, names ...NameCallback) {
	SelectOneOption(SelectConfigOptionBottonSheet, r, names...)
}

type NameCallback struct {
	Name             string
	Callback         func(meta *Meta)
	PrepareOneConfig func(cfg *SelectOneConfig)
}

func SelectOneOption(baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	DoSelectOption(false, baseOpt, r, names...)
}

func SelectManyOption(baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	DoSelectOption(true, baseOpt, r, names...)
}

func DoSelectOption(many bool, baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	onResource := func(index int, name, scheme string, opt SelectConfigOption, rs *Resource) {
		var (
			cfg    = r.Meta(&Meta{Name: name}).Config
			oneCfg *SelectOneConfig
		)

		if cfg == nil {
			oneCfg = &SelectOneConfig{}
		} else {
			if many {
				oneCfg = &cfg.(*SelectManyConfig).SelectOneConfig
			} else {
				oneCfg = cfg.(*SelectOneConfig)
			}
		}

		oneCfg.Basic = true
		oneCfg.AllowBlank = opt.Has(SelectConfigOptionAllowBlank)

		var res = oneCfg.RemoteDataResource
		if res == nil {
			res = NewDataResource(rs)
			oneCfg.RemoteDataResource = res
		} else {
			oneCfg.RemoteDataResource.Resource = rs
		}
		if opt.Has(SelectConfigOptionNotIcon) {
			res.Layout = BASIC_LAYOUT_HTML
		}
		var mode string
		if opt.Has(SelectConfigOptionBottonSheet) {
			mode = "bottom_sheet"
		}
		oneCfg.RemoteDataResource = res
		oneCfg.SelectMode = mode
		oneCfg.Scheme = scheme

		if cfg == nil {
			if many {
				cfg = &SelectManyConfig{
					SelectOneConfig: *oneCfg,
				}
			} else {
				cfg = oneCfg
			}
		}

		var meta *Meta
		meta = r.Meta(&Meta{
			Resource: rs,
			Name:     name,
			Config:   cfg,
		})

		meta.NewFormattedValuer(func(meta *Meta, old MetaFValuer, record interface{}, ctx *core.Context) *FormattedValue {
			if old != nil {
				return old(record, ctx)
			}
			if record == nil {
				return nil
			}
			value := meta.Value(ctx, record)
			if helpers.IsNilInterface(value) {
				return nil
			}
			fv := &FormattedValue{Record: record, Raw: value}
			switch t := value.(type) {
			case Stringer:
				fv.Value = t.AdminString(ContextFromCoreContext(ctx), maps.Map{})
			case core.ContextStringer:
				fv.Value = t.ContextString(ctx)
			}
			return fv
		})

		if field, ok := r.ModelStruct.FieldsByName[name]; ok && field.Relationship != nil {
			rel := field.Relationship
			switch rel.Kind {
			case aorm.BELONGS_TO, aorm.HAS_ONE:
				meta.SetSetter(func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) (err error) {
					if metaValue.Value == nil {
						return nil
					}

					defer func() {
						if err != nil {
							err = errors.Wrap(err, PKG+".SelectOneOption auto setter")
						}
					}()

					var (
						v  = metaValue.StringValue()
						ID aorm.ID
					)

					if v == "" {
						if rel.GetRelatedID(recorde).IsZero() {
							return
						}
						ID = rel.AssociationModel.DefaultID()
					} else if ID, err = rel.AssociationModel.ParseIDString(v); err != nil {
						return
					} else if ID.Eq(rel.GetRelatedID(recorde)) {
						return
					}
					rel.SetRelatedID(recorde, ID)
					if rel.Field != nil {
						if field := reflect.Indirect(reflect.ValueOf(recorde)).FieldByIndex(rel.Field.StructIndex); field.IsValid() {
							for field.IsValid() {
								switch field.Kind() {
								case reflect.Ptr, reflect.Interface:
									field.Set(reflect.Zero(field.Type()))
									return nil
								default:
									if field.IsValid() {
										aorm.SetZero(field)
									}
									return
								}
							}
						}
					}

					return nil
				})
			default:
				panic("not implemented")
			}
		}

		if names[index].Callback != nil {
			names[index].Callback(meta)
		}
	}
	Admin := r.GetAdmin()
	m := func(index int, name, scheme string, opt SelectConfigOption) {
		id_ := r.FullID() + "." + name
		opt |= baseOpt

		if meta := r.GetDefinedMeta(name); meta != nil && meta.Resource != nil {
			onResource(index, name, scheme, opt, meta.Resource)
			return
		}
		if err := Admin.OnResourcesAdded(func(e *ResourceEvent) error {
			if e.Resource.Config.NotMount {
				onResource(index, name, scheme, opt, e.Resource)
			} else {
				e.Resource.PostMount(func() {
					onResource(index, name, scheme, opt, e.Resource)
				})
			}
			return nil
		}, id_); err != nil {
			panic(errors.Wrap(err, id_))
		}
	}

	for i, nameCb := range names {
		var name = nameCb.Name
		parts := strings.Split(name, ":")
		var (
			opt    SelectConfigOption
			scheme string
		)
		if len(parts) > 1 {
			opt.Parse(parts[0])
			parts = parts[1:]
		}

		name = parts[0]
		parts = strings.Split(name, ">")

		if len(parts) > 1 {
			name = parts[0]
			scheme = parts[1]
		}

		m(i, name, scheme, opt)
	}
}

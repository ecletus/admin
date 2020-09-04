package admin

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ecletus/helpers"
	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
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
	Name     string
	Callback func(meta *Meta)
}

func SelectOneOption(baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	DoSelectOption(false, baseOpt, r, names...)
}

func SelectManyOption(baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	DoSelectOption(true, baseOpt, r, names...)
}

func DoSelectOption(many bool, baseOpt SelectConfigOption, r *Resource, names ...NameCallback) {
	onResource := func(index int, name, scheme string, opt SelectConfigOption, rs *Resource) {
		res := NewDataResource(rs)
		if opt.Has(SelectConfigOptionNotIcon) {
			res.Layout = BASIC_LAYOUT_HTML
		}
		var mode string
		if opt.Has(SelectConfigOptionBottonSheet) {
			mode = "bottom_sheet"
		}
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
			FormattedValuer: func(record interface{}, context *core.Context) (result interface{}) {
				if record != nil {
					value := meta.Value(context, record)
					if !helpers.IsNilInterface(value) {
						if s, ok := value.(fmt.Stringer); ok {
							return s.String()
						}
						return value
					}
				}
				return ""
			},

			Config: cfg,
		})

		if field, ok := r.ModelStruct.FieldsByName[name]; ok && field.Relationship != nil {
			if rel := field.Relationship; rel.Kind == "belongs_to" || rel.Kind == "has_one" {
				meta.SetSetter(func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
					var (
						rev               = reflect.ValueOf(recorde).Elem()
						valuesS           = metaValue.Value.([]string)
						foreignFieldNames = rel.ForeignFieldNames
					)
					if lv, lf := len(valuesS), len(foreignFieldNames); lv > lf {
						valuesS = valuesS[0:len(foreignFieldNames)]
					} else if lf > lv {
						foreignFieldNames = foreignFieldNames[0:lv]
					}
					values := reflect.ValueOf(valuesS)
					for i, name := range foreignFieldNames {
						value := rev.FieldByName(name)
						switch value.Kind() {
						case reflect.String:
							value.Set(values.Index(i))
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							var uiv, err = strconv.ParseUint(values.Index(i).String(), 10, 64)
							if err != nil {
								return errors.Wrap(err, PKG+".SelectOneOption meta auto setter: parse id failed")
							}
							value.SetUint(uiv)
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							var uiv, err = strconv.ParseInt(values.Index(i).String(), 10, 64)
							if err != nil {
								return errors.Wrap(err, PKG+".SelectOneOption meta auto setter: parse id failed")
							}
							value.SetInt(uiv)
						default:
							vi := value.Addr().Interface()
							if parse, ok := vi.(aorm.StringParser); ok {
								return parse.ParseString(valuesS[i])
							}
							return errors.New(PKG + ".SelectOneOption meta auto setter: invalid type")
						}
					}
					return nil
				})
			}
		} else {
			meta.SetSetter(func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				// fake
				return nil
			})
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
				e.Resource.AfterMount(func() {
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

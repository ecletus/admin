package admin_helpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ecletus/core/utils"

	"github.com/ecletus/admin"
	"github.com/ecletus/core"
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

func SelectOne(r *admin.Resource, names ...string) {
	SelectOneOption(0, r, names...)
}

func SelectOneOption(baseOpt SelectConfigOption, r *admin.Resource, names ...string) {
	typ := utils.IndirectType(r.Value)

	Admin := r.GetAdmin()
	m := func(name, scheme string, opt SelectConfigOption) {
		opt |= baseOpt

		field, _ := typ.FieldByName(name)
		value := reflect.New(utils.IndirectType(field.Type)).Interface()
		_ = Admin.OnResourceValueAdded(value, func(e *admin.ResourceEvent) {
			res := admin.NewDataResource(e.Resource)
			if opt.Has(SelectConfigOptionNotIcon) {
				res.Layout = admin.BASIC_LAYOUT_HTML
			}
			var mode string
			if opt.Has(SelectConfigOptionBottonSheet) {
				mode = "bottom_sheet"
			}
			var meta *admin.Meta
			meta = r.Meta(&admin.Meta{
				Resource: e.Resource,
				Name:     name,
				FormattedValuer: func(record interface{}, context *core.Context) (result interface{}) {
					if record != nil {
						rv := reflect.ValueOf(record).Elem().FieldByName(name)
						if rv.Kind() != reflect.Ptr {
							rv = rv.Addr()
						}
						if !rv.IsNil() {
							if s, ok := rv.Interface().(fmt.Stringer); ok {
								return s.String()
							}
							if meta.Valuer != nil {
								return meta.Valuer(record, context)
							}
						}
					}
					return ""
				},

				Config: &admin.SelectOneConfig{
					Basic:              true,
					AllowBlank:         opt.Has(SelectConfigOptionAllowBlank),
					RemoteDataResource: res,
					SelectMode:         mode,
					Scheme:             scheme,
				},
			})
		})
	}

	for _, name := range names {
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

		m(name, scheme, opt)
	}
}

package admin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
)

type FieldFilter struct {
	FieldName               string
	Struct, OriginalStruct  *aorm.StructField
	Virtual                 *aorm.VirtualField
	InlineQueryName         string
	TermFormatterFunc       func(term interface{}) interface{}
	QueryFieldFormatterFunc func(f *FieldFilter, queryField string) string
	Applyer                 func(arg interface{}) (query string, argx interface{})
	Handler                 func(db *aorm.DB, filterArgument *FilterArgument) *aorm.DB
}

func (f *FieldFilter) With(fn func(f *FieldFilter)) *FieldFilter {
	fn(f)
	return f
}

func (f *FieldFilter) QueryField() string {
	if f.QueryFieldFormatterFunc != nil {
		return f.QueryFieldFormatterFunc(f, f.InlineQueryName)
	}
	return f.InlineQueryName
}

func (f *FieldFilter) FormatTerm(term interface{}) interface{} {
	if f.TermFormatterFunc != nil {
		return f.TermFormatterFunc(term)
	}

	if s, ok := term.(string); ok {
		if reflect.PtrTo(f.Struct.Struct.Type).Implements(reflect.TypeOf((*aorm.StringParser)(nil)).Elem()) {
			v := reflect.New(f.Struct.Struct.Type)
			_ = v.Interface().(aorm.StringParser).ParseString(s)
			term = v.Elem().Interface()
		}
	}
	return term
}

func NewFieldFilter(res *Resource, fieldName string) (f *FieldFilter) {
	f = &FieldFilter{FieldName: fieldName}
	if fpq := res.ModelStruct.FieldPathQueryOf(fieldName); fpq != nil {
		f.Struct = fpq.Struct
		f.Virtual = fpq.Virtual
		f.InlineQueryName = fpq.Query()
		f.OriginalStruct = res.ModelStruct.FieldsByName[fieldName]
		if f.OriginalStruct != nil && f.OriginalStruct.Relationship != nil {
			switch f.OriginalStruct.Relationship.Kind {
			case aorm.BELONGS_TO:
				f.TermFormatterFunc = func(term interface{}) interface{} {
					var (
						err error
						s   = term.(string)
					)
					if s == "" {
						id := f.OriginalStruct.Model.GetID(reflect.New(f.OriginalStruct.Model.Type).Interface())
						return id
					}
					if term, err = f.OriginalStruct.Model.ParseIDString(s); err != nil {
						panic(fmt.Errorf("filter of %s#%q: convert term %q to ID: %s", res.UID, s, fieldName, err.Error()))
					}
					return term
				}
				f.Applyer = func(arg interface{}) (query string, argx interface{}) {
					return "", aorm.FilterOfID(arg.(aorm.ID), f.InlineQueryName).Fpq()
				}
			}
		} else if f.OriginalStruct == nil && f.Struct.IsPrimaryKey {
			f.TermFormatterFunc = func(term interface{}) interface{} {
				term, _ = f.Struct.BaseModel.ParseIDString(term.(string))
				return term
			}
			f.Applyer = func(arg interface{}) (query string, argx interface{}) {
				return "", aorm.FilterOfID(arg.(aorm.ID), f.InlineQueryName).Fpq()
			}
		}
		return f
	}
	return nil
}

// Filter filter definiation
type Filter struct {
	Scheme *Scheme

	ID                 uintptr
	Name               string
	Label              string
	LabelDisabled      bool
	Type               string
	DefaultOperation   string
	Operations         []string // eq, cont, gt, gteq, lt, lteq
	NotChooseOperation bool
	Resource           *Resource
	Handler            func(db *aorm.DB, arg *FilterArgument) *aorm.DB
	HandleEmpty        bool
	Config             FilterConfigInterface
	Available          func(context *Context) bool
	Visible            func(context *Context) bool
	Hidden             bool
	AllowBlank         bool
	advanced           bool

	DefaultLabel  string
	FieldLabel    bool
	LabelPairFunc func(ctx *core.Context) (key, defaul string)

	FieldName string
	Field     *FieldFilter
	Valuer    func(arg *FilterArgument) (value interface{}, err error)
	index     uint32
}

func (this *Filter) Index() uint32 {
	return this.index
}

func (this *Filter) With(fn func(f *Filter)) *Filter {
	fn(this)
	return this
}

func (this *Filter) IsAdvanced() bool {
	return this.advanced
}

func (this *Filter) Advanced() *Filter {
	this.advanced = true
	return this
}

func (this *Filter) IsVisible(context *Context) bool {
	if this.Hidden {
		return false
	}
	if this.Visible != nil {
		return this.Visible(context)
	}
	return true
}

func (this *Filter) GetLabelC(ctx *core.Context) string {
	if key, defaul := this.GetLabelPair(ctx); key != "" {
		return ctx.Ts(key, defaul)
	} else {
		return defaul
	}
}

func (this *Filter) GetLabelPair(ctx *core.Context) (string, string) {
	if this.LabelPairFunc != nil {
		return this.LabelPairFunc(ctx)
	}

	return fmt.Sprintf("%v.filter.%v", this.Scheme.Resource.I18nPrefix, this.Name), this.Label
}

// FilterConfigInterface filter config interface
type FilterConfigInterface interface {
	ConfigureQORAdminFilter(*Filter)
}

// FilterArgument filter argument that used in handler
type FilterArgument struct {
	Default  bool
	Scheme   *Scheme
	Filter   *Filter
	Value    *resource.MetaValues
	Resource *Resource
	Context  *core.Context
	GoValue  interface{}
}

func (this *FilterArgument) QueryValues() (res []struct{ Name, Value string }) {
	for _, v := range this.Value.Values {
		v.EachQueryVal([]string{""}, func(prefix []string, name string, value interface{}) {
			res = append(res, struct{ Name, Value string }{Name: fmt.Sprintf("filter[%s]%s", this.Filter.Name, strings.Join(append(prefix, name), ".")), Value: fmt.Sprint(value)})
		})
	}
	return
}

// SavedFilter saved filter settings
type SavedFilter struct {
	Name string
	URL  string
}

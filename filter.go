package admin

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/moisespsena-go/aorm"
)

type FieldFilter struct {
	FieldName               string
	Struct                  *aorm.StructField
	Virtual                 *aorm.VirtualField
	InlineQueryName         string
	TermForamtterFunc       func(term interface{}) interface{}
	QueryFieldFormatterFunc func(f *FieldFilter, queryField string) string
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
	if f.TermForamtterFunc != nil {
		return f.TermForamtterFunc(term)
	}
	return term
}

func NewFieldFilter(res *Resource, fieldName string) (f *FieldFilter) {
	f = &FieldFilter{FieldName: fieldName}
	f.Struct, f.Virtual = res.FakeScope.GetModelStruct().FieldDiscovery(fieldName)
	if f.Struct != nil {
		if f.Struct.Relationship != nil {
			typ := f.Struct.Struct.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}

			ms := res.FakeScope.New(reflect.New(typ).Interface()).GetModelStruct()
			f.Struct = ms.PrimaryFields[0]
			f.FieldName += "." + f.Struct.Name
		}
		parts := strings.Split(f.FieldName, ".")
		f.InlineQueryName = "{" + strings.Join(parts[0:len(parts)-1], ".") + "}." + f.Struct.DBName
		return f
	}
	return nil
}

// Filter filter definiation
type Filter struct {
	Scheme *Scheme

	Name               string
	Label              string
	Type               string
	DefaultOperation   string
	Operations         []string // eq, cont, gt, gteq, lt, lteq
	NotChooseOperation bool
	Resource           *Resource
	Handler            func(*aorm.DB, *FilterArgument) *aorm.DB
	Config             FilterConfigInterface
	Available          func(context *Context) bool
	Visible            func(context *Context) bool
	Hidden             bool
	advanced           bool

	DefaultLabel string
	FieldLabel   bool
	LabelFunc    func() (string, string)

	FieldName string
	Field     *FieldFilter
}

func (f *Filter) With(fn func(f *Filter)) *Filter {
	fn(f)
	return f
}

func (f *Filter) IsAdvanced() bool {
	return f.advanced
}

func (f *Filter) Advanced() *Filter {
	f.advanced = true
	return f
}

func (f *Filter) IsVisible(context *Context) bool {
	if f.Hidden {
		return false
	}
	if f.Visible != nil {
		return f.Visible(context)
	}
	return true
}

func (f *Filter) GetLabelPair() (string, string) {
	if f.LabelFunc != nil {
		return f.LabelFunc()
	}

	return fmt.Sprintf("%v.filter.%v", f.Scheme.Resource.I18nPrefix, f.Name), f.Label
}

func (f *Filter) GetLabel(ctx *core.Context) template.HTML {
	key, defaul := f.GetLabelPair()
	return ctx.T(key, defaul)
}

// FilterConfigInterface filter config interface
type FilterConfigInterface interface {
	ConfigureQORAdminFilter(*Filter)
}

// FilterArgument filter argument that used in handler
type FilterArgument struct {
	Scheme   *Scheme
	Filter   *Filter
	Value    *resource.MetaValues
	Resource *Resource
	Context  *core.Context
}

// SavedFilter saved filter settings
type SavedFilter struct {
	Name string
	URL  string
}

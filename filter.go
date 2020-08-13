package admin

import (
	"fmt"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

type FieldFilter struct {
	FieldName               string
	Struct                  *aorm.StructField
	Virtual                 *aorm.VirtualField
	InlineQueryName         string
	TermFormatterFunc       func(term interface{}) interface{}
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
	if f.TermFormatterFunc != nil {
		return f.TermFormatterFunc(term)
	}
	return term
}

func NewFieldFilter(res *Resource, fieldName string) (f *FieldFilter) {
	f = &FieldFilter{FieldName: fieldName}
	if fpq := res.ModelStruct.FieldPathQueryOf(fieldName); fpq != nil {
		f.Struct = fpq.Struct
		f.Virtual = fpq.Virtual
		f.InlineQueryName = fpq.Query()
		return f
	}
	return nil
}

// Filter filter definiation
type Filter struct {
	Scheme *Scheme

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
	Scheme   *Scheme
	Filter   *Filter
	Value    *resource.MetaValues
	Resource *Resource
	Context  *core.Context
	GoValue  interface{}
}

// SavedFilter saved filter settings
type SavedFilter struct {
	Name string
	URL  string
}

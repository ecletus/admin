package admin

import (
	"strings"

	"github.com/ecletus/fragment"
	"github.com/ecletus/roles"
	"github.com/go-errors/errors"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

var NotFragmentableError = errors.New("Not fragmentable")
var FRAGMENT_KEY = PKG + ".fragment"

type FormFragmentRecordState struct {
	*Fragment
	Enabled bool
	Value   fragment.FormFragmentModelInterface
	IsNil   bool
}

func (f *FormFragmentRecordState) EditSections(context *Context) (sections []*Section) {
	if !f.Config.NotInline && (f.Value != nil && f.Value.Enabled()) {
		sections = append(sections, f.Resource.EditAttrs()...)
	}
	sections = f.Resource.allowedSections(f.Value, sections, context, roles.Update)

	return append([]*Section{{Resource: f.Resource, Rows: [][]string{{AttrFragmentEnabled}}}}, sections...)
}
func (f *FormFragmentRecordState) ShowSections(context *Context) (sections []*Section) {
	if f.Config.NotInline {
		sections = append(sections, &Section{Resource: f.Resource, Rows: [][]string{{AttrFragmentEnabled}}})
	} else {
		sections = f.Resource.allowedSections(f.Value, f.Resource.ShowAttrs(), context, roles.Read)
	}
	return
}
func (f *FormFragmentRecordState) OnlyEnabledField(context *Context) bool {
	if f.Config.NotInline {
		return true
	}
	if context.Type.Has(EDIT) {
		return f.Value == nil || !f.Value.Enabled()
	}
	return false
}

type FragmentCategoryConfig struct {
	Label      string
	IndexAttrs func(res *Resource, ctx *core.Context) []interface{}
}

type FragmentConfig struct {
	Config      *Config
	NotInline   bool
	Priority    int
	Category    *FragmentCategoryConfig
	IsLabel     string
	Is          bool
	Enabled     func(record fragment.FragmentedModelInterface, ctx *core.Context) bool
	Available   func(record fragment.FragmentedModelInterface, ctx *core.Context) bool
	SchemeSetup func(s *Scheme)
	Schemes     []*Scheme
	Sections    []*Section
	LoadInFindMany bool
}

func (fc *FragmentConfig) Inline() bool {
	return !fc.NotInline
}

type Fragment struct {
	IsForm      bool
	Resource    *Resource
	ID          string
	Config      *FragmentConfig
	fieldsCount int
	joinQuery   string
	fields      []*aorm.StructField
	query       string
	isURI       string
	scheme      *Scheme
}

func (f *Fragment) FieldsCount() int {
	return f.fieldsCount
}

func (f *Fragment) Scheme() *Scheme {
	if f.scheme == nil {
		return f.Resource.Scheme
	}
	return f.scheme
}

func (f *Fragment) BaseResource() *Resource {
	baseResource := f.Resource
	for baseResource.Fragment != nil {
		baseResource = baseResource.ParentResource
	}
	return baseResource
}

func (f *Fragment) Build() {
	if f.fieldsCount > 0 || f.Resource.Config.Virtual {
		return
	}

	f.buildFields()
	f.buildQuery()

	f.joinQuery = "JOIN !"
	if !f.BaseResource().IsSingleton() {
		f.joinQuery += " ON !.id = ?.id"
	} else {
		f.joinQuery += " ON 1 = 1"
	}
	f.fieldsCount = len(f.fields)

	if f.Resource.Fragments != nil {
		f.Resource.Fragments.Build()
		f.fieldsCount += f.Resource.Fragments.columnsCount
	}
}

func (f *Fragment) IsURI() string {
	return f.isURI
}

func (f *Fragment) buildFields() {
	fields := append([]*aorm.StructField{}, f.Resource.ModelStruct.NonIgnoredStructFields()...)
	if !f.IsForm {
		for i, field := range fields {
			if field.Name == AttrFragmentEnabled {
				if i != 0 {
					fields = append([]*aorm.StructField{field}, append(fields[0:i], fields[i+1:]...)...)
				}
				break
			}
		}
	}

	var newFields []*aorm.StructField

	for _, field := range fields {
		if !field.IsPrimaryKey && field.Relationship == nil {
			newFields = append(newFields, field)
		}
	}

	f.fields = newFields
}

func (f *Fragment) buildQuery() {
	columns := make([]string, len(f.fields))
	for i, field := range f.fields {
		columns[i] = "!." + field.DBName
	}
	f.query = strings.Join(columns, ", ")
}

func (f *Fragment) Fields() []*aorm.StructField {
	return f.fields
}

func (f *Fragment) FieldsNames() []string {
	names := make([]string, len(f.fields))
	for i, field := range f.fields {
		names[i] = field.Name
	}
	return names
}

func (f *Fragment) AllFields() []*aorm.StructField {
	fields := f.Fields()
	if f.Resource.Fragments != nil {
		f.Resource.Fragments.Walk(func(fr *Fragment) error {
			fields = append(fields, fr.Fields()...)
			return nil
		})
	}
	return fields
}

func (f *Fragment) formatQuery(DB *aorm.DB, query string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(query, "?", f.Resource.ParentResource.QuotedTableName(DB)),
		"!", f.Resource.QuotedTableName(DB))
}

func (f *Fragment) Query(DB *aorm.DB) string {
	return f.formatQuery(DB, f.query)
}

func (f *Fragment) AllQuery(DB *aorm.DB) string {
	queries := []string{f.Query(DB)}
	if f.Resource.Fragments != nil {
		f.Resource.Fragments.Walk(func(fr *Fragment) error {
			queries = append(queries, fr.Query(DB))
			return nil
		})
	}
	return strings.Join(queries, ", ")
}

func (f *Fragment) JoinLeft(DB *aorm.DB) *aorm.DB {
	if !f.Resource.Config.Virtual {
		DB = DB.Joins("LEFT " + f.formatQuery(DB, f.joinQuery))
		if f.Resource.Fragments != nil {
			return f.Resource.Fragments.JoinLeft(DB)
		}
	}
	return DB
}

func (f *Fragment) Join(DB *aorm.DB) *aorm.DB {
	if !f.Resource.Config.Virtual {
		DB = DB.Joins(f.formatQuery(DB, f.query))
		if f.Resource.Fragments != nil {
			return f.Resource.Fragments.JoinLeft(DB)
		}
	}
	return DB
}

func (f *Fragment) Filter(DB *aorm.DB) *aorm.DB {
	if !f.Resource.Config.Virtual {
		super := f
		for super != nil {
			DB = DB.Where(aorm.IQ("{" + f.ID + "}.fragment_enabled"))
			super = super.Resource.ParentResource.Fragment
		}
	}
	return DB
}

func (f *Fragment) Parents() (parents []*Fragment) {
	super := f
	for super != nil {
		parents = append(parents, super)
		super = super.Resource.ParentResource.Fragment
	}
	if l := len(parents); l > 1 {
		for i, j := 0, l-1; i < j; i, j = i+1, j-1 {
			parents[i], parents[j] = parents[j], parents[i]
		}
	}
	return
}

func (f *Fragment) Enabled(record fragment.FragmentedModelInterface, ctx *core.Context) bool {
	if !f.IsForm {
		return true
	} else if f.Config.NotInline && record.GetFormFragment(f.ID) != nil {
		fv := record.GetFormFragment(f.ID)
		if GetContext(ctx).Type.Has(SHOW) {
			if fv.Enabled() && (f.Config.Enabled == nil || f.Config.Enabled(record, ctx)) {
				return true
			}
			return false
		}
		return true
	}

	fv := record.GetFormFragment(f.ID)
	if GetContext(ctx).Type.Has(SHOW) && (fv == nil || !fv.Enabled()) {
		return false
	}
	enabled := f.Config.Enabled == nil || f.Config.Enabled(record, ctx)
	return enabled
}

func (f *Fragment) FormRecordValue(recorde fragment.FragmentedModelInterface, ctx *core.Context) *FormFragmentRecordState {
	value := recorde.GetFormFragment(f.ID)
	return &FormFragmentRecordState{f, f.Enabled(recorde, ctx), value, value == nil}
}

func (f *Fragment) FormGetOrNew(recorde fragment.FragmentedModelInterface, ctx *core.Context) fragment.FormFragmentModelInterface {
	value := recorde.GetFormFragment(f.ID)
	if value == nil {
		value = f.Resource.NewStruct(ctx.Site).(fragment.FormFragmentModelInterface)
		recorde.SetFormFragment(recorde, f.ID, value)
	}
	return value
}

func (f *Fragment) GetOrNew(recorde fragment.FragmentedModelInterface, ctx *core.Context) fragment.FragmentModelInterface {
	value := recorde.GetFragment(f.ID)
	if value == nil {
		value = f.Resource.NewStruct(ctx.Site).(fragment.FragmentModelInterface)
		recorde.SetFragment(recorde, f.ID, value)
	}
	return value
}

type fragmentActionArgument struct {
	Name string
}

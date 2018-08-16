package admin

import (
	"sort"

	"strings"

	"database/sql"

	"github.com/go-errors/errors"
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/fragment"
	"github.com/aghape/aghape"
	"github.com/aghape/aghape/utils"
	"github.com/aghape/roles"
)

var NotFragmentableError = errors.New("Not fragmentable")

type Fragments struct {
	Sorted       []string
	Fragments    map[string]*Fragment
	columnsCount int
	query        string
	fields       []*aorm.StructField
}

func NewFragments() *Fragments {
	return &Fragments{Fragments: make(map[string]*Fragment)}
}

func (f *Fragments) Build() {
	if f.columnsCount > 0 {
		return
	}
	for _, id := range f.Sorted {
		fr := f.Fragments[id]
		fr.Build()
		f.columnsCount += f.columnsCount
	}
}

func (f *Fragments) Walk(cb func(fr *Fragment) error) (err error) {
	for _, id := range f.Sorted {
		fr := f.Fragments[id]
		err = cb(fr)
		if err != nil {
			return
		}
		if fr.Resource.Fragments != nil {
			err = fr.Resource.Fragments.Walk(cb)
			if err != nil {
				return
			}
		}
	}
	return err
}

func (f *Fragments) JoinLeft(DB *aorm.DB) *aorm.DB {
	f.Build()
	for _, id := range f.Sorted {
		DB = f.Fragments[id].JoinLeft(DB)
	}
	return DB
}

func (f *Fragments) Join(DB *aorm.DB) *aorm.DB {
	f.Build()
	for _, id := range f.Sorted {
		DB = f.Fragments[id].Join(DB)
	}
	return DB
}

func (f *Fragments) NewSlice() []interface{} {
	r := make([]interface{}, len(f.Fragments))
	for i, _ := range r {
		r[i] = sql.NullString{}
	}
	return r
}

func (f *Fragments) Fields() (fields []*aorm.StructField) {
	if len(f.fields) == 0 {
		f.Build()
		f.Walk(func(fr *Fragment) error {
			fields = append(fields, fr.AllFields()...)
			return nil
		})
		f.fields = fields
	}
	return f.fields
}

func (f *Fragments) Query() (query string) {
	var queries []string
	f.Build()
	f.Walk(func(fr *Fragment) error {
		queries = append(queries, fr.AllQuery())
		return nil
	})
	return strings.Join(queries, ", ")
}

func (f *Fragments) add(res *Resource, isForm bool, cfg *FragmentConfig) *Fragment {
	fr := &Fragment{IsForm: isForm, Resource: res, ID: strings.TrimPrefix(res.ID, "fragments."), Config: cfg}
	res.Fragment = fr
	f.Fragments[fr.ID] = fr
	f.Sorted = append(f.Sorted, fr.ID)

	sort.Slice(f.Sorted, func(i, j int) bool {
		a, b := f.Fragments[f.Sorted[i]], f.Fragments[f.Sorted[j]]
		return a.Config.Priority < b.Config.Priority || b.Resource.Name < b.Resource.Name
	})

	return fr
}

func (f *Fragments) Add(res *Resource, cfg *FragmentConfig) *Fragment {
	return f.add(res, false, cfg)
}

func (f *Fragments) AddForm(res *Resource, cfg *FragmentConfig) *Fragment {
	return f.add(res, true, cfg)
}

func (f *Fragments) ExtraFieldsScan(result fragment.FragmentedModelInterface, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
	low := 0
	f.extraFieldsScan(result, values, &low, set)
}

func (f *Fragments) extraFieldsScan(result fragment.FragmentedModelInterface, values []interface{}, low *int, set func(result interface{}, low, hight int) interface{}) {
	f.Build()
	for _, id := range f.Sorted {
		fr := f.Fragments[id]

		// FragmentEnabled field. If is nil, is not defined.
		if values[*low].(*aorm.ValueScanner).IsNil() {
			// skip columns scan for this fragment and sub fragments
			*low += fr.fieldsCount
			if fr.Resource.Fragments != nil {
				fr.Resource.Fragments.Walk(func(fr *Fragment) error {
					*low += fr.fieldsCount
					return nil
				})
			}
		} else {
			value := set(fr.Resource.Value, *low, *low+fr.fieldsCount).(fragment.FragmentModelInterface)
			if fr.IsForm {
				result.SetFormFragment(id, value.(fragment.FormFragmentModelInterface))
			} else {
				result.SetFragment(id, value)
			}
			value.SetSuper(result)
			*low += fr.fieldsCount
			if fr.Resource.Fragments != nil {
				fr.Resource.Fragments.extraFieldsScan(value.(fragment.FragmentedModelInterface), values, low, set)
			}
		}
	}
}

func (f *Fragments) Get(id string) *Fragment {
	return f.Fragments[id]
}

func (f *Fragments) Slice(filter ...func(f *Fragment) bool) []*Fragment {
	filt := func(f *Fragment) bool {
		return true
	}
	if len(filter) > 0 && filter[0] != nil {
		filt = filter[0]
	}
	fs := make([]*Fragment, len(f.Fragments))
	for i := range fs {
		if fr := f.Fragments[f.Sorted[i]]; filt(fr) {
			fs[i] = fr
		}
	}
	return fs
}

type FormFragmentRecordState struct {
	*Fragment
	Enabled bool
	Value   fragment.FormFragmentModelInterface
	IsNil   bool
}

func (f *FormFragmentRecordState) EditSections(context *Context) []*Section {
	if f.Config.NotInline || (f.Value == nil || !f.Value.Enabled()) {
		return f.Resource.allowedSections(f.Value, []*Section{{Resource: f.Resource, Rows: [][]string{{"FragmentEnabled"}}}}, context, roles.Read)
	}
	return f.Resource.allowedSections(f.Value, f.Resource.EditAttrs(), context, roles.Read)
}
func (f *FormFragmentRecordState) ShowSections(context *Context) []*Section {
	if f.Config.NotInline {
		return f.Resource.allowedSections(f.Value, []*Section{{Resource: f.Resource, Rows: [][]string{{"FragmentEnabled"}}}}, context, roles.Read)
	}
	return f.Resource.allowedSections(f.Value, f.Resource.ShowAttrs(), context, roles.Read)
}
func (f *FormFragmentRecordState) OnlyEnabledField(context *Context) bool {
	if f.Config.NotInline {
		return true
	}
	switch context.Type {
	case EDIT:
		return f.Value == nil || !f.Value.Enabled()
	}
	return false
}

type FragmentCategoryConfig struct {
	Label      string
	IndexAttrs func(res *Resource, ctx *qor.Context) []interface{}
}

type FragmentConfig struct {
	Config    *Config
	NotInline bool
	Priority  int
	Category  *FragmentCategoryConfig
	IsLabel   string
	Is        bool
	Enabled   func(record fragment.FragmentedModelInterface, ctx *qor.Context) bool
	Available func(record fragment.FragmentedModelInterface, ctx *qor.Context) bool
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
	QTN         string
	fields      []*aorm.StructField
	query       string
	isURI       string
}

func (f *Fragment) FieldsCount() int {
	return f.fieldsCount
}

func (f *Fragment) Build() {
	if f.fieldsCount > 0 {
		return
	}

	f.QTN = f.Resource.FakeScope.QuotedTableName()

	f.buildFields()
	f.buildQuery()

	f.joinQuery = "JOIN " + f.QTN + " ON " + f.QTN +
		".id = " + f.Resource.ParentResource.FakeScope.QuotedTableName() + ".id"
	f.fieldsCount = len(f.fields)

	if f.Resource.Fragments != nil {
		f.Resource.Fragments.Build()
		f.fieldsCount += f.Resource.Fragments.columnsCount
	}

	if f.Config.Is {
		var param []string
		super := f.Resource
		for super.Fragment != nil {
			param = append(param, utils.ToParamString(super.PluralName))
			super = super.ParentResource
		}
		f.isURI = "/" + strings.Join(param, "/")

		index := super.Router.FindHandler("GET", P_INDEX).(*RouteHandler)
		newIndex := index.Clone()
		newIndex.Intercept(func(chain *Chain) {
			uri := chain.Context.Resource.GetContextIndexURI(chain.Context.Context)
			chain.Context.Breadcrumbs().Append(qor.NewBreadcrumb(uri, chain.Context.Resource.PluralLabelKey(), ""))
			chain.Context.PageTitle = f.Resource.PluralLabelKey()
			chain.Pass()
		})
		super.Router.Get(f.isURI, newIndex)
	}
}

func (f *Fragment) IsURI() string {
	return f.isURI
}

func (f *Fragment) buildFields() {
	fields := append([]*aorm.StructField{}, f.Resource.FakeScope.GetNonIgnoredStructFields()...)
	if !f.IsForm {
		for i, field := range fields {
			if field.Name == "FragmentEnabled" {
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
		columns[i] = f.QTN + "." + field.DBName
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

func (f *Fragment) Query() string {
	return f.query
}

func (f *Fragment) AllQuery() string {
	queries := []string{f.Query()}
	if f.Resource.Fragments != nil {
		f.Resource.Fragments.Walk(func(fr *Fragment) error {
			queries = append(queries, fr.Query())
			return nil
		})
	}
	return strings.Join(queries, ", ")
}

func (f *Fragment) JoinLeft(DB *aorm.DB) *aorm.DB {
	DB = DB.Joins("LEFT " + f.joinQuery)
	if f.Resource.Fragments != nil {
		return f.Resource.Fragments.JoinLeft(DB)
	}
	return DB
}

func (f *Fragment) Join(DB *aorm.DB) *aorm.DB {
	DB = DB.Joins(f.joinQuery)
	if f.Resource.Fragments != nil {
		return f.Resource.Fragments.JoinLeft(DB)
	}
	return DB
}

func (f *Fragment) Filter(DB *aorm.DB) *aorm.DB {
	super := f
	for super != nil {
		DB = DB.Where(super.QTN + ".fragment_enabled")
		super = super.Resource.ParentResource.Fragment
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

func (f *Fragment) Enabled(record fragment.FragmentedModelInterface, ctx *qor.Context) bool {
	if !f.IsForm {
		return true
	} else if f.Config.NotInline && record.GetFormFragment(f.ID) != nil {
		fv := record.GetFormFragment(f.ID)
		if ctx.Data().Get(CONTEXT_KEY).(*Context).Type == SHOW {
			if fv.Enabled() && (f.Config.Enabled == nil || f.Config.Enabled(record, ctx)) {
				return true
			}
			return false
		}
		return true
	}

	fv := record.GetFormFragment(f.ID)
	if ctx.Data().Get(CONTEXT_KEY).(*Context).Type == SHOW && (fv == nil || !fv.Enabled()) {
		return false
	}
	enabled := f.Config.Enabled == nil || f.Config.Enabled(record, ctx)
	return enabled
}

func (f *Fragment) FormRecordValue(record fragment.FragmentedModelInterface, ctx *qor.Context) *FormFragmentRecordState {
	value := record.GetFormFragment(f.ID)
	return &FormFragmentRecordState{f, f.Enabled(record, ctx), value, value == nil}
}

func (f *Fragment) FormGetOrNew(record fragment.FragmentedModelInterface, ctx *qor.Context) fragment.FormFragmentModelInterface {
	value := record.GetFormFragment(f.ID)
	if value == nil {
		value = f.Resource.NewStruct(ctx.Site).(fragment.FormFragmentModelInterface)
		record.SetFormFragment(f.ID, value)
	}
	value.SetSuper(record)
	return value
}

func (f *Fragment) GetOrNew(record fragment.FragmentedModelInterface, ctx *qor.Context) fragment.FragmentModelInterface {
	value := record.GetFragment(f.ID)
	if value == nil {
		value = f.Resource.NewStruct(ctx.Site).(fragment.FragmentModelInterface)
		record.SetFragment(f.ID, value)
	}
	value.SetSuper(record)
	return value
}

type fragmentActionArgument struct {
	Name string
}
package admin

import (
	"sort"

	"github.com/ecletus/core/resource"

	"strings"

	"database/sql"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/fragment"
	"github.com/ecletus/roles"
	"github.com/go-errors/errors"
	"github.com/moisespsena-go/aorm"
)

var NotFragmentableError = errors.New("Not fragmentable")
var FRAGMENT_KEY = PKG + ".fragment"

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
		_ = f.Walk(func(fr *Fragment) error {
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
	_ = f.Walk(func(fr *Fragment) error {
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

	if fr.Config.Is {
		var param []string
		super := fr.Resource
		for super.Fragment != nil {
			param = append(param, utils.ToParamString(super.PluralName))
			super = super.ParentResource
		}
		vf := res.ParentResource.FakeScope.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			r.SetFormFragment(r, fr.ID, value.(fragment.FormFragmentModelInterface))
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFormFragment(fr.ID)
			return v, v != nil
		}
		fr.isURI = "/" + strings.Join(param, "/")
		res.DefaultFilter(func(context *core.Context, db *aorm.DB) *aorm.DB {
			return db.AutoInlinePreload(res.Value)
		})
		super.Scheme.AddChild(fr.ID, &SchemeConfig{
			Visible: true,
			Setup: func(s *Scheme) {
				s.Categories = []string{"fragment"}
				s.SetI18nKey(res.PluralLabelKey())
				s.SchemeParam = utils.ToParamString(res.PluralName)
				s.DefaultFilter(func(context *core.Context, db *aorm.DB) *aorm.DB {
					db = db.InlinePreload(fr.ID, &aorm.InlinePreloadOptions{Join: aorm.JoinInner})
					return fr.Filter(db)
				})
				fr.scheme = s
				if fr.Config.SchemeSetup != nil {
					fr.Config.SchemeSetup(s)
				}
			},
		})
	} else if isForm {
		vf := res.ParentResource.FakeScope.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			r.SetFormFragment(r, fr.ID, value.(fragment.FormFragmentModelInterface))
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFormFragment(fr.ID)
			return v, v != nil
		}
	} else {
		vf := res.ParentResource.FakeScope.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			r.SetFragment(r, fr.ID, value.(fragment.FragmentModelInterface))
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFragment(fr.ID)
			return v, v != nil
		}
	}

	_ = res.ParentResource.OnDBAction(func(e *resource.DBEvent) {
		e.SetDB(e.DB().InlinePreload(fr.ID))
	}, resource.E_DB_ACTION_FIND_ONE.Before())

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
				result.SetFormFragment(result, id, value.(fragment.FormFragmentModelInterface))
			} else {
				result.SetFragment(result, id, value)
			}
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
}

func (f *Fragment) IsURI() string {
	return f.isURI
}

func (f *Fragment) buildFields() {
	fields := append([]*aorm.StructField{}, f.Resource.FakeScope.GetNonIgnoredStructFields()...)
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
		DB = DB.Where(aorm.IQ("{" + f.ID + "}.fragment_enabled"))
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

func (f *Fragment) Enabled(record fragment.FragmentedModelInterface, ctx *core.Context) bool {
	if !f.IsForm {
		return true
	} else if f.Config.NotInline && record.GetFormFragment(f.ID) != nil {
		fv := record.GetFormFragment(f.ID)
		if ctx.Data().Get(CONTEXT_KEY).(*Context).Type.Has(SHOW) {
			if fv.Enabled() && (f.Config.Enabled == nil || f.Config.Enabled(record, ctx)) {
				return true
			}
			return false
		}
		return true
	}

	fv := record.GetFormFragment(f.ID)
	if ctx.Data().Get(CONTEXT_KEY).(*Context).Type.Has(SHOW) && (fv == nil || !fv.Enabled()) {
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

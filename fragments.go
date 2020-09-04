package admin

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/fragment"
	"github.com/moisespsena-go/aorm"
)

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
		if fr.Resource.Config.Virtual {
			continue
		}
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

func (f *Fragments) WalkAll(cb func(fr *Fragment) error) (err error) {
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
		if fr := f.Fragments[id]; !fr.Resource.Config.Virtual {
			DB = fr.JoinLeft(DB)
		}
	}
	return DB
}

func (f *Fragments) Join(DB *aorm.DB) *aorm.DB {
	f.Build()
	for _, id := range f.Sorted {
		if fr := f.Fragments[id]; !fr.Resource.Config.Virtual {
			DB = fr.Join(DB)
		}
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

func (f *Fragments) Query(ctx *core.Context) (query string) {
	var queries []string
	f.Build()
	_ = f.Walk(func(fr *Fragment) error {
		if q := fr.AllQuery(ctx.DB()); q != "" {
			queries = append(queries, q)
		}
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

	if fr.Config.Mode.Cast() {
		var param []string
		super := fr.Resource
		for super.Fragment != nil {
			param = append(param, utils.ToParamString(super.PluralName))
			super = super.ParentResource
		}
		vf := res.ParentResource.ModelStruct.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			if value, ok := value.(fragment.FormFragmentModelInterface); ok {
				r.SetFormFragment(r, fr.ID, value.(fragment.FormFragmentModelInterface))
			}
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFormFragment(fr.ID)
			return v, v != nil
		}
		fr.isURI = "/" + strings.Join(param, "/")
		if !res.Config.Virtual {
			res.DefaultFilter(&DBFilter{
				Name: "admin:fragment:"+fr.ID+":auto_inline_preload",
				Handler: func(context *core.Context, db *aorm.DB) (DB *aorm.DB, err error) {
					return db.AutoInlinePreload(res.Value), nil
				},
			})
		}
		super.Scheme.AddChild(fr.ID, &SchemeConfig{
			Visible: true,
			Setup: func(s *Scheme) {
				s.Categories = []string{"fragment"}
				s.SetI18nKey(res.PluralLabelKey())
				s.SchemeParam = utils.ToParamString(res.PluralName)
				if !res.Config.Virtual {
					s.DefaultFilter(&DBFilter{
						Name: "admin:fragment:"+fr.ID+":inline_preload_join_inner",
						Handler: func(context *core.Context, db *aorm.DB) (DB *aorm.DB, err error) {
							db = db.InlinePreload(fr.ID, &aorm.InlinePreloadOptions{Join: aorm.JoinInner})
							return fr.Filter(db), nil
						},
					})
				}
				fr.scheme = s
				if fr.Config.SchemeSetup != nil {
					fr.Config.SchemeSetup(s)
				}
			},
		})
	} else if isForm {
		vf := res.ParentResource.ModelStruct.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			r.SetFormFragment(r, fr.ID, value.(fragment.FormFragmentModelInterface))
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFormFragment(fr.ID)
			return v, v != nil
		}
	} else {
		vf := res.ParentResource.ModelStruct.SetVirtualField(fr.ID, res.Value)
		vf.Setter = func(_ *aorm.VirtualField, recorde, value interface{}) {
			r := recorde.(fragment.FragmentedModelInterface)
			r.SetFragment(r, fr.ID, value.(fragment.FragmentModelInterface))
		}
		vf.Getter = func(_ *aorm.VirtualField, recorde interface{}) (value interface{}, ok bool) {
			v := recorde.(fragment.FragmentedModelInterface).GetFragment(fr.ID)
			return v, v != nil
		}
	}

	if !res.Config.Virtual {
		if cfg.LoadInFindMany {
			_ = res.ParentResource.OnDBAction(func(e *resource.DBEvent) {
				e.DB(e.DB().InlinePreload(fr.ID))
			}, resource.E_DB_ACTION_FIND_MANY.Before())
		}
		_ = res.ParentResource.OnDBAction(func(e *resource.DBEvent) {
			e.DB(e.DB().InlinePreload(fr.ID))
		}, resource.E_DB_ACTION_FIND_ONE.Before())
	}

	return fr
}

func (f *Fragments) Add(res *Resource, cfg *FragmentConfig) *Fragment {
	return f.add(res, false, cfg)
}

func (f *Fragments) AddForm(res *Resource, cfg *FragmentConfig) *Fragment {
	return f.add(res, true, cfg)
}

func (f *Fragments) ExtraFieldsScan(result fragment.FragmentedModelInterface, values []interface{}, set func(model *aorm.ModelStruct, result interface{}, low, hight int) interface{}) {
	low := 0
	f.extraFieldsScan(result, values, &low, set)
}

func (f *Fragments) extraFieldsScan(result fragment.FragmentedModelInterface, values []interface{}, low *int, set func(model *aorm.ModelStruct, result interface{}, low, hight int) interface{}) {
	f.Build()
	for _, id := range f.Sorted {
		fr := f.Fragments[id]

		// FragmentEnabledAttribute field. If is nil, is not defined.
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
			value := set(fr.Resource.ModelStruct, fr.Resource.Value, *low, *low+fr.fieldsCount).(fragment.FragmentModelInterface)
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

package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/fragment"
	"github.com/ecletus/helpers"
	"github.com/moisespsena-go/aorm"
	errwrap "github.com/moisespsena/go-error-wrap"
)

func (res *Resource) AddFragment(value fragment.FragmentModelInterface) *Resource {
	return res.AddFragmentConfig(value, &FragmentConfig{})
}

func (res *Resource) AddFragmentConfig(value fragment.FragmentModelInterface, cfg *FragmentConfig) *Resource {
	if _, ok := res.Value.(fragment.FragmentedModelInterface); !ok {
		panic(NotFragmentableError)
	}

	if cfg == nil {
		cfg = &FragmentConfig{}
	}
	if cfg.Config == nil {
		cfg.Config = &Config{}
	}
	if cfg.Config.Sub == nil {
		cfg.Config.Sub = &SubConfig{}
	}

	cfg.Config.Singleton = true
	cfg.Config.Sub.Parent = res
	cfg.Config.DisableFormID = true
	cfg.Config.Sub.ParentFieldName = "ID"

	_, isForm := value.(fragment.FormFragmentModelInterface)

	setup := cfg.Config.Setup

	cfg.Config.Setup = func(fragRes *Resource) {
		fragRes.SetMeta(&Meta{Name: "ID", Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}}, true)
		if isForm {
			meta := &Meta{
				Name:              AttrFragmentEnabled,
				SectionNotAllowed: true,
			}
			if cfg.Is {
				meta.SkipDefaultLabel = true
				if cfg.IsLabel != "" {
					meta.Label = cfg.IsLabel
				} else {
					meta.Label = fragRes.SingularLabelKey()
				}
				if meta.Label[0] == '.' {
					meta.Label = fragRes.I18nPrefix + meta.Label
				}
				meta.Enabled = func(recorde interface{}, context *Context, meta *Meta) bool {
					if context.Type.Has(SHOW) {
						return false
					}
					return true
				}
			} else {
				meta.SkipDefaultLabel = true
				meta.Label = fragRes.SingularLabelKey()
			}
			fragRes.SetMeta(meta, true)
			res.Fragments.AddForm(fragRes, cfg)
		} else {
			res.Fragments.Add(fragRes, cfg)
		}

		fragRes.Fragment.Build()

		if setup != nil {
			setup(fragRes)
		}
	}

	if !isForm {
		cfg.Config.Invisible = true
	} else {
		if cfg.Is || cfg.NotInline {
			old := cfg.Config.MenuEnabled
			if old == nil {
				old = func(menu *Menu, ctx *Context) bool { return true }
			}
			cfg.Config.MenuEnabled = func(menu *Menu, ctx *Context) bool {
				if r, ok := ctx.Result.(fragment.FragmentedModelInterface); ok {
					if f := r.GetFormFragment(menu.Resource.Fragment.ID); f != nil {
						return f.Enabled() && (menu.Resource.Fragment.Config.Is || menu.Resource.Fragment.Config.NotInline)
					} else if menu == menu.Resource.defaultMenu {
						return false
					}
				}
				return old(menu, ctx)
			}
		}
	}

	fragRes := res.AddResourceConfig(value, cfg.Config)

	if len(res.Fragments.Fragments) == 1 {
		_ = res.GetAdmin().OnDone(func(e *AdminEvent) {
			res.Fragments.Build()
		})
		_ = res.OnDBActionE(func(e *resource.DBEvent) (err error) {
			context := e.Context
			if v := context.Data().Get("skip.fragments"); v == nil {
				r := e.Result().(fragment.FragmentedModelInterface)
				for id, fr := range r.GetFragments() {
					fragRes := res.Fragments.Get(id).Resource
					if fr.GetID() == "" {
						fr.SetID(r.GetID())
					}
					if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
						return errwrap.Wrap(err, "Fragment "+id)
					}
				}
				for id, fr := range r.GetFormFragments() {
					fragRes := res.Fragments.Get(id).Resource
					if fr.GetID() == "" {
						fr.SetID(r.GetID())
					}
					if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
						return errwrap.Wrap(err, "Fragment "+id)
					}
				}
			}
			return nil
		}, resource.E_DB_ACTION_SAVE.After())

		_ = res.Scheme.OnDBActionE(func(e *resource.DBEvent) (err error) {
			context := e.Context
			if v := context.Data().Get("skip.fragments"); v == nil {
				DB := context.DB
				fields, query := res.Fragments.Fields(), res.Fragments.Query()
				DB = DB.ExtraSelectFieldsSetter(
					PKG+".fragments",
					func(result interface{}, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
						res.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
					}, fields, query)
				DB = res.Fragments.JoinLeft(DB)
				context.SetDB(DB)
			}
			return nil
		}, resource.BEFORE|resource.E_DB_ACTION_FIND_MANY|resource.E_DB_ACTION_FIND_ONE)

		res.FakeScope.GetModelStruct().BeforeRelatedCallback(func(fromScope *aorm.Scope, toScope *aorm.Scope, DB *aorm.DB, fromField *aorm.Field) *aorm.DB {
			fields, query := res.Fragments.Fields(), res.Fragments.Query()
			DB = DB.ExtraSelectFieldsSetter(
				PKG+".fragments",
				func(result interface{}, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
					res.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
				}, fields, query)
			DB = res.Fragments.JoinLeft(DB)
			return DB
		})
	}

	metaName := fragRes.Fragment.ID

	if !isForm {
		fieldsNames := fragRes.Fragment.FieldsNames()

		for _, fieldName := range fieldsNames {
			meta := NewMetaProxy(fieldName, fragRes.Meta(&Meta{Name: fieldName}), func(meta *Meta, recorde interface{}) interface{} {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				return fragmentedRecorde.GetFragment(frag.ID)
			})
			meta.Fragment = fragRes.Fragment
			meta.Resource = res
			meta.NewValuer(func(meta *Meta, old MetaValuer, recorde interface{}, context *core.Context) interface{} {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				value := frag.GetOrNew(fragmentedRecorde, context)
				return meta.ProxyTo.GetValuer()(value, context)
			})
			meta.NewSetter(func(meta *Meta, old MetaSetter, recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				value := frag.GetOrNew(fragmentedRecorde, context)
				return meta.ProxyTo.GetSetter()(value, metaValue, context)
			})
			meta = res.Meta(meta)
		}

		fieldsNamesInterface := helpers.StringsToInterfaces(fieldsNames)
		res.EditAttrs(append([]interface{}{res.EditAttrs()}, fieldsNamesInterface...)...)
		res.ShowAttrs(append([]interface{}{res.ShowAttrs()}, fieldsNamesInterface...)...)
	} else {
		res.Meta(&Meta{
			Name: metaName,
			Type: "fragment",
			Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
				if recorde != nil {
					return fragRes.Fragment.Enabled(recorde.(fragment.FragmentedModelInterface), context.Context)
				}
				return false
			},
			ContextMetas: func(record interface{}, ctx *Context) []*Meta {
				return fragRes.ConvertSectionToMetas(fragRes.ContextSections(ctx, record))
			},
			Setter: func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if _, ok := res.Fragments.Fragments[metaValue.Name]; !ok {
					return nil
				}
				value := fragRes.Fragment.FormGetOrNew(recorde.(fragment.FragmentedModelInterface), context)
				err := resource.DecodeToResource(fragRes, value, metaValue.MetaValues, context).Start()
				return err
			},
			Valuer: func(recorde interface{}, context *core.Context) interface{} {
				r := recorde.(fragment.FragmentedModelInterface)
				value := r.GetFormFragment(fragRes.Fragment.ID)
				isNil := value == nil
				if isNil {
					value = fragRes.NewStruct(context.Site).(fragment.FormFragmentModelInterface)
				}
				return &FormFragmentRecordState{
					fragRes.Fragment,
					fragRes.Fragment.Enabled(r, context),
					value,
					isNil,
				}
			},
		})

		var hasEditMeta bool
		if len(res.editSections) > 0 {
		root:
			for _, sec := range res.editSections {
				for _, row := range sec.Rows {
					for _, col := range row {
						if col == metaName {
							hasEditMeta = true
							break root
						}
					}
				}
			}
		}

		if !hasEditMeta {
			res.EditAttrs(res.EditAttrs(), metaName)
		}

		if len(res.showSections) == 0 {
			res.ShowAttrs(res.EditAttrs(), metaName)
		} else {
			var hasShowMeta bool
		root2:
			for _, sec := range res.showSections {
				for _, row := range sec.Rows {
					for _, col := range row {
						if col == metaName {
							hasShowMeta = true
							break root2
						}
					}
				}
			}

			if !hasShowMeta {
				res.ShowAttrs(res.ShowAttrs(), metaName)
			}
		}
	}

	return fragRes
}

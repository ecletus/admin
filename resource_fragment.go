package admin

import (
	"github.com/ecletus/fragment"
	"github.com/ecletus/helpers"
	errwrap "github.com/moisespsena-go/error-wrap"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) AddFragment(value fragment.FragmentModelInterface) *Resource {
	return this.AddFragmentConfig(value, &FragmentConfig{})
}

func (this *Resource) AddFragmentConfig(value fragment.FragmentModelInterface, cfg *FragmentConfig) *Resource {
	_ = this.Value.(fragment.FragmentedModelInterface)
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
	cfg.Config.Sub.Parent = this
	cfg.Config.DisableFormID = true
	cfg.Config.Sub.ParentFieldName = "ID"

	if len(cfg.Schemes) == 0 {
		cfg.Schemes = append(cfg.Schemes, this.Scheme)
		for _, scheme := range this.Scheme.Children {
			cfg.Schemes = append(cfg.Schemes, scheme)
		}
	}

	_, isForm := value.(fragment.FormFragmentModelInterface)

	setup := cfg.Config.Setup

	cfg.Config.Setup = func(fragRes *Resource) {
		if !this.Singleton && !fragRes.Config.Virtual {
			fragRes.SetMeta(&Meta{Name: "ID", Type: "-"})
		}
		if isForm {
			meta := &Meta{
				Name:              AttrFragmentEnabled,
				SectionNotAllowed: true,
			}
			if cfg.Mode.Cast() {
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
			fragRes.SetMeta(meta).DisableSiblingsRequirement = resource.SiblingsRequirementCheckDisabledOnTrue
			this.Fragments.AddForm(fragRes, cfg)
		} else {
			this.Fragments.Add(fragRes, cfg)
		}

		fragRes.Fragment.Build()

		if setup != nil {
			setup(fragRes)
		}
	}

	if isForm {
		old := cfg.Config.MenuEnabled
		if old == nil {
			old = func(menu *Menu, ctx *Context) bool {
				return true
			}
		}
		cfg.Config.MenuEnabled = func(menu *Menu, ctx *Context) bool {
			if r, ok := ctx.Result.(fragment.FragmentedModelInterface); ok {
				if f := r.GetFormFragment(menu.Resource.Fragment.ID); f != nil {
					return f.Enabled() && (menu.Resource.Fragment.Config.Mode.Cast() || !menu.Resource.Fragment.Config.Mode.Inline())
				} else if menu == menu.Resource.defaultMenu {
					return false
				}

				return old(menu, ctx)
			}
			return false
		}
	} else {
		cfg.Config.Invisible = true
	}

	fragRes := this.AddResourceConfig(value, cfg.Config)

	for _, scheme := range cfg.Schemes {
		if !scheme.hasFragments {
			scheme.hasFragments = true
			_ = scheme.OnDBActionE(func(e *resource.DBEvent) (err error) {
				context := e.Context
				if v := context.Value("skip.fragments"); v == nil {
					r := e.Result().(fragment.FragmentedModelInterface)
					for id, fr := range r.GetFragments() {
						fragRes := this.Fragments.Get(id).Resource
						if ID := aorm.IdOf(fr); ID.IsZero() {
							aorm.IdOf(r).SetTo(fr)
						}
						if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
							return errwrap.Wrap(err, "Fragment "+id)
						}
					}
					for id, fr := range r.GetFormFragments() {
						fragRes := this.Fragments.Get(id).Resource
						if ID := aorm.IdOf(fr); ID.IsZero() {
							ID, _ = aorm.CopyIdTo(aorm.IdOf(r), ID)
							ID.SetTo(fr)
						}
						if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
							return errwrap.Wrap(err, "Fragment "+id)
						}
					}
				}
				return nil
			}, resource.E_DB_ACTION_UPDATE.After())
		}
	}

	if len(this.Fragments.Fragments) == 1 {
		_ = this.GetAdmin().OnDone(func(e *AdminEvent) {
			this.Fragments.Build()
		})

		this.ModelStruct.BeforeRelatedCallback(func(fromScope *aorm.Scope, toScope *aorm.Scope, DB *aorm.DB, fromField *aorm.Field) *aorm.DB {
			fields, query := this.Fragments.Fields(), this.Fragments.Query(core.ContextFromDB(DB))
			DB = DB.ExtraSelectFieldsSetter(
				PKG+".fragments",
				func(result interface{}, values []interface{}, set func(model *aorm.ModelStruct, result interface{}, low, hight int) interface{}) {
					this.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
				}, fields, query)
			DB = this.Fragments.JoinLeft(DB)
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
			meta.Resource = this
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
			meta = this.Meta(meta)
		}

		fieldsNamesInterface := helpers.StringsToInterfaces(fieldsNames)
		if len(cfg.Sections) > 0 {
			fieldsNamesInterface = []interface{}{}
			for _, sec := range cfg.Sections {
				fieldsNamesInterface = append(fieldsNamesInterface, sec)
			}
		}
		this.EditAttrs(append([]interface{}{this.EditAttrs()}, fieldsNamesInterface...)...)
		this.ShowAttrs(append([]interface{}{this.ShowAttrs()}, fieldsNamesInterface...)...)
	} else if !fragRes.Config.Virtual {
		this.Meta(&Meta{
			Name: metaName,
			Type: "fragment",
			Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
				if recorde == nil {
					return true
				}

				if _, ok := recorde.(fragment.FragmentedModelInterface); ok {
					return fragRes.Fragment.Enabled(recorde.(fragment.FragmentedModelInterface), context.Context)
				}
				return false
			},
			ContextMetas: func(record interface{}, ctx *Context) []*Meta {
				f := fragRes.Fragment
				if f.IsForm {
					if ctx.Type.Has(NEW) {
						return fragRes.ConvertSectionToMetas([]*Section{{Rows: [][]string{{AttrFragmentEnabled}}}})
					}
				}
				return fragRes.ConvertSectionToMetas(fragRes.ContextSections(ctx, record))
			},
			Setter: func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if _, ok := this.Fragments.Fragments[metaValue.Name]; !ok {
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
		if len(this.editSections) > 0 {
		root:
			for _, sec := range this.editSections {
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
			this.EditAttrs(this.EditAttrs(), metaName)
		}

		if len(this.showSections) == 0 {
			this.ShowAttrs(this.EditAttrs(), metaName)
		} else {
			var hasShowMeta bool
		root2:
			for _, sec := range this.showSections {
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
				this.ShowAttrs(this.ShowAttrs(), metaName)
			}
		}
	}

	return fragRes
}

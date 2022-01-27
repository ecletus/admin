package admin

import (
	"github.com/ecletus/roles"
	tag_scanner "github.com/unapu-go/tag-scanner"
)

type SectionsContext struct {
	Ctx    *Context
	Record interface{}
}

type SchemeSectionsProvider struct {
	Name           string
	sections       Sections
	setMiddlewares []func(sections Sections) Sections
	Parent         *SchemeSectionsProvider
	GetContextFunc func(ctx *SectionsContext) []*Section
	Excludes       tag_scanner.Set
}

func (s *SchemeSectionsProvider) Exclude(names ...string) {
	s.Excludes.Add(names...)
}

func (s *SchemeSectionsProvider) Update(do func(sections Sections) Sections) *SchemeSectionsProvider {
	s.sections = do(s.MustSections())
	return s
}

func (s *SchemeSectionsProvider) SetMiddleware(md ...func(sections Sections) Sections) *SchemeSectionsProvider {
	s.setMiddlewares = append(s.setMiddlewares, md...)
	return s
}

func (s *SchemeSectionsProvider) SetSections(sections Sections) *SchemeSectionsProvider {
	for _, md := range s.setMiddlewares {
		sections = md(sections)
	}
	s.sections = sections
	return s
}

func (s *SchemeSectionsProvider) GetSections() (name string, sections Sections) {
	p := s
	for p != nil {
		if p.sections != nil {
			if len(s.Excludes) > 0 {
				return p.Name, sections.Exclude(s.Excludes.Strings()...)
			}
			return p.Name, p.sections
		}
		p = p.Parent
	}
	return
}

func (s *SchemeSectionsProvider) Sections(cb func(pv *SchemeSectionsProvider, s *Section) (cont bool)) {
	p := s
	for p != nil {
		if p.sections != nil {
			if len(s.Excludes) > 0 {
				for _, s := range p.sections.Exclude(s.Excludes.Strings()...) {
					if !cb(p, s) {
						return
					}
				}
			} else {
				for _, s := range p.sections {
					if !cb(p, s) {
						return
					}
				}
			}
			return
		}
		p = p.Parent
	}
	return
}

func (s *SchemeSectionsProvider) GetSectionsCopy() (_ string, sections Sections) {
	name, secs := s.GetSections()
	sections = make(Sections, len(secs))
	copy(sections, secs)
	return name, sections
}

func (s *SchemeSectionsProvider) MustSections() (sections Sections) {
	_, sections = s.GetSections()
	return
}

func (s *SchemeSectionsProvider) MetasNames() (names []string) {
	s.Sections(func(pv *SchemeSectionsProvider, s *Section) (cont bool) {
		s.MetasNamesCb(func(name string) {
			names = append(names, name)
		})
		return true
	})
	return
}

func (s *SchemeSectionsProvider) Metas(cb func(meta *Meta) (cont bool), getter ...func(res *Resource, name string) *Meta) {
	s.Sections(func(_ *SchemeSectionsProvider, s *Section) (cont bool) {
		s.Metas(func(meta *Meta) bool {
			cont = cb(meta)
			return cont
		}, getter...)
		return cont
	})
	return
}

func (s *SchemeSectionsProvider) GetMetas(getter ...func(res *Resource, name string) *Meta) (metas []*Meta) {
	for _, s := range s.MustSections() {
		s.Metas(func(meta *Meta) bool {
			metas = append(metas, meta)
			return true
		}, getter...)
	}
	return
}

func (s *SchemeSectionsProvider) MustSectionsCopy() (sections Sections) {
	_, sections = s.GetSectionsCopy()
	return
}

func (s *SchemeSectionsProvider) GetContextSections(ctx *SectionsContext) (name string, sections Sections) {
	p := s
	for p != nil {
		if p.GetContextFunc != nil {
			return p.Name, p.GetContextFunc(ctx)
		} else if len(p.sections) > 0 {
			return p.Name, p.sections
		}
		p = p.Parent
	}
	return
}

func (s *SchemeSectionsProvider) MustContextSections(ctx *SectionsContext) (sections Sections) {
	_, sections = s.GetContextSections(ctx)
	return
}

func (s *SchemeSectionsProvider) MustContextSectionsCopy(ctx *SectionsContext) (sections Sections) {
	_, secs := s.GetContextSections(ctx)
	sections = make(Sections, len(secs))
	copy(sections, secs)
	return
}

func (s *SchemeSectionsProvider) IsSet() bool {
	if s == nil {
		return false
	}
	return s.sections != nil
}

func (s *SchemeSectionsProvider) IsSetI() bool {
	p := s
	for p != nil {
		if p.IsSet() {
			return true
		}
		p = p.Parent
	}
	return false
}

type CRUDSchemeSectionsLayout struct {
	Name string

	Index,
	New,
	Edit,
	Show *SchemeSectionsProvider

	Custom map[string]*SchemeSectionsProvider
}

func (s *CRUDSchemeSectionsLayout) Exclude(names ...string) {
	s.Index.Exclude(names...)
	s.New.Exclude(names...)
	s.Edit.Exclude(names...)
	s.Show.Exclude(names...)
	for _, c := range s.Custom {
		c.Exclude(names...)
	}
}

func (l *CRUDSchemeSectionsLayout) MetasNamesCb(cb func(name string)) {
	var names = map[string]interface{}{}
	for _, s := range l.All() {
		s.MustSections().MetasNamesCb(func(name string) {
			if _, ok := names[name]; !ok {
				cb(name)
				names[name] = nil
			}
			return
		})
	}
}

func (l *CRUDSchemeSectionsLayout) Get(name string) *SchemeSectionsProvider {
	switch name {
	case "index":
		return l.Index
	case "new":
		return l.New
	case "edit":
		return l.Edit
	case "show":
		return l.Show
	default:
		return l.Custom[name]
	}
}

func (l *CRUDSchemeSectionsLayout) Cruds() []*SchemeSectionsProvider {
	return []*SchemeSectionsProvider{
		l.Show,
		l.Edit,
		l.New,
		l.Index,
	}
}

func (l *CRUDSchemeSectionsLayout) All() (ret []*SchemeSectionsProvider) {
	ret = []*SchemeSectionsProvider{
		l.Show,
		l.Edit,
		l.New,
		l.Index,
	}
	for _, c := range l.Custom {
		ret = append(ret, c)
	}
	return
}

func NewCRUDSchemeSectionsLayout(name string, from ...*CRUDSchemeSectionsLayout) *CRUDSchemeSectionsLayout {
	l := &CRUDSchemeSectionsLayout{
		Name:   name,
		Edit:   &SchemeSectionsProvider{Name: name + "@edit"},
		Index:  &SchemeSectionsProvider{Name: name + "@index"},
		New:    &SchemeSectionsProvider{Name: name + "@new"},
		Show:   &SchemeSectionsProvider{Name: name + "@show"},
		Custom: map[string]*SchemeSectionsProvider{},
	}

	if len(from) == 1 {
		f := from[0]

		l.Edit.Parent = f.Edit
		l.Index.Parent = f.Index
		l.New.Parent = f.New
		l.Show.Parent = f.Show
		for name, prov := range f.Custom {
			l.Custom[name] = &SchemeSectionsProvider{Parent: prov}
		}
	}

	return l
}

func (this *CRUDSchemeSectionsLayout) Update(do func(sections Sections) Sections) *CRUDSchemeSectionsLayout {
	this.Index.Update(do)
	this.New.Update(do)
	this.Edit.Update(do)
	this.Show.Update(do)
	return this
}

type DefaultSchemeSectionsLayouts struct {
	Default,
	Inline *SectionsLayout

	Layouts *SchemeSectionsLayouts
}

func NewDefaultSchemeSectionsLayout(layouts *SchemeSectionsLayouts) *DefaultSchemeSectionsLayouts {
	return &DefaultSchemeSectionsLayouts{
		Layouts: layouts,
		Default: layouts.Layouts[SectionLayoutDefault],
		Inline:  layouts.Layouts[SectionLayoutDefault+"."+SectionLayoutInline],
	}
}

func (this *DefaultSchemeSectionsLayouts) Only(name string) *DefaultSchemeSectionsLayouts {
	return NewDefaultSchemeSectionsLayout(NewSchemeSectionsLayouts(name, NewSchemeSectionsLayoutsOptions{
		Default:       this.Layouts.Layouts[name],
		DefaultInline: this.Layouts.Layouts[name+"."+SectionLayoutInline],
	}))
}

func (this *DefaultSchemeSectionsLayouts) MakeChild(name string) *DefaultSchemeSectionsLayouts {
	return NewDefaultSchemeSectionsLayout(NewSchemeSectionsLayouts(name, NewSchemeSectionsLayoutsOptions{From: this.Layouts}))
}

type SectionsLayout struct {
	Name string
	Screen,
	Print *CRUDSchemeSectionsLayout
}

func (s *SectionsLayout) Get(fieldName string) *CRUDSchemeSectionsLayout {
	switch fieldName {
	case "screen", "Screen":
		return s.Screen
	case "print", "Print":
		return s.Print
	}
	return nil
}

func (s *SectionsLayout) Set(fieldName string, value *CRUDSchemeSectionsLayout) {
	switch fieldName {
	case "screen", "Screen":
		s.Screen = value
	case "print", "Print":
		s.Print = value
	}
}

func (s *SectionsLayout) Exclude(names ...string) {
	s.Screen.Exclude(names...)
	s.Print.Exclude(names...)
}

type NewSchemeSectionsLayoutsOptions struct {
	From                   *SchemeSectionsLayouts
	Default, DefaultInline *SectionsLayout
	DefaultProvider        *SchemeSectionsProvider
}

type SchemeSectionsLayouts struct {
	Layouts map[string]*SectionsLayout
}

func NewSchemeSectionsLayouts(name string, opts ...NewSchemeSectionsLayoutsOptions) (l *SchemeSectionsLayouts) {
	l = &SchemeSectionsLayouts{Layouts: map[string]*SectionsLayout{}}
	var opt NewSchemeSectionsLayoutsOptions
	for _, opt = range opts {
	}

	if opt.From == nil {
		if opt.DefaultProvider == nil {
			opt.DefaultProvider = &SchemeSectionsProvider{Name: "default_provider"}
		}

		if opt.Default == nil {
			opt.Default = &SectionsLayout{}
		}

		if opt.Default.Screen == nil {
			opt.Default.Screen = &CRUDSchemeSectionsLayout{
				Name:   name + "#default/screen",
				Index:  &SchemeSectionsProvider{Name: "#index", Parent: opt.DefaultProvider},
				New:    &SchemeSectionsProvider{Name: "#new", Parent: opt.DefaultProvider},
				Edit:   &SchemeSectionsProvider{Name: "#edit", Parent: opt.DefaultProvider},
				Show:   &SchemeSectionsProvider{Name: "#show", Parent: opt.DefaultProvider},
				Custom: map[string]*SchemeSectionsProvider{},
			}
		}

		if opt.Default.Print == nil {
			opt.Default.Print = opt.Default.Screen
		}

		if opt.DefaultInline == nil {
			opt.DefaultInline = &SectionsLayout{}
			opt.DefaultInline.Screen = &CRUDSchemeSectionsLayout{
				Name:   name + "#inline/screen",
				Index:  &SchemeSectionsProvider{Name: "#index", Parent: opt.Default.Screen.Index},
				New:    &SchemeSectionsProvider{Name: "#new", Parent: opt.Default.Screen.New},
				Edit:   &SchemeSectionsProvider{Name: "#edit", Parent: opt.Default.Screen.Edit},
				Show:   &SchemeSectionsProvider{Name: "#show", Parent: opt.Default.Screen.Show},
				Custom: map[string]*SchemeSectionsProvider{},
			}
		}

		if opt.DefaultInline.Print == nil {
			opt.DefaultInline.Print = &CRUDSchemeSectionsLayout{
				Name:   name + "#inline/print",
				Index:  &SchemeSectionsProvider{Name: "#index", Parent: opt.DefaultInline.Screen.Index},
				New:    &SchemeSectionsProvider{Name: "#new", Parent: opt.DefaultInline.Screen.New},
				Edit:   &SchemeSectionsProvider{Name: "#edit", Parent: opt.DefaultInline.Screen.Edit},
				Show:   &SchemeSectionsProvider{Name: "#show", Parent: opt.DefaultInline.Screen.Show},
				Custom: map[string]*SchemeSectionsProvider{},
			}
		}
		l.Layouts[SectionLayoutDefault] = &SectionsLayout{name + "#default", opt.Default.Screen, opt.Default.Print}
		l.Layouts[SectionLayoutDefault+"."+SectionLayoutInline] = &SectionsLayout{name + "#inline", opt.DefaultInline.Screen, opt.DefaultInline.Print}
	} else {
		for _, key := range []string{
			SectionLayoutDefault,
			SectionLayoutDefault + "." + SectionLayoutInline,
		} {
			from := opt.From.Layouts[key]
			layout := &SectionsLayout{Name: name + "#" + key}

			for _, pair := range [][2]**CRUDSchemeSectionsLayout{{&layout.Screen, &from.Screen}, {&layout.Print, &from.Print}} {
				dst, from := pair[0], *pair[1]
				*dst = NewCRUDSchemeSectionsLayout(key, &CRUDSchemeSectionsLayout{
					Name:   key,
					Index:  &SchemeSectionsProvider{Name: name + ">" + from.Index.Name, Parent: from.Index},
					New:    &SchemeSectionsProvider{Name: name + ">" + from.New.Name, Parent: from.New},
					Edit:   &SchemeSectionsProvider{Name: name + ">" + from.Edit.Name, Parent: from.Edit},
					Show:   &SchemeSectionsProvider{Name: name + ">" + from.Show.Name, Parent: from.Show},
					Custom: map[string]*SchemeSectionsProvider{},
				})
				for key, custom := range from.Custom {
					(*dst).Custom[key] = &SchemeSectionsProvider{Parent: custom, Name: key + ">" + custom.Name}
				}
			}
			l.Layouts[key] = layout
		}
	}

	return l
}

func (this *Scheme) EditSections(context *Context, record interface{}) (sections Sections) {
	sections = this.GetSectionsProvider(context, EDIT).MustContextSectionsCopy(&SectionsContext{
		Ctx:    context,
		Record: record,
	})

	if this == this.Resource.Scheme && this.Resource.Fragment != nil {
		sections = append(sections, &Section{Resource: this.Resource, Rows: [][]interface{}{{AttrFragmentEnabled}}})
	}
	return sections.Allowed(record, context, roles.Update)
}

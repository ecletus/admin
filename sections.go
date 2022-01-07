package admin

import (
	"fmt"
	"strings"

	"github.com/ecletus/core/resource"
	"github.com/ecletus/roles"
)

type Sections []*Section

func (s Sections) AddPrefix(prefix string) []*Section {
	items := make([]*Section, len(s))
	for i, section := range s {
		items[i] = section.AddPrefix(prefix)
	}
	return items
}

func (s Sections) Filter(f *SectionsFilter) (secs Sections) {
	if len(f.Exclude) == 0 && f.Excludes == nil && len(f.Include) == 0 {
		return s
	}

	var changed bool

	for i, sec := range s {
		if len(sec.Rows) == 0 {
			changed = true
			continue
		}
		if sec := sec.Filter(f); sec != s[i] {
			changed = true
			if sec != nil {
				secs = append(secs, sec)
			}
		} else {
			secs = append(secs, sec)
		}
	}

	if !changed {
		return s
	}

	return
}

func (this Sections) Exclude(names ...string) Sections {
	return this.Filter(new(SectionsFilter).SetExcludes(names...))
}

func (this Sections) Only(names ...string) Sections {
	return this.Filter(new(SectionsFilter).SetInclude(names...))
}

func (this Sections) ReplaceMeta(from, to string) Sections {
	for _, s := range this {
		s.ReplaceMeta(from, to)
	}
	return this
}

func (this Sections) ReplaceMetas(pairs [][]string) Sections {
	for _, s := range this {
		s.ReplaceMetas(pairs)
	}
	return this
}

func (this Sections) Allowed(record interface{}, context *Context, modes ...roles.PermissionMode) Sections {
	var newSections Sections
	for _, section := range this {
		newSection := section.CopyWoRows()
		for _, row := range section.Rows {
			func() {
				var columns []interface{}
				for _, column := range row {
					switch column := column.(type) {
					case string:
						meta := section.Resource.GetMeta(column)
						if meta != nil {
							if meta.SectionNotAllowed {
								continue
							}

							if !meta.IsEnabled(record, context, meta, len(modes) == 1 && modes[0] == roles.Read) {
								continue
							}

							if context.HasAnyPermissionDefault(meta, true, modes...) {
								if len(columns) == 0 {
									if sec := meta.Tags.Section(); sec != nil {
										var add = func() {
											// add current
											newSections = append(newSections, &Section{
												Resource:     section.Resource,
												Title:        sec.Title,
												Help:         sec.Help,
												ReadOnlyHelp: sec.ReadOnlyHelp,
												Rows:         [][]interface{}{{column}},
											})
										}
										if newSection.Title != "" {
											defer add()
										} else {
											// split section
											if len(newSection.Rows) > 0 {
												// add previous
												newSections = append(newSections, newSection)
											}
											add()
											// create new empty section
											newSection = section.CopyWoRows()
										}
										continue
									}
								}
								columns = append(columns, column)
							}
						}
					case *Section:
						secs := Sections{column}.Allowed(record, context, modes...)
						if len(secs) == 1 {
							columns = append(columns, secs[0])
						}
					}
				}
				if len(columns) > 0 {
					newSection.Rows = append(newSection.Rows, columns)
				}
			}()
		}

		if len(newSection.Rows) > 0 {
			newSections = append(newSections, newSection)
		}
	}

	return newSections
}

func (this Sections) MetasNamesCb(cb func(name string)) {
	for _, section := range this {
		section.MetasNamesCb(cb)
	}
}

func (this Sections) MetasNames() (names []string) {
	this.MetasNamesCb(func(name string) {
		names = append(names, name)
	})
	return
}

func (this Sections) Metas(cb func(meta *Meta) (cont bool), getter ...func(res *Resource, name string) *Meta) {
	for _, section := range this {
		section.Metas(cb, getter...)
	}
}

func (this Sections) ToMetas(getter ...func(res *Resource, name string) *Meta) (metas []*Meta) {
	this.Metas(func(meta *Meta) bool {
		metas = append(metas, meta)
		return true
	}, getter...)
	return
}

func (this Sections) ToMetaors(getter ...func(res *Resource, name string) *Meta) (metas []resource.Metaor) {
	this.Metas(func(meta *Meta) bool {
		metas = append(metas, meta)
		return true
	}, getter...)
	return
}

func (this Sections) Uniquefy() Sections {
	return this.Filter(new(SectionsFilter).Unique())
}

func RemoveAttrsFromSections(attrs []string, getSeter ...func(val ...interface{}) Sections) {
	f := new(SectionsFilter).SetExcludes(attrs...)
	for _, gs := range getSeter {
		gs(gs().Filter(f))
	}
}

func ReplaceAttrsFromSections(attrs [][]string, getSeter ...func(val ...interface{}) Sections) {
	for _, gs := range getSeter {
		gs(gs().ReplaceMetas(attrs))
	}
}

type SectionsAttribute struct {
	Sections        *DefaultSchemeSectionsLayouts
	Resource        *Resource
	AllSectionsFunc func() Sections
}

func (this *SectionsAttribute) DefaultSections() *SectionsLayout {
	return this.Sections.Default
}

func (this *SectionsAttribute) GetSectionsByName(name string) *SectionsLayout {
	if name == "" || name == SectionLayoutDefault {
		return this.DefaultSections()
	}
	return this.Sections.Layouts.Layouts[name]
}

func (this *SectionsAttribute) GetSectionsProvider(ctx *Context, typ ContextType, provName ...string) *SchemeSectionsProvider {
	var (
		name = ctx.SectionLayout
		sl   *SectionsLayout
	)

	if name == "" {
		sl = this.DefaultSections()
		name = SectionLayoutDefault
	} else {
		sl = this.Sections.Layouts.Layouts[name]
	}

	if ctx.Type.Has(INLINE) {
		typ |= INLINE
	}

	if ctx.Type.Has(PRINT) {
		typ |= PRINT
	}

	if typ.Has(INLINE) {
		if l2 := this.Sections.Layouts.Layouts[name+".inline"]; l2 != nil {
			sl = l2
		}
	}

	l := sl.Screen
	if typ.Has(PRINT) {
		l = sl.Print
	}

	if len(provName) == 1 && provName[0] != "" {
		return l.Custom[provName[0]]
	}

	if typ.Has(INDEX) {
		return l.Index
	}
	if typ.Has(NEW) {
		return l.New
	}
	if typ.Has(EDIT) {
		return l.Edit
	}
	if typ.Has(SHOW) {
		return l.Show
	}
	return nil
}

func (this *SectionsAttribute) generateSections(values ...interface{}) (sections Sections) {
	var (
		hasColumns      = map[string]interface{}{}
		excludedColumns []string
	)

	// Reverse values to make the last one as a key one
	// e.g. Name, Code, -Name (`-Name` will get first and will skip `Name`)
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]
		if s, ok := value.(*Section); ok {
			if s.Resource == nil {
				s.Resource = this.Resource
			}
			s.EachSections(func(r, i int, s *Section) {
				if s.Resource == nil {
					s.Resource = this.Resource
				}
			})
			sections = append(sections, s.UniqueMetas(hasColumns))
		} else if column, ok := value.(string); ok {
			if strings.HasPrefix(column, "-") {
				excludedColumns = append(excludedColumns, column)
			} else if !isContainsColumn(excludedColumns, column) {
				sections = append(sections, &Section{Rows: [][]interface{}{{column}}})
			}
			hasColumns[column] = true
		} else if row, ok := value.([]string); ok {
			for j := len(row) - 1; j >= 0; j-- {
				column = row[j]
				sections = append(sections, &Section{Rows: [][]interface{}{{column}}})
				hasColumns[column] = true
			}
		} else {
			panic(fmt.Errorf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}

	sections = reverseSections(sections)
	for _, section := range sections {
		if section.Resource == nil {
			section.Resource = this.Resource
		}
	}
	return sections
}

func (this *SectionsAttribute) SetSectionsTo(provider *SchemeSectionsProvider, values []interface{}, cb ...func(sections Sections) Sections) Sections {
	sections := provider.MustSectionsCopy()

	var (
		replaces [][]string
		removes  []string
	)

	if len(values) == 0 {
		if len(sections) == 0 && this.AllSectionsFunc != nil {
			sections = this.AllSectionsFunc()
		}
	} else {
		var flattenValues []interface{}

		for _, value := range values {
			switch t := value.(type) {
			case string:
				// old replaces (deprecated): "~a:b" (replaces a to b)
				if t[0] == '~' {
					replaces = append(replaces, strings.Split(t[1:], ":"))
					continue
				} // old replaces (deprecated): "~a:b" (replaces a to b)
				if t[0] == '-' {
					removes = append(removes, t[1:])
					continue
				}
				// new replaces: 'a->b' (replaces a to b)
				if strings.Contains(t, "->") {
					replaces = append(replaces, strings.Split(t, "->"))
					continue
				}
				flattenValues = append(flattenValues, t)
			case []string:
				if len(values) == 1 {
					for _, column := range t {
						flattenValues = append(flattenValues, column)
					}
				} else {
					var colsi = make([]interface{}, len(t))
					for i, c := range t {
						colsi[i] = c
					}
					flattenValues = append(flattenValues, &Section{Resource: this.Resource, Rows: [][]interface{}{colsi}})
				}
			case [][]string:
				var colsi = make([][]interface{}, len(t))
				for i, cols := range t {
					ci := make([]interface{}, len(cols))
					for i, c := range cols {
						ci[i] = c
					}
					colsi[i] = ci
				}
				flattenValues = append(flattenValues, &Section{Resource: this.Resource, Rows: colsi})
			case []*Section:
				for _, section := range t {
					flattenValues = append(flattenValues, section)
				}
			case Sections:
				for _, section := range t {
					flattenValues = append(flattenValues, section)
				}
			case *Section:
				flattenValues = append(flattenValues, t)
			default:
				panic(fmt.Errorf("Resource: attributes should be Section or String, but it is %+v", value))
			}
		}

		if containsPositiveValue(flattenValues...) {
			sections = this.generateSections(flattenValues...)
		} else {
			var columns, availbleColumns []string

			for _, value := range flattenValues {
				if column, ok := value.(string); ok {
					columns = append(columns, column)
				}
			}

			for _, column := range this.AllSectionsFunc().MetasNames() {
				if !isContainsColumn(columns, column) {
					availbleColumns = append(availbleColumns, column)
				}
			}

			sections = this.generateSections(availbleColumns)
		}

		if len(replaces) > 0 {
			sections.ReplaceMetas(replaces)
		}
	}

	sections = sections.Uniquefy()
	if len(removes) > 0 {
		sections = sections.Filter(new(SectionsFilter).SetExcludes(removes...))
	}
	provider.sections = sections
	for _, cb := range cb {
		sections = cb(sections)
	}
	return sections
}

func (this *SectionsAttribute) IndexSections(context *Context) Sections {
	return this.GetSectionsProvider(context, INDEX).MustContextSectionsCopy(&SectionsContext{Ctx: context}).
		Allowed(nil, context, roles.Read)
}

func (this *SectionsAttribute) EditSections(context *Context, record interface{}) (sections Sections) {
	sections = this.GetSectionsProvider(context, EDIT).MustContextSectionsCopy(&SectionsContext{
		Ctx:    context,
		Record: record,
	})
	return sections.Allowed(record, context, roles.Update)
}

func (this *SectionsAttribute) NewSections(context *Context) Sections {
	return this.GetSectionsProvider(context, NEW).MustContextSectionsCopy(&SectionsContext{
		Ctx:    context,
		Record: context.ResourceRecord,
	}).Allowed(context.ResourceRecord, context, roles.Create)
}

func (this *SectionsAttribute) ShowSections(context *Context, record interface{}) []*Section {
	return this.ShowSectionsOriginal(context, record)
}

func (this *SectionsAttribute) ShowSectionsOriginal(context *Context, record interface{}) Sections {
	return this.GetSectionsProvider(context, SHOW).MustContextSectionsCopy(&SectionsContext{
		Ctx:    context,
		Record: record,
	}).Allowed(context.ResourceRecord, context, roles.Read)
}

func (this *SectionsAttribute) ContextSections(context *Context, recorde interface{}, action ...string) Sections {
	var b ContextType
	if len(action) > 0 && action[0] != "" {
		b = ParseContextType(action[0])
	} else {
		b = context.Type
	}

	if b.Has(NEW) {
		return this.NewSections(context)
	}
	if b.Has(SHOW) {
		return this.ShowSections(context, recorde)
	}
	if b.Has(EDIT) {
		return this.EditSections(context, recorde)
	}
	if b.Has(INDEX) {
		return this.IndexSections(context)
	}
	return nil
}

func (this *SectionsAttribute) CustomAttrsOf(prov *CRUDSchemeSectionsLayout, name string, values ...interface{}) Sections {
	customs := prov.Custom
	provider, ok := customs[name]
	if !ok {
		provider = &SchemeSectionsProvider{Name: name}
		customs[name] = provider
	}
	return this.SetSectionsTo(provider, values)
}

func (this *SectionsAttribute) IndexAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) Sections {
	return this.SetSectionsTo(prov.Index, values)
}

func (this *SectionsAttribute) NewAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) Sections {
	return this.SetSectionsTo(prov.New, values, this.ExcludeReadOnlyAttrs)
}

func (this *SectionsAttribute) EditAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) Sections {
	return this.SetSectionsTo(prov.Edit, values, this.ExcludeReadOnlyAttrs)
}

func (this *SectionsAttribute) ShowAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) Sections {
	return this.SetSectionsTo(prov.Show, values)
}

func (this *SectionsAttribute) NESAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) {
	this.NewAttrsOf(prov, values...)
	this.EditAttrsOf(prov, values...)
	this.ShowAttrsOf(prov, values...)
}

func (this *SectionsAttribute) INESAttrsOf(prov *CRUDSchemeSectionsLayout, values ...interface{}) {
	this.IndexAttrsOf(prov, values...)
	this.NewAttrsOf(prov, values...)
	this.EditAttrsOf(prov, values...)
	this.ShowAttrsOf(prov, values...)
}

func (this *SectionsAttribute) SectionsList(values ...interface{}) (dest Sections) {
	return this.SetSectionsTo(new(SchemeSectionsProvider), values)
}

func (this *SectionsAttribute) ExcludeReadOnlyAttrs(sections Sections) Sections {
	var names []string
	for _, f := range this.Resource.ModelStruct.ReadOnlyFields {
		var tags = ParseMetaTags(f.Tag)
		if tags.Tags == nil || !tags.Visible() {
			names = append(names, f.Name)
		}
	}

	if len(names) == 0 {
		return sections
	}

	return sections.Exclude(names...)
}

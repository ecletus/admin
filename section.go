package admin

import (
	"bytes"
	"fmt"
	"strings"
)

// Section is used to structure forms, it could group your fields into sections, to make your form clean & tidy
//    product.EditAttrs(
//      &admin.Section{
//      	Title: "Basic Information",
//      	Rows: [][]string{
//      		{"Name"},
//      		{"Code", "Price"},
//      	}},
//      &admin.Section{
//      	Title: "Organization",
//      	Rows: [][]string{
//      		{"Category", "Collections", "MadeCountry"},
//      	}},
//      "Description",
//      "ColorVariations",
//    }
type Section struct {
	Resource     *Resource
	Title        string
	Help         string
	ReadOnlyHelp string
	Rows         [][]interface{}
}

func (this Section) Copy() *Section {
	return &this
}

func (this Section) CopyWoRows() *Section {
	this.Rows = nil
	return &this
}

func (this *Section) MetasNamesCb(cb func(name string)) {
	for _, row := range this.Rows {
		for _, col := range row {
			switch col := col.(type) {
			case string:
				cb(col)
			case *Section:
				col.MetasNamesCb(cb)
			}
		}
	}
}

func (this *Section) MetasNames() (names []string) {
	this.MetasNamesCb(func(name string) {
		names = append(names, name)
	})
	return
}

func (this *Section) Metas(cb func(meta *Meta) (cont bool), getter ...func(res *Resource, name string) *Meta) {
	var g func(res *Resource, name string) *Meta
	for _, g = range getter {
	}
	if g == nil {
		g = func(res *Resource, name string) *Meta {
			if res == nil {
				return nil
			}
			return res.GetMeta(name)
		}
	}
	for _, row := range this.Rows {
		for _, col := range row {
			switch col := col.(type) {
			case string:
				if m := g(this.Resource, col); m != nil {
					if !cb(m) {
						return
					}
				}
			case *Section:
				var cont bool
				col.Metas(func(meta *Meta) bool {
					cont = cb(meta)
					return cont
				}, getter...)
				if !cont {
					return
				}
			}
		}
	}
}

func (this *Section) Filter(f *SectionsFilter) *Section {
	var (
		rows    [][]interface{}
		changed bool
	)
	if f.Include != nil {
		for _, row := range this.Rows {
			var newRow []interface{}
			for _, col := range row {
				switch t := col.(type) {
				case string:
					if _, ok := f.Include[t]; ok {
						newRow = append(newRow, t)
					} else {
						changed = true
					}
				case *Section:
					if _, ok := f.Include["`"+t.Title]; !ok {
						if sec := t.Filter(f); sec != nil {
							newRow = append(newRow, t)
							if sec != t {
								changed = true
							}
						} else {
							changed = true
						}
					}
				}
			}
			if newRow != nil {
				rows = append(rows, newRow)
			}
		}
	} else if f.Excludes != nil {
		for _, row := range this.Rows {
			var newRow []interface{}
			for _, col := range row {
				switch t := col.(type) {
				case string:
					if f.Excludes(t) {
						changed = true
					} else {
						newRow = append(newRow, t)
					}
				case *Section:
					if f.Excludes("`" + t.Title) {
						changed = true
					} else {
						if sec := t.Filter(f); sec != nil {
							newRow = append(newRow, t)
							if sec != t {
								changed = true
							}
						} else {
							changed = true
						}
					}
				}
			}
			if newRow != nil {
				rows = append(rows, newRow)
			}
		}
	}

	if !changed {
		return this
	}

	if rows == nil {
		return nil
	}

	s2 := *this
	s2.Rows = rows
	return &s2
}

type SectionsFilter struct {
	Include,
	Exclude map[string]interface{}
	Excludes func(name string) bool
}

func (this *SectionsFilter) Unique() *SectionsFilter {
	if this.Exclude == nil {
		this.Exclude = map[string]interface{}{}
	}
	this.Excludes = func(name string) (ok bool) {
		if _, ok = this.Exclude[name]; !ok {
			this.Exclude[name] = nil
		}
		return
	}

	return this
}

func (this *SectionsFilter) SetInclude(names ...string) *SectionsFilter {
	if this.Exclude != nil {
		for _, name := range names {
			delete(this.Exclude, name)
		}
	} else {
		this.Include = map[string]interface{}{}
		for _, name := range names {
			this.Include[name] = nil
		}
	}

	return this
}

func (this *SectionsFilter) SetExcludes(names ...string) *SectionsFilter {
	if this.Include != nil {
		for _, name := range names {
			delete(this.Exclude, name)
		}
	} else {
		this.Exclude = map[string]interface{}{}
		for _, name := range names {
			this.Exclude[name] = nil
		}
		this.Excludes = func(name string) (ok bool) {
			_, ok = this.Exclude[name]
			return
		}
	}
	return this
}

func (this *Section) EachColumns(cb func(r, i int, col string)) {
	for r, row := range this.Rows {
		for i, col := range row {
			switch col := col.(type) {
			case string:
				cb(r, i, col)
			case *Section:
				col.EachColumns(cb)
			}
		}
	}
}

func (this *Section) Each(cb func(r, i int, col interface{})) {
	for r, row := range this.Rows {
		for i, col := range row {
			switch col := col.(type) {
			case string:
				cb(r, i, col)
			case *Section:
				cb(r, i, col)
				col.Each(cb)
			}
		}
	}
}

func (this *Section) Walk(parentDot interface{}, at func(pDot interface{}, this *Section) interface{}, cb func(dot interface{}, r, i int, col interface{})) {
	var dot = parentDot
	if this.Title != "" {
		dot = at(parentDot, this)
	}
	for r, row := range this.Rows {
		for i, col := range row {
			switch col := col.(type) {
			case string:
				cb(dot, r, i, col)
			case *Section:
				if col.Title != "" {
					cb(dot, r, i, col)
				}
				col.Walk(dot, at, cb)
			}
		}
	}
}

func (this *Section) EachSections(cb func(r, i int, s *Section)) {
	for r, row := range this.Rows {
		for i, col := range row {
			switch s := col.(type) {
			case *Section:
				cb(r, i, s)
				s.EachSections(cb)
			}
		}
	}
}

func (this *Section) Uniquefy() *Section {
	return this.Filter(new(SectionsFilter).Unique())
}

// String stringify section
func (this *Section) String() string {
	var buf bytes.Buffer
	if this.Title != "" {
		fmt.Fprintf(&buf, "%q:", this.Title)
	}
	buf.WriteString("{")

	rl := len(this.Rows) - 1
	for i, r := range this.Rows {
		buf.WriteString("{")
		l := len(r) - 1
		for i, col := range r {
			switch t := col.(type) {
			case string:
				fmt.Fprintf(&buf, "%q", col)
			case *Section:
				fmt.Fprintf(&buf, "%s", t)
			}
			if i < l {
				buf.WriteString(";")
			}
		}
		buf.WriteString("}")
		if i < rl {
			buf.WriteString(";")
		}
	}
	buf.WriteString("}")
	return buf.String()
}

func (this *Section) AddPrefix(prefix string) *Section {
	s := &Section{Resource: this.Resource, Title: this.Title, Rows: make([][]interface{}, len(this.Rows))}
	for i, columns := range this.Rows {
		s.Rows[i] = make([]interface{}, len(columns))
		for j, v := range columns {
			switch t := v.(type) {
			case string:
				s.Rows[i][j] = prefix + "." + t
			case *Section:
				s.Rows[i][j] = t.AddPrefix(prefix)
			}
		}
	}
	return s
}

func (this *Section) ReplaceMetas(pairs [][]string) *Section {
	for _, p := range pairs {
		this.ReplaceMeta(p[0], p[1])
	}
	return this
}

func (this *Section) ReplaceMeta(from, to string) *Section {
	for i, columns := range this.Rows {
		for j, col := range columns {
			switch t := col.(type) {
			case string:
				if t == from {
					this.Rows[i][j] = to
				}
			case *Section:
				t.ReplaceMeta(from, to)
			}
		}
	}
	return this
}

func (this *Section) ReplaceSection(from *Section, to *Section) {
	for i, columns := range this.Rows {
		for j, col := range columns {
			switch t := col.(type) {
			case *Section:
				if t == from {
					this.Rows[i][j] = to
				}
			}
		}
	}
}

func (this Section) UniqueMetas(has map[string]interface{}) *Section {
	var newRows [][]interface{}
	for _, row := range this.Rows {
		var newColumns []interface{}
		for _, column := range row {
			switch t := column.(type) {
			case string:
				if _, ok := has[t]; !ok {
					has[t] = true
					newColumns = append(newColumns, t)
				}
			case *Section:
				if t.Resource != this.Resource {
					// skip different resource
					newColumns = append(newColumns, t)
				} else {
					newColumns = append(newColumns, t.UniqueMetas(has))
				}
			}

		}
		if len(newColumns) > 0 {
			newRows = append(newRows, newColumns)
		}
	}
	this.Rows = newRows
	return &this
}

func reverseSections(sections []*Section) []*Section {
	var results []*Section
	for i := 0; i < len(sections); i++ {
		results = append(results, sections[len(sections)-i-1])
	}
	return results
}

func isContainsColumn(hasColumns []string, column string) bool {
	for _, col := range hasColumns {
		if strings.TrimLeft(col, "-") == strings.TrimLeft(column, "-") {
			return true
		}
	}
	return false
}

func containsPositiveValue(values ...interface{}) bool {
	for _, value := range values {
		if _, ok := value.(*Section); ok {
			return true
		} else if column, ok := value.(string); ok {
			if !strings.HasPrefix(column, "-") {
				return true
			}
		} else {
			panic(fmt.Errorf("Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	return false
}

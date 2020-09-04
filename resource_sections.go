package admin

import (
	"fmt"
	"strings"

	"github.com/ecletus/roles"
)

func (this *Resource) SectionsList(values ...interface{}) (dest []*Section) {
	this.setSections(&dest, values...)
	return
}

func (this *Resource) allowedSections(record interface{}, sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	var newSections []*Section
	for _, section := range sections {
		newSection := &Section{Resource: section.Resource, Title: section.Title}
		for _, row := range section.Rows {
			func() {
				var columns []string
				for _, column := range row {
					meta := section.Resource.GetMeta(column)
					if meta != nil {
						if meta.SectionNotAllowed {
							continue
						}

						if meta.Enabled != nil && !meta.Enabled(record, context, meta) {
							continue
						}

						if context.HasAnyPermission(meta, roles...) {
							if len(columns) == 0 {
								if sec := meta.Tags.Section(); sec != nil {
									var add = func() {
										// add current
										newSections = append(newSections, &Section{
											Resource:     section.Resource,
											Title:        sec.Title,
											Help:         sec.Help,
											ReadOnlyHelp: sec.ReadOnlyHelp,
											Rows:         [][]string{{column}},
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
										newSection = &Section{Resource: section.Resource, Title: section.Title}
									}
									continue
								}
							}
							columns = append(columns, column)
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

func (this *Resource) generateSections(values ...interface{}) []*Section {
	var sections []*Section
	var hasColumns, excludedColumns []string

	// Reverse values to make the last one as a key one
	// e.g. Name, Code, -Name (`-Name` will get first and will skip `Name`)
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]
		if section, ok := value.(*Section); ok {
			sections = append(sections, uniqueSection(section, &hasColumns))
		} else if column, ok := value.(string); ok {
			if strings.HasPrefix(column, "-") {
				excludedColumns = append(excludedColumns, column)
			} else if !isContainsColumn(excludedColumns, column) {
				sections = append(sections, &Section{Rows: [][]string{{column}}})
			}
			hasColumns = append(hasColumns, column)
		} else if row, ok := value.([]string); ok {
			for j := len(row) - 1; j >= 0; j-- {
				column = row[j]
				sections = append(sections, &Section{Rows: [][]string{{column}}})
				hasColumns = append(hasColumns, column)
			}
		} else {
			panic(fmt.Errorf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}

	sections = reverseSections(sections)
	for _, section := range sections {
		if section.Resource == nil {
			section.Resource = this
		}
	}
	return sections
}

// ConvertSectionToMetas convert section to metas
func (this *Resource) ConvertSectionToMetas(sections []*Section, metaGetter ...func(name string) *Meta) []*Meta {
	var mg func(name string) *Meta
	for _, mg = range metaGetter {
	}
	if mg == nil {
		mg = func(name string) *Meta {
			return this.GetMeta(name)
		}
	}
	var metas []*Meta
	for _, section := range sections {
		for _, row := range section.Rows {
			for _, col := range row {
				meta := mg(col)
				if meta != nil && meta.Type != "-" {
					metas = append(metas, meta)
				}
			}
		}
	}
	return metas
}

// ConvertSectionToStrings convert section to strings
func (this *Resource) ConvertSectionToStrings(sections []*Section) []string {
	var columns []string
	for _, section := range sections {
		for _, row := range section.Rows {
			for _, col := range row {
				columns = append(columns, col)
			}
		}
	}
	return columns
}

func (this *Resource) setSections(sections *[]*Section, values ...interface{}) {
	var replaces [][]string

	if len(values) == 0 {
		if len(*sections) == 0 {
			*sections = this.generateSections(this.allAttrs())
		}
	} else {
		var flattenValues []interface{}

		for _, value := range values {
			if columns, ok := value.([]string); ok {
				for _, column := range columns {
					flattenValues = append(flattenValues, column)
				}
			} else if _sections, ok := value.([]*Section); ok {
				for _, section := range _sections {
					flattenValues = append(flattenValues, section)
				}
			} else if section, ok := value.(*Section); ok {
				flattenValues = append(flattenValues, section)
			} else if column, ok := value.(string); ok {
				// old replaces: "~a:b" (replaces a to b)
				if column[0] == '~' {
					replaces = append(replaces, strings.Split(column[1:], ":"))
					continue
				}
				// new replaces: 'a->b' (replaces a to b)
				if strings.Contains(column, "->") {
					replaces = append(replaces, strings.Split(column, "->"))
					continue
				}
				flattenValues = append(flattenValues, column)
			} else if columns, ok := value.([]string); ok {
				flattenValues = append(flattenValues, &Section{Resource: this, Rows: [][]string{columns}})
			} else if column, ok := value.([][]string); ok {
				flattenValues = append(flattenValues, &Section{Resource: this, Rows: column})
			} else {
				panic(fmt.Errorf("Resource: attributes should be Section or String, but it is %+v", value))
			}
		}

		if containsPositiveValue(flattenValues...) {
			*sections = this.generateSections(flattenValues...)
		} else {
			var columns, availbleColumns []string
			for _, value := range flattenValues {
				if column, ok := value.(string); ok {
					columns = append(columns, column)
				}
			}

			for _, column := range this.allAttrs() {
				if !isContainsColumn(columns, column) {
					availbleColumns = append(availbleColumns, column)
				}
			}
			*sections = this.generateSections(availbleColumns)
		}

	replLoop:
		for _, rpl := range replaces {
			from, to := rpl[0], rpl[1]
			for _, sec := range *sections {
				for _, row := range sec.Rows {
					for i, col := range row {
						if col == from {
							row[i] = to
							continue replLoop
						}
					}
				}
			}
		}
	}
}

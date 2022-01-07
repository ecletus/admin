package admin

import (
	"fmt"
	"strings"
)

func (this *Resource) generateSections(values ...interface{}) (sections Sections) {
	var (
		hasColumns      = map[string]interface{}{}
		excludedColumns []string
	)

	// Reverse values to make the last one as a key one
	// e.g. Name, Code, -Name (`-Name` will get first and will skip `Name`)
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]
		if section, ok := value.(*Section); ok {
			if section.Resource == nil {
				section.Resource = this
			}
			sections = append(sections, section.UniqueMetas(hasColumns))
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
			section.Resource = this
		}
	}
	return sections
}

// ConvertSectionToMetasTable convert section to metas table
func (this *Resource) ConvertSectionToMetasTable(sections Sections, metaGetter ...func(name string) *Meta) *MetasTable {
	var mg func(name string) *Meta

	for _, mg = range metaGetter {
	}

	if mg == nil {
		mg = func(name string) *Meta {
			return this.GetMeta(name)
		}
	}
	return ConvertSectionsToMetasTable(sections, mg)
}

func (this *Resource) AllSections() (sections Sections) {
	if sections = this.AllSectionsProvider.sections; sections == nil {
		sections = this.generateSections(this.allAttrs())
		this.AllSectionsProvider.sections = sections
	}
	return
}

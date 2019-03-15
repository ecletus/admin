package admin

import (
	"fmt"
	"strings"

	"github.com/aghape/core"
	"github.com/aghape/core/utils"
	"github.com/aghape/roles"
)

func (res *Resource) SectionsList(values ...interface{}) (dest []*Section) {
	res.setSections(&dest, values...)
	return
}

func (res *Resource) allowedSections(record interface{}, sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	var newSections []*Section
	for _, section := range sections {
		newSection := &Section{Resource: section.Resource, Title: section.Title}
		var editableRows [][]string
		for _, row := range section.Rows {
			var editableColumns []string
			for _, column := range row {
				meta := section.Resource.GetMeta(column)
				if meta != nil {
					if meta.Enabled != nil && !meta.Enabled(record, context, meta) {
						continue
					}

					for _, role := range roles {
						if core.HasPermission(meta, role, context.Context) {
							editableColumns = append(editableColumns, column)
							break
						}
					}
				}
			}
			if len(editableColumns) > 0 {
				editableRows = append(editableRows, editableColumns)
			}
		}

		if len(editableRows) > 0 {
			newSection.Rows = editableRows
			newSections = append(newSections, newSection)
		}
	}
	return newSections
}

func (res *Resource) generateSections(values ...interface{}) []*Section {
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
			utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}

	sections = reverseSections(sections)
	for _, section := range sections {
		if section.Resource == nil {
			section.Resource = res
		}
	}
	return sections
}

// ConvertSectionToMetas convert section to metas
func (res *Resource) ConvertSectionToMetas(sections []*Section) []*Meta {
	var metas []*Meta
	for _, section := range sections {
		for _, row := range section.Rows {
			for _, col := range row {
				meta := res.GetMeta(col)
				if meta != nil && meta.Type != "-" {
					metas = append(metas, meta)
				}
			}
		}
	}
	return metas
}

// ConvertSectionToStrings convert section to strings
func (res *Resource) ConvertSectionToStrings(sections []*Section) []string {
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

func (res *Resource) setSections(sections *[]*Section, values ...interface{}) {
	if len(values) == 0 {
		if len(*sections) == 0 {
			*sections = res.generateSections(res.allAttrs())
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
				flattenValues = append(flattenValues, column)
			} else if columns, ok := value.([]string); ok {
				flattenValues = append(flattenValues, &Section{Resource: res, Rows: [][]string{columns}})
			} else if column, ok := value.([][]string); ok {
				flattenValues = append(flattenValues, &Section{Resource: res, Rows: column})
			} else {
				utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
			}
		}

		if containsPositiveValue(flattenValues...) {
			*sections = res.generateSections(flattenValues...)
		} else {
			var columns, availbleColumns []string
			for _, value := range flattenValues {
				if column, ok := value.(string); ok {
					columns = append(columns, column)
				}
			}

			for _, column := range res.allAttrs() {
				if !isContainsColumn(columns, column) {
					availbleColumns = append(availbleColumns, column)
				}
			}
			*sections = res.generateSections(availbleColumns)
		}
	}
}

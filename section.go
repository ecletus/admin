package admin

import (
	"fmt"
	"strings"

	"github.com/ecletus/core/utils"
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
	Resource *Resource
	Title    string
	Rows     [][]string
}

type Sections []*Section

func (s Sections) AddPrefix(prefix string) []*Section {
	items := make([]*Section, len(s))
	for i, section := range s {
		items[i] = section.AddPrefix(prefix)
	}
	return items
}

// String stringify section
func (section *Section) String() string {
	return fmt.Sprint(section.Rows)
}

func (section *Section) AddPrefix(prefix string) *Section {
	s := &Section{section.Resource, section.Title, make([][]string, len(section.Rows))}
	for i, columns := range section.Rows {
		s.Rows[i] = make([]string, len(columns))
		for j, v := range columns {
			s.Rows[i][j] = prefix + "." + v
		}
	}
	return s
}

func uniqueSection(section *Section, hasColumns *[]string) *Section {
	newSection := Section{Title: section.Title}
	var newRows [][]string
	for _, row := range section.Rows {
		var newColumns []string
		for _, column := range row {
			if !isContainsColumn(*hasColumns, column) {
				newColumns = append(newColumns, column)
				*hasColumns = append(*hasColumns, column)
			}
		}
		if len(newColumns) > 0 {
			newRows = append(newRows, newColumns)
		}
	}
	newSection.Rows = newRows
	return &newSection
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
			utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	return false
}

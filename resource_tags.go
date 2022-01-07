package admin

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/moisespsena/template/text/template"

	"github.com/moisespsena-go/aorm"
	tag_scanner "github.com/unapu-go/tag-scanner"
)

type ResourceTags struct {
	Tags
	stringifyTag *StringifyTag
}

func (this ResourceTags) ShowPage() bool {
	return this.Flag("SHOW_PAGE")
}

func (this ResourceTags) Attrs() []*Section {
	if sections, _ := this.GetOk("ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) AttrsInclude() []*Section {
	if sections, _ := this.GetOk("ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) AttrsIncludeBeginning() []*Section {
	if sections, _ := this.GetOk("ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) AttrsExclude() []string {
	if sections, _ := this.GetOk("ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}

func (this ResourceTags) NewAttrs() []*Section {
	if sections, _ := this.GetOk("NEW_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) NewAttrsExclude() []string {
	if sections, _ := this.GetOk("NEW_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) NewAttrsInclude() []*Section {
	if sections, _ := this.GetOk("NEW_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) NewAttrsIncludeBeginning() []*Section {
	if sections, _ := this.GetOk("NEW_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) EditAttrs() []*Section {
	if sections, _ := this.GetOk("EDIT_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) EditAttrsExclude() []string {
	if sections, _ := this.GetOk("EDIT_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) EditAttrsInclude() []*Section {
	if sections, _ := this.GetOk("EDIT_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) EditAttrsIncludeBeginning() []*Section {
	if sections, _ := this.GetOk("EDIT_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) ShowAttrs() []*Section {
	if sections, _ := this.GetOk("SHOW_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) ShowAttrsExclude() []string {
	if sections, _ := this.GetOk("SHOW_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) ShowAttrsInclude() []*Section {
	if sections, _ := this.GetOk("SHOW_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) ShowAttrsIncludeBeginning() []*Section {
	if sections, _ := this.GetOk("SHOW_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) IndexAttrs() []*Section {
	if sections, _ := this.GetOk("INDEX_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) IndexAttrsAliasTo() string {
	if val, _ := this.GetOk("INDEX_ATTRS"); val != "" && val[0] == '@' {
		return val[1:]
	}
	return ""
}
func (this ResourceTags) IndexAttrsExclude() []string {
	if sections, _ := this.GetOk("INDEX_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) ReadOnlyAttrs() []string {
	if tags := this.GetTags("RO_ATTRS", tag_scanner.FlagPreserveKeys|tag_scanner.FlagForceTags); tags != nil {
		return tags.Flags()
	}
	return nil
}
func (this ResourceTags) SortAttrs() []string {
	if tags := this.GetTags("SORT_ATTRS", tag_scanner.FlagTags|tag_scanner.FlagPreserveKeys|tag_scanner.FlagForceTags); tags != nil {
		return tags.Flags()
	}
	return nil
}

func (this ResourceTags) SetAttrsTo(sa *SectionsAttribute, layout *CRUDSchemeSectionsLayout) {
	if sections := this.Attrs(); len(sections) > 0 {
		sa.NESAttrsOf(layout, sections)
	}
	if sections := this.AttrsInclude(); len(sections) > 0 {
		sa.NESAttrsOf(layout, sa.NewAttrsOf(layout), sections)
	}
	if sections := this.AttrsIncludeBeginning(); len(sections) > 0 {
		sa.NESAttrsOf(layout, sections, sa.NewAttrsOf(layout))
	}
	if names := this.AttrsExclude(); len(names) > 0 {
		sa.NESAttrsOf(layout, sa.NewAttrsOf(layout).Exclude(names...))
	}
	if sections := this.ShowAttrs(); len(sections) > 0 {
		sa.ShowAttrsOf(layout, sections)
	}
	if names := this.ShowAttrsExclude(); len(names) > 0 {
		sa.ShowAttrsOf(layout, sa.ShowAttrsOf(layout).Exclude(names...))
	}
	if sections := this.ShowAttrsInclude(); len(sections) > 0 {
		sa.ShowAttrsOf(layout, sa.ShowAttrsOf(layout), sections)
	}
	if sections := this.ShowAttrsIncludeBeginning(); len(sections) > 0 {
		sa.ShowAttrsOf(layout, sections, sa.ShowAttrsOf(layout))
	}
	if sections := this.NewAttrs(); len(sections) > 0 {
		sa.NewAttrsOf(layout, sections)
	}
	if names := this.NewAttrsExclude(); len(names) > 0 {
		sa.NewAttrsOf(layout, sa.NewAttrsOf(layout).Exclude(names...))
	}
	if sections := this.NewAttrsInclude(); len(sections) > 0 {
		sa.NewAttrsOf(layout, sa.NewAttrsOf(layout), sections)
	}
	if sections := this.NewAttrsIncludeBeginning(); len(sections) > 0 {
		sa.NewAttrsOf(layout, sections, sa.NewAttrsOf(layout))
	}
	if sections := this.EditAttrs(); len(sections) > 0 {
		sa.EditAttrsOf(layout, sections)
	}
	if names := this.EditAttrsExclude(); len(names) > 0 {
		sa.EditAttrsOf(layout, sa.EditAttrsOf(layout).Exclude(names...))
	}
	if sections := this.EditAttrsInclude(); len(sections) > 0 {
		sa.EditAttrsOf(layout, sa.EditAttrsOf(layout), sections)
	}
	if sections := this.EditAttrsIncludeBeginning(); len(sections) > 0 {
		sa.EditAttrsOf(layout, sections, sa.EditAttrsOf(layout))
	}
	if alias := this.IndexAttrsAliasTo(); alias != "" {
		layout.Index.Parent = layout.Get(alias)

		if names := this.IndexAttrsExclude(); len(names) > 0 {
			layout.Index.sections = layout.Index.MustSections().Exclude(names...)
		}
	} else {
		if sections := this.IndexAttrs(); len(sections) > 0 {
			sa.IndexAttrsOf(layout, sections)
		}
		if names := this.IndexAttrsExclude(); len(names) > 0 {
			sa.IndexAttrsOf(layout, sa.IndexAttrsOf(layout).Exclude(names...))
		}
	}
}

func (this ResourceTags) PkOrder() aorm.Order {
	if this.Flag("DESC") {
		return aorm.DESC
	} else if this.Flag("ASC") {
		return aorm.ASC
	}
	switch strings.ToUpper(this.Get("PK_ORDER")) {
	case "ASC":
		return aorm.ASC
	case "DESC":
		return aorm.DESC
	default:
		return 0
	}
}

// Stringify field or expression
// Acceptable values: `FieldName` or `#MethodName` or `{go template code}`.
// Example of go template code: `{User Name: {{.Name}}}
func (this *ResourceTags) Stringify() (tag *StringifyTag, err error) {
	for _, key := range []string{"STRINGIFY", "STR"} {
		value := this.GetString(key)
		if value == "" {
			continue
		}
		if this.stringifyTag == nil {
			this.stringifyTag = &StringifyTag{}
			switch value[0] {
			case '{':
				this.stringifyTag.Template, err = template.New("").Parse(value[1 : len(value)-1])
				if err != nil {
					err = errors.Wrapf(err, "compile stringify template %q", value)
					return nil, err
				}
			case '#':
				this.stringifyTag.MethodName = value[1:]
			default:
				this.stringifyTag.FieldName = value
			}
		}
		return this.stringifyTag, nil
	}
	return
}

// Show the Stringify type
func (this ResourceTags) Show() string {
	return this.Tags["SHOW"]
}

// Show the Stringify type
func (this ResourceTags) Resource() string {
	return this.Tags["RES"]
}

func (this ResourceTags) Search() []string {
	if names, _ := this.GetOk("SEARCH"); names != "" {
		if this.Scanner().IsTags(names) {
			return tag_scanner.Flags(this.Scanner(), this.Scanner().String(names))
		}
		return []string{names}
	}
	return nil
}

func (this ResourceTags) Order() []string {
	if names, _ := this.GetOk("ORDER"); names != "" {
		if this.Scanner().IsTags(names) {
			return tag_scanner.Flags(this.Scanner(), this.Scanner().String(names))
		}
		return []string{names}
	}
	return nil
}

func (this ResourceTags) ParentPreload() []string {
	if tags := this.GetTags("PARENT_PRELOAD"); tags != nil {
		return tags.Flags()
	}
	return nil
}

func ParseResourceTags(tag reflect.StructTag) (tags ResourceTags) {
	tags.ParseDefault(aorm.StructTag(tag), "admin")
	return
}

type StringifyTag struct {
	FieldName, MethodName string
	Template              *template.Template
}

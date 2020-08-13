package admin

import (
	"reflect"
	"strings"

	"github.com/moisespsena/template/text/template"
	"github.com/pkg/errors"

	"github.com/moisespsena-go/aorm"
	tag_scanner "github.com/unapu-go/tag-scanner"
)

type MetaTags struct {
	aorm.TagSetting
	typeOptions aorm.TagSetting
}

func (this MetaTags) DefaultInvisible() bool {
	return this.Flag("DEFAULT_INVISIBLE") || this.Flag("--")
}
func (this MetaTags) Hidden() bool {
	return this.Flag("-")
}
func (this MetaTags) Readonly() bool {
	return this.Flag("RO")
}
func (this MetaTags) Managed() bool {
	return this.Flag("MANAGED")
}
func (this MetaTags) Required() bool {
	return this.Flag("REQUIRED")
}
func (this MetaTags) Help() string {
	return this.GetString("HELP")
}
func (this MetaTags) ShowHelp() (s string) {
	if s := this.GetString("SHOW_HELP"); s == "" {
		// read only
		s = this.GetString("RO_HELP")
	}
	return
}
func (this MetaTags) Label() string {
	return this.GetString("LABEL")
}
func (this MetaTags) Type() string {
	return this.TagSetting["TYPE"]
}
func (this MetaTags) TypeOptions() (opt aorm.TagSetting) {
	if this.typeOptions == nil {
		if v, ok := this.TagSetting["TYPE_OPT"]; ok {
			opt.ParseString(v)
			this.typeOptions = opt
		}
	}
	return this.typeOptions
}
func (this MetaTags) Section() (sec *struct{ Title, Help, ReadOnlyHelp string }) {
	if v, ok := this.TagSetting["SECTION"]; ok {
		var opt aorm.TagSetting
		opt.ParseString(v)
		if len(opt) > 0 {
			sec = &struct{ Title, Help, ReadOnlyHelp string }{
				opt["TITLE"],
				tag_scanner.Default.String(opt["HELP"]),
				tag_scanner.Default.String(opt["RO_HELP"]),
			}
		}
	}
	return
}
func (this MetaTags) NilAsZero() bool {
	return this.TagSetting.Flag("NILZ")
}
func (this MetaTags) Sort() bool {
	return this.TagSetting.Flag("SORT")
}
func (this MetaTags) Search() bool {
	return this.TagSetting.Flag("SEARCH")
}
func (this MetaTags) Filter() bool {
	return this.TagSetting.Flag("FILTER")
}
func (this MetaTags) SelectOne() (cfg *SelectOneConfig, resID string) {
	if value := this.TagSetting["SELECT_ONE"]; value != "" {
		cfg = &SelectOneConfig{}
		if value == "SELECT_ONE" {
			return
		}
		if tgs := this.Tags(value); len(tgs) == 1 && tgs["RES"] != "" {
			resID = tgs["RES"]
		} else {
			cfg.Collection = tag_scanner.KeyValuePairs(value)
		}
	}
	return
}
func (this MetaTags) SelectMany() (cfg *SelectManyConfig, resID string) {
	if value := this.TagSetting["SELECT_MANY"]; value != "" {
		cfg = &SelectManyConfig{}
		if value == "SELECT_MANY" {
			return
		}
		if tgs := this.Tags(value); len(tgs) == 1 && tgs["RES"] != "" {
			resID = tgs["RES"]
		} else {
			cfg.Collection = tag_scanner.KeyValuePairs(value)
		}
	}
	return
}

func ParseMetaTags(tag reflect.StructTag) (tags MetaTags) {
	tags.ParseCallback(aorm.StructTag(tag), []string{"admin"}, func(dest map[string]string, n tag_scanner.Node) {
		if n.Type() == tag_scanner.KeyValue {
			kv := n.(tag_scanner.NodeKeyValue)
			switch kv.Key {
			case "type":
				// "type:number:zero:-"
				if len(kv.KeyArgs) == 2 {
					dest["TYPE_OPT"] = kv.KeyArgs[1] + ":" + kv.Value
					kv.Value = kv.KeyArgs[0]
				}
			}
		}
	})
	return
}

type ResourceTags struct {
	aorm.TagSetting
	stringifyTag *StringifyTag
}

func (this ResourceTags) ShowPage() bool {
	return this.Flag("SHOW_PAGE")
}

func (this ResourceTags) Attrs() []*Section {
	if sections, _ := this.Get("ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) AttrsInclude() []*Section {
	if sections, _ := this.Get("ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) AttrsIncludeBeginning() []*Section {
	if sections, _ := this.Get("ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) AttrsExclude() []string {
	if sections, _ := this.Get("ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}

func (this ResourceTags) NewAttrs() []*Section {
	if sections, _ := this.Get("NEW_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) NewAttrsExclude() []string {
	if sections, _ := this.Get("NEW_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) NewAttrsInclude() []*Section {
	if sections, _ := this.Get("NEW_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) NewAttrsIncludeBeginning() []*Section {
	if sections, _ := this.Get("NEW_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) EditAttrs() []*Section {
	if sections, _ := this.Get("EDIT_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) EditAttrsExclude() []string {
	if sections, _ := this.Get("EDIT_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) EditAttrsInclude() []*Section {
	if sections, _ := this.Get("EDIT_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) EditAttrsIncludeBeginning() []*Section {
	if sections, _ := this.Get("EDIT_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) ShowAttrs() []*Section {
	if sections, _ := this.Get("SHOW_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) ShowAttrsExclude() []string {
	if sections, _ := this.Get("SHOW_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}
func (this ResourceTags) ShowAttrsInclude() []*Section {
	if sections, _ := this.Get("SHOW_ATTRS+"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) ShowAttrsIncludeBeginning() []*Section {
	if sections, _ := this.Get("SHOW_ATTRS^"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}

func (this ResourceTags) IndexAttrs() []*Section {
	if sections, _ := this.Get("INDEX_ATTRS"); sections != "" {
		return ParseSections(sections)
	}
	return nil
}
func (this ResourceTags) IndexAttrsExclude() []string {
	if sections, _ := this.Get("INDEX_ATTRS-"); sections != "" {
		return strings.Split(sections[1:len(sections)-1], ";")
	}
	return nil
}

// Stringify field or expression
// Acceptable values: `FieldName` or `#MethodName` or `{go template code}`.
// Example of go template code: `{User Name: {{.Name}}}
func (this *ResourceTags) Stringify() (tag *StringifyTag, err error) {
	for _, value := range []string{"STRINGIFY", "STR"} {
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
	return this.TagSetting["SHOW"]
}

// Show the Stringify type
func (this ResourceTags) Resource() string {
	return this.TagSetting["RES"]
}

func ParseResourceTags(tag reflect.StructTag) (tags ResourceTags) {
	tags.Parse(aorm.StructTag(tag), "admin")
	return
}

type StringifyTag struct {
	FieldName, MethodName string
	Template              *template.Template
}

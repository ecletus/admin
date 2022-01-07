package admin

import (
	"html/template"
	"reflect"

	tag_scanner "github.com/unapu-go/tag-scanner"

	"github.com/moisespsena-go/aorm"
)

type Tags = tag_scanner.Map

type MetaTags struct {
	Tags
	typeOptions aorm.TagSetting
}

func (this MetaTags) DefaultInvisible() bool {
	return this.Flag("DEFAULT_INVISIBLE") || this.Flag("--")
}
func (this MetaTags) Hidden() bool {
	return this.Flag("-")
}
func (this MetaTags) Visible() bool {
	return this.Flag("VISIBLE")
}
func (this MetaTags) Readonly() bool {
	return this.Flag("RO")
}
func (this MetaTags) ReadonlyStringer() bool {
	return this.Flag("RO_STR")
}
func (this MetaTags) Managed() (modes []string) {
	if this.Flag("MANAGED") {
		return []string{}
	} else if m := this.GetTags("MANAGED", tag_scanner.FlagForceTags); m != nil {
		return m.Flags()
	}
	return
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
	return this.Tags["TYPE"]
}
func (this MetaTags) TypeOptions() (opt aorm.TagSetting) {
	if this.typeOptions == nil {
		if v, ok := this.Tags["TYPE_OPT"]; ok {
			opt.ParseStringDefault(v)
			this.typeOptions = opt
		}
	}
	return this.typeOptions
}
func (this MetaTags) Section() (sec *struct{ Title, Help, ReadOnlyHelp string }) {
	if v, ok := this.Tags["SECTION"]; ok {
		var opt aorm.TagSetting
		opt.ParseStringDefault(v)
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
	return this.Tags.Flag("NILZ")
}
func (this MetaTags) Sort() bool {
	return this.Tags.Flag("SORT")
}
func (this MetaTags) Search() bool {
	return this.Tags.Flag("SEARCH")
}
func (this MetaTags) Filter() bool {
	return this.Tags.Flag("FILTER")
}
func (this MetaTags) LockedField() string {
	return this.Tags.Get("LOCKED_FIELD")
}
func (this MetaTags) Severity() string {
	return this.Tags.Get("SEVERITY")
}
func (this MetaTags) SelectOne() (cfg *SelectOneConfig, resID string, advanced bool, opts SelectConfigOption) {
	if value := this.Tags["SELECT_ONE"]; value != "" {
		cfg = &SelectOneConfig{}
		if value == "SELECT_ONE" {
			return
		}
		if tgs := this.TagsOf(value); len(tgs) == 1 && tgs["RES"] != "" {
			resID = tgs["RES"]
			if tgs = this.TagsOf(resID); tgs != nil {
				advanced = true
				resID = tgs["ID"]

				if tgs.Flag("NOT_ICON") {
					opts |= SelectConfigOptionNotIcon
				}
				if tgs.Flag("BLANK") {
					opts |= SelectConfigOptionAllowBlank
				}
				if tgs.Flag("BS") {
					opts |= SelectConfigOptionBottonSheet
				}
				if scopes := tgs.GetTags("SCOPES", tag_scanner.FlagPreserveKeys); scopes != nil {
					if cfg.RemoteDataResource == nil {
						cfg.RemoteDataResource = &DataResource{}
					}
					cfg.RemoteDataResource.Scopes = scopes.Flags()
				}
				if filters := tgs.GetTags("FILTERS", tag_scanner.FlagPreserveKeys); filters != nil {
					if cfg.RemoteDataResource == nil {
						cfg.RemoteDataResource = &DataResource{}
					}
					cfg.RemoteDataResource.Filters = filters
				}
				if scheme := tgs.GetString("SCHEME"); scheme != "" {
					if cfg.RemoteDataResource == nil {
						cfg.RemoteDataResource = &DataResource{}
					}
					cfg.RemoteDataResource.Scheme = scheme
				}
			}
			cfg.PrimaryField = tgs["PK_FIELD"]
			cfg.DisplayField = tgs["DISPLAY"]
			if blankValue := tgs.GetString("BLANK_VAL"); blankValue != "" {
				cfg.BlankFormattedValuer = func(ctx *Context, record interface{}) template.HTML {
					return template.HTML(tgs.GetString("BLANK_VAL"))
				}
			}
		} else {
			cfg.Collection = tag_scanner.KeyValuePairs(this.Scanner(), value)
		}
	}
	return
}

func (this MetaTags) SelectMany() (cfg *SelectManyConfig, resID string, advanced bool, opts SelectConfigOption) {
	if value := this.Tags["SELECT_MANY"]; value != "" {
		cfg = &SelectManyConfig{}
		if value == "SELECT_MANY" {
			return
		}
		if tgs := this.TagsOf(value); len(tgs) == 1 && tgs["RES"] != "" {
			resID = tgs["RES"]
			if tgs = this.TagsOf(resID); tgs != nil {
				advanced = true
				resID = tgs["ID"]

				if tgs.Flag("NOT_ICON") {
					opts |= SelectConfigOptionNotIcon
				}
				if tgs.Flag("BLANK") {
					opts |= SelectConfigOptionAllowBlank
				}
				if tgs.Flag("BS") {
					opts |= SelectConfigOptionBottonSheet
				}
			}
			var (
				display     = tgs["DISPLAY"]
				displayTags = this.Scanner().IsTags(display)
			)
			if displayTags {
				cfg.BottomSheetSelectedTemplate = display
			} else {
				cfg.DisplayField = display
			}
		} else {
			cfg.Collection = tag_scanner.KeyValuePairs(this.Scanner(), value)
		}
	}
	return
}

func (this MetaTags) UI() Tags {
	tags := this.GetTags("UI")
	if tags == nil {
		tags = make(Tags)
	}
	return tags
}

func (this MetaTags) Edit() (tags *MetaTagsEdit) {
	if t := this.GetTags("EDIT"); t != nil {
		tags = &MetaTagsEdit{t}
	}
	return
}

type MetaTagsEdit struct {
	Tags
}

func (this MetaTagsEdit) ReadOnly() bool {
	return this.Flag("RO")
}

func ParseMetaTags(tag reflect.StructTag) (tags MetaTags) {
	tags.ParseCallbackDefault(aorm.StructTag(tag), []string{"admin"}, func(dest Tags, n tag_scanner.Node) {
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

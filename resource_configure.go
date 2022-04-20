package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	uurl "github.com/ecletus/core/utils/url"
	"github.com/ecletus/roles"
	tag_scanner "github.com/unapu-go/tag-scanner"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/go-aorm/aorm"
)

func (this *Resource) ParseTagsFromUnderlineField(typ reflect.Type) *ResourceTags {
	var (
		getUnderlineField func(typ reflect.Type) (reflect.StructField, bool)
	)

	getUnderlineField = func(typ reflect.Type) (reflect.StructField, bool) {
		typ = indirectType(typ)

		underlineField, ok := typ.FieldByName("_")
		if ok && (typ.NumField() == 1 || len(underlineField.Index) == 1) {
			var tag Tags
			if tag.ParseDefault(aorm.StructTag(underlineField.Tag), "admin") {
				return underlineField, true
			}
		}

		for i, l := 0, typ.NumField(); i < l; i++ {
			f := typ.Field(i)
			var tag Tags
			if tag.ParseDefault(aorm.StructTag(f.Tag), "admin") {
				if tag.Flag("INHERIT") {
					return getUnderlineField(f.Type)
				}
			}
		}
		return reflect.StructField{}, false
	}

	if f, ok := getUnderlineField(typ); ok {
		tags := ParseResourceTags(f.Tag)
		return &tags
	}
	return nil
}

func (this *Resource) ParseSectionLayoutTags(name string, tags Tags, dst, from *SectionsLayout) {
	var (
		parseFrom = func(ctags tag_scanner.Map) (from2 *SectionsLayout, exclude []string) {
			value := ctags["_"]
			if value == "" {
				return from, nil
			}
			scanner := ctags.Scanner()
			if !scanner.IsTags(value) && value != "" {
				return this.Sections.Layouts.Layouts[value], nil
			}
			tags, flags := tag_scanner.Parse(ctags.Scanner(), value)
			exclude = tags.GetTags("EXCLUDE", tag_scanner.FlagPreserveKeys).Flags()
			if len(flags) > 0 {
				from2 = this.Sections.Layouts.Layouts[flags.Strings()[0]]
			}
			delete(ctags, "_")
			if from2 == nil {
				from2 = from
			}
			return
		}

		parse = func(id string, tags tag_scanner.Map) {
			var (
				from, exclude = parseFrom(tags)
				value         = dst.Get(id)
			)
			if from != nil && value == nil {
				value = NewCRUDSchemeSectionsLayout(name+"/"+id, from.Get(id))
				dst.Set(id, value)
			}
			ResourceTags{Tags: tags}.SetAttrsTo(&this.SectionsAttribute, value)
			if len(exclude) > 0 {
				value.Exclude(exclude...)
			}
		}
	)
	for key := range tags {
		tags := tags.GetTags(key)
		if tags == nil {
			continue
		}

		switch key {
		case "SCREEN", "PRINT":
			parse(strings.ToLower(key), tags)
		}
	}

	if from != nil {
		if dst.Screen == nil {
			dst.Screen = NewCRUDSchemeSectionsLayout(name+"/screen", from.Screen)
		}
		if dst.Print == nil {
			dst.Print = NewCRUDSchemeSectionsLayout(name+"/print", dst.Screen)
		}
	}
}

func (this *Resource) ParseSectionsLayouts(raw string) *Resource {
	var parseTags = this.ParseSectionLayoutTags
	tags := this.Tags.TagsOf(raw, tag_scanner.FlagPreserveKeys)
	if tags != nil {
		hasDefault, hasInline := false, false

		if this.Config.DefaultSectionsLayout != "" {
			this.Sections = NewDefaultSchemeSectionsLayout(
				NewSchemeSectionsLayouts(this.Config.DefaultSectionsLayout+"~default",
					NewSchemeSectionsLayoutsOptions{DefaultProvider: this.AllSectionsProvider}))

			if _, ok := tags[this.Config.DefaultSectionsLayout]; ok {
				parseTags("", tags.GetTags(this.Config.DefaultSectionsLayout), this.Sections.Default, nil)
				delete(tags, this.Config.DefaultSectionsLayout)
				hasDefault = true
			}
			if _, ok := tags[this.Config.DefaultSectionsLayout+".inline"]; ok {
				parseTags("", tags.GetTags(this.Config.DefaultSectionsLayout+".inline"), this.Sections.Inline, nil)
				delete(tags, this.Config.DefaultSectionsLayout+".inline")
				hasInline = true
			}
		}
		if _, ok := tags["default"]; ok {
			if !hasDefault {
				parseTags("", tags.GetTags("default"), this.Sections.Default, nil)
			}
			delete(tags, "default")
		}
		if _, ok := tags["inline"]; ok {
			if !hasInline {
				parseTags("", tags.GetTags("inline"), this.Sections.Inline, nil)
			}
			delete(tags, "inline")
		}

		for key := range tags {
			layoutTags := tags.GetTags(key)
			if layoutTags == nil {
				continue
			}
			layout := &SectionsLayout{Name: key}
			parseTags(key, layoutTags, layout, this.Sections.Default)
			this.Sections.Layouts.Layouts[key] = layout
		}
	}

	return this
}

func (this *Resource) configureParseSections(tags *ResourceTags) {
	tags.SetAttrsTo(&this.SectionsAttribute, this.Sections.Default.Screen)

	if this.Config.DefaultSectionsLayout == "" {
		if tags := tags.GetTags("PRINT"); tags != nil {
			ResourceTags{Tags: tags}.SetAttrsTo(&this.SectionsAttribute, this.Sections.Default.Print)
		}
		if tags := tags.GetTags("INLINE"); tags != nil {
			ResourceTags{Tags: tags}.SetAttrsTo(&this.SectionsAttribute, this.Sections.Inline.Screen)
		}
		if tags := tags.GetTags("INLINE_PRINT"); tags != nil {
			ResourceTags{Tags: tags}.SetAttrsTo(&this.SectionsAttribute, this.Sections.Inline.Print)
		}
	}

	this.ParseSectionsLayouts(tags.Get("SECTION_LAYOUTS"))
}

func (this *Resource) configure() {
	var (
		getUnderlineField func(typ reflect.Type) (reflect.StructField, bool)
		tags              *ResourceTags
	)

	getUnderlineField = func(typ reflect.Type) (reflect.StructField, bool) {
		typ = indirectType(typ)

		underlineField, ok := typ.FieldByName("_")
		if ok && (typ.NumField() == 1 || len(underlineField.Index) == 1) {
			var tag Tags
			if tag.ParseDefault(aorm.StructTag(underlineField.Tag), "admin") {
				return underlineField, true
			}
		}

		for i, l := 0, typ.NumField(); i < l; i++ {
			f := typ.Field(i)
			var tag Tags
			if tag.ParseDefault(aorm.StructTag(f.Tag), "admin") {
				if tag.Flag("INHERIT") {
					return getUnderlineField(f.Type)
				}
			}
		}
		return reflect.StructField{}, false
	}

	for _, name := range []string{"CreatedAt", "CreatedBy", "UpdatedAt", "UpdatedBy", "DeletedAt", "DeletedBy"} {
		if _, ok := this.ModelStruct.FieldsByName[name]; ok {
			this.Meta(&Meta{Name: name, ReadOnly: true})
		}
	}

	if tagsGetter, ok := this.Value.(ResourceTagsGetter); ok {
		tags = tagsGetter.AdminGetResourceTags(this)
	} else {
		tags = this.ParseTagsFromUnderlineField(this.ModelStruct.Type)
	}

	if len(this.Tags.Tags) == 0 {
		if tags != nil {
			this.Tags = tags
		}
	} else {
		this.Tags.Tags.Update(tags.Tags)
	}

	tags = this.Tags

	if uitags := this.Tags.GetTags("UI"); uitags != nil {
		this.UITags = uitags
	}

	if order := tags.PkOrder(); order != 0 {
		this.SetDefaultPrimaryKeyOrder(order)
	}

	if names := tags.ReadOnlyAttrs(); len(names) > 0 {
		this.PostInitialize(func() {
			for _, name := range names {
				this.Meta(&Meta{Name: name, ReadOnly: true})
			}
		})
	}
	if names := tags.SortAttrs(); len(names) > 0 {
		this.SortableAttrs(names...)
	} else if len(this.PrimaryFields) > 0 {
		var names []string
		for _, f := range this.PrimaryFields {
			names = append(names, f.Name)
		}
		this.SortableAttrs(names...)
	}

	this.configureParseSections(tags)

	if tags.ShowPage() {
		if !this.Sections.Default.Screen.Show.IsSet() {
			this.ShowAttrs(this.EditAttrs())
		}
	}
	if showMetaType := tags.Show(); showMetaType != "" {
		this.ShowAttrs(META_STRINGIFY)
		this.Meta(&Meta{Name: META_STRINGIFY, Type: showMetaType})
	} else if stringify, err := tags.Stringify(); stringify != nil {
		// template
		if stringify.Template != nil {
			// field
			this.Meta(&Meta{Name: META_STRINGIFY, Valuer: func(recorde interface{}, context *core.Context) interface{} {
				var w bytes.Buffer
				if err := stringify.Template.Execute(&w, recorde); err != nil {
					return "{ERROR: " + err.Error() + "}"
				}
				return w.String()
			}})
		} else if stringify.FieldName != "" {
			// field
			this.Meta(&Meta{Name: META_STRINGIFY, Valuer: func(recorde interface{}, context *core.Context) interface{} {
				return fmt.Sprint(reflect.Indirect(reflect.ValueOf(recorde)).FieldByName(stringify.FieldName).Interface())
			}})
		} else {
			// field
			this.Meta(&Meta{Name: META_STRINGIFY, Valuer: func(recorde interface{}, context *core.Context) interface{} {
				m := reflect.Indirect(reflect.ValueOf(recorde)).MethodByName(stringify.MethodName)
				result := m.Call([]reflect.Value{})[0].Interface()
				return fmt.Sprint(result)
			}})
		}
	} else if err != nil {
		log.Fatal(err)
	}
	if attrs := tags.Search(); attrs != nil {
		this.SearchAttrs(attrs...)
	}
	if attrs := tags.Order(); attrs != nil {
		this.SetOrder(attrs)
	}
	for _, parentPreload := range tags.ParentPreload() {
		switch parentPreload {
		case "*":
			this.Config.ParentPreload |=
				ParentPreloadIndex |
					ParentPreloadNew |
					ParentPreloadCreate |
					ParentPreloadShow |
					ParentPreloadEdit |
					ParentPreloadUpdate |
					ParentPreloadDelete |
					ParentPreloadAction
		case "INDEX":
			this.Config.ParentPreload |= ParentPreloadIndex
		case "NEW":
			this.Config.ParentPreload |= ParentPreloadNew
		case "CREATE":
			this.Config.ParentPreload |= ParentPreloadCreate
		case "SHOW":
			this.Config.ParentPreload |= ParentPreloadShow
		case "EDIT":
			this.Config.ParentPreload |= ParentPreloadEdit
		case "UPDATE":
			this.Config.ParentPreload |= ParentPreloadUpdate
		case "DELETE":
			this.Config.ParentPreload |= ParentPreloadDelete
		case "ACTION":
			this.Config.ParentPreload |= ParentPreloadAction
		}
	}

	this.PostInitialize(func() {
		if this.GetDefinedMeta(META_STRINGIFY) == nil {
			this.Meta(&Meta{
				Name: META_STRINGIFY,
				Type: "string",
				FormattedValuer: func(recorde interface{}, context *core.Context) *FormattedValue {
					return &FormattedValue{Record: recorde, Raw: recorde}
				},
			})
		}
		if this.ModelStruct.HasManyChild {
			this.MetaDisable("Parent")
			for _, name := range this.ModelStruct.FieldsByName["Parent"].Relationship.ForeignFieldNames {
				this.MetaDisable(name)
			}
		}
	})

	modelType := utils.ModelType(this.Value)

	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceInterface); ok {
				injector.ConfigureResource(this)
			}
		}
	}

	if injector, ok := this.Value.(resource.ConfigureResourceInterface); ok {
		injector.ConfigureResource(this)
	}

	// set primary fields as default invisible
	for _, f := range this.ModelStruct.PrimaryFields {
		this.Meta(&Meta{Name: f.Name, DefaultInvisible: true, ReadOnly: f.TagSettings.Flag("SERIAL")})
	}

	if this.Config.Alone {
		return
	}

	if !this.Config.Singleton && this.ControllerBuilder.IsDeleter() {
		this.Action(&Action{
			Name:     ActionDelete,
			Method:   http.MethodDelete,
			LabelKey: I18NGROUP + ".actions.delete",
			Type:     ActionSuperDanger,
			URL: func(record interface{}, context *Context, args ...interface{}) string {
				return this.GetContextURI(context, this.GetKey(record))
			},
			Modes: []string{"menu_item"},
			Visible: func(recorde interface{}, context *Context) bool {
				if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX || this.IsSoftDeleted(recorde) {
					return false
				}
				if this.HasRecordPermission(roles.Delete, context.Context, recorde).Deny() {
					return false
				}
				if f, ok := this.ModelStruct.FieldsByName[aorm.SoftDeletionDisableFieldDisabledAt]; ok {
					v := this.ModelStruct.InstanceOf(recorde, f.Name).FirstField().Interface()
					if !v.(time.Time).IsZero() {
						return false
					}
				}
				return true
			},
			RefreshURL: func(record interface{}, context *Context) string {
				return this.GetContextIndexURI(context)
			},
		})

		if !this.Config.BulkDeletionDisabled && this.ControllerBuilder.IsBulkDeleter() {
			this.Action(&Action{
				Name:     ActionBulkDelete,
				LabelKey: I18NGROUP + ".actions.bulk_delete",
				Method:   http.MethodPost,
				Type:     ActionSuperDanger,
				URL: func(record interface{}, context *Context, args ...interface{}) string {
					return this.GetContextIndexURI(context) + P_BULK_DELETE
				},
				Modes: []string{"index"},
				Visible: func(recorde interface{}, context *Context) bool {
					if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
						return false
					}
					return !this.IsSoftDeleted(recorde)
				},
				IndexVisible: func(context *Context) bool {
					if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
						return false
					}
					return true
				},
			})
		}

		if this.softDelete && this.ControllerBuilder.IsRestorer() {
			this.PostInitialize(this.configureRestorer)
		}
	}

	this.PostInitialize(this.configureAudited)

	this.AddMenu(&Menu{
		Name:     "Print",
		Label:    "Print",
		MdlIcon:  "print",
		LabelKey: I18NGROUP + ".menus." + PrintMenu,
		MakeLink: func(context *Context, args ...interface{}) string {
			uri, _ := context.PatchCurrentURL("print", uurl.Flag(true))
			return uri
		},
		MakeItemLink: func(context *Context, item interface{}, args ...interface{}) string {
			return context.Path(this.GetRecordURI(context, item)) + "?print"
		},
		EnabledFunc: func(menu *Menu, context *Context) bool {
			return !context.Type.Has(PRINT)
		},
	})
}

func (this *Resource) configureRestorer() {
	this.AddMenu(&Menu{
		Name:     A_DELETED_INDEX,
		Label:    M_DELETED,
		Icon:     "Delete",
		LabelKey: I18NGROUP + ".schemes." + A_DELETED_INDEX,
		MakeLink: func(context *Context, args ...interface{}) string {
			return this.GetContextIndexURI(context) + "/" + A_DELETED_INDEX
		},
		EnabledFunc: func(menu *Menu, context *Context) bool {
			return context.Resource == this && context.ResourceID == nil
		},
	})
	this.Action(&Action{
		Name:     A_RESTORE,
		LabelKey: I18NGROUP + ".actions.restore",
		Modes:    []string{"menu_item"},
		Method:   http.MethodPut,
		Type:     ActionDanger,
		Visible: func(record interface{}, context *Context) bool {
			if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
				return true
			}
			return false
		},
		FindRecord: func(s *Searcher) (rec interface{}, err error) {
			s.DB(s.DB().Unscoped())
			return s.FindOne()
		},
		Handler: func(arg *ActionArgument) error {
			ctx := arg.Context
			return ctx.WithTransaction(func() (err error) {
				return ctx.Resource.RestoreRecord(arg.Context, arg.Record)
			})
		},
		RefreshURL: func(record interface{}, context *Context) string {
			return this.GetContextIndexURI(context) + "/deleted_index"
		},
	})

	this.RegisterScheme(A_DELETED_INDEX, &SchemeConfig{
		Setup: func(scheme *Scheme) {
			scheme.DefaultFilter(&DBFilter{
				Name: "deleted",
				Handler: func(_ *Context, db *aorm.DB) (*aorm.DB, error) {
					return db.Where(aorm.IQ("_.deleted_at IS NOT NULL")).Unscoped(), nil
				},
			})
			scheme.NotMount = true
			scheme.i18nKey = I18NGROUP + ".schemes.deleted_index"
		},
	})
}

package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) configure() {
	if f, ok := indirectType(this.ModelStruct.Type).FieldByName("_"); ok && len(f.Index) == 1 {
		if tags := ParseResourceTags(f.Tag); len(this.Tags.TagSetting) == 0 {
			this.Tags = &tags
		} else {
			this.Tags.TagSetting.Update(tags.TagSetting)
		}
		tags := this.Tags

		var exclude = func(s []string) []string {
			for i, v := range s {
				s[i] = "-" + v
			}
			return s
		}
		if sections := tags.Attrs(); len(sections) > 0 {
			this.NESAttrs(sections)
		}
		if sections := tags.AttrsInclude(); len(sections) > 0 {
			this.NESAttrs(this.NewAttrs(), sections)
		}
		if sections := tags.AttrsIncludeBeginning(); len(sections) > 0 {
			this.NESAttrs(sections, this.NewAttrs())
		}
		if names := tags.AttrsExclude(); len(names) > 0 {
			this.NESAttrs(this.NewAttrs(), exclude(names))
		}
		if sections := tags.ShowAttrs(); len(sections) > 0 {
			this.ShowAttrs(sections)
		}
		if names := tags.ShowAttrsExclude(); len(names) > 0 {
			this.ShowAttrs(this.ShowAttrs(), exclude(names))
		}
		if sections := tags.ShowAttrsInclude(); len(sections) > 0 {
			this.ShowAttrs(this.ShowAttrs(), sections)
		}
		if sections := tags.ShowAttrsIncludeBeginning(); len(sections) > 0 {
			this.ShowAttrs(sections, this.ShowAttrs())
		}
		if sections := tags.NewAttrs(); len(sections) > 0 {
			this.NewAttrs(sections)
		}
		if names := tags.NewAttrsExclude(); len(names) > 0 {
			this.NewAttrs(this.NewAttrs(), exclude(names))
		}
		if sections := tags.NewAttrsInclude(); len(sections) > 0 {
			this.NewAttrs(this.NewAttrs(), sections)
		}
		if sections := tags.NewAttrsIncludeBeginning(); len(sections) > 0 {
			this.NewAttrs(sections, this.NewAttrs())
		}
		if sections := tags.EditAttrs(); len(sections) > 0 {
			this.EditAttrs(sections)
		}
		if names := tags.EditAttrsExclude(); len(names) > 0 {
			this.EditAttrs(this.EditAttrs(), exclude(names))
		}
		if sections := tags.EditAttrsInclude(); len(sections) > 0 {
			this.EditAttrs(this.EditAttrs(), sections)
		}
		if sections := tags.EditAttrsIncludeBeginning(); len(sections) > 0 {
			this.EditAttrs(sections, this.EditAttrs())
		}
		if sections := tags.IndexAttrs(); len(sections) > 0 {
			this.IndexAttrs(sections)
		}
		if names := tags.IndexAttrsExclude(); len(names) > 0 {
			this.IndexAttrs(this.IndexAttrs(), exclude(names))
		}
		if tags.ShowPage() {
			if len(this.Scheme.showSections) == 0 {
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
	}

	this.AfterRegister(func() {
		if this.GetDefinedMeta(META_STRINGIFY) == nil {
			this.Meta(&Meta{
				Name: META_STRINGIFY,
				Type: "string",
				FormattedValuer: func(recorde interface{}, context *core.Context) interface{} {
					return fmt.Sprint(recorde)
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
		this.Meta(&Meta{Name: f.Name, DefaultInvisible: true})
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
				return this.GetContextURI(context.Context, this.GetKey(record))
			},
			Modes: []string{"menu_item"},
			Visible: func(recorde interface{}, context *Context) bool {
				if context.RouteHandler != nil && context.RouteHandler.Name == A_DELETED_INDEX {
					return false
				}
				return !this.IsSoftDeleted(recorde)
			},
			RefreshURL: func(record interface{}, context *Context) string {
				return this.GetContextIndexURI(context.Context)
			},
		})

		if this.ControllerBuilder.IsBulkDeleter() {
			this.Action(&Action{
				Name:     ActionBulkDelete,
				LabelKey: I18NGROUP + ".actions.bulk_delete",
				Method:   http.MethodPost,
				Type:     ActionSuperDanger,
				URL: func(record interface{}, context *Context, args ...interface{}) string {
					return this.GetContextIndexURI(context.Context) + P_BULK_DELETE
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
			this.AfterRegister(this.configureRestorer)
		}
	}

	this.AfterRegister(this.configureAudited)
}

func (this *Resource) configureRestorer() {
	this.AddMenu(&Menu{
		Name:     A_DELETED_INDEX,
		Label:    M_DELETED,
		Icon:     "Delete",
		LabelKey: I18NGROUP + ".schemes." + A_DELETED_INDEX,
		MakeLink: func(context *Context, args ...interface{}) string {
			return this.GetContextIndexURI(context.Context) + "/" + A_DELETED_INDEX
		},
		Enabled: func(menu *Menu, context *Context) bool {
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
		Handler: func(argument *ActionArgument) error {
			var record = reflect.New(this.ModelStruct.Type).Interface()
			argument.Context.ResourceID.SetTo(record)
			return argument.Context.DB().ModelStruct(this.ModelStruct, record).Opt(aorm.OptStoreBlankField()).Unscoped().UpdateColumn(map[string]interface{}{
				"DeletedAt": nil,
				"DeletedByID": nil,
			}).Error
		},
		RefreshURL: func(record interface{}, context *Context) string {
			return this.GetContextIndexURI(context.Context) + "/deleted_index"
		},
	})

	this.RegisterScheme(A_DELETED_INDEX, &SchemeConfig{
		Setup: func(scheme *Scheme) {
			scheme.DefaultFilter(&DBFilter{
				Name: "deleted",
				Handler: func(context *core.Context, db *aorm.DB) (*aorm.DB, error) {
					return db.Where(aorm.IQ("{}.deleted_at IS NOT NULL")).Unscoped(), nil
				},
			})
			scheme.NotMount = true
			scheme.i18nKey = I18NGROUP + ".schemes.deleted_index"
		},
	})
}

package admin

import (
	"database/sql"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

func (this *Meta) updateMeta() {
	if this.Meta == nil {
		this.Meta = &resource.Meta{
			MetaName:         &resource.MetaName{this.Name, this.EncodedName},
			FieldName:        this.FieldName,
			Setter:           this.Setter,
			Valuer:           this.Valuer,
			FormattedValuer:  this.FormattedValuer,
			BaseResource:     this.BaseResource,
			ContextResourcer: this.ContextResourcer,
			Resource:         this.Resource,
			Permission:       this.Permission,
			Config:           this.Config,
			Required:         this.Required,
			Icon:             this.Icon,
		}
	} else if this.ProxyTo == nil {
		// proxy does not updates resource Meta
		this.Meta.Alias = this.Alias
		this.Meta.Name = this.Name
		this.Meta.FieldName = this.FieldName
		this.Meta.EncodedName = this.EncodedName
		this.Meta.Setter = this.Setter
		this.Meta.Valuer = this.Valuer
		this.Meta.FormattedValuer = this.FormattedValuer
		this.Meta.BaseResource = this.BaseResource
		this.Meta.Resource = this.Resource
		this.Meta.Permission = this.Permission
		this.Meta.Config = this.Config
		this.Meta.ContextResourcer = this.ContextResourcer
		this.Meta.Required = this.Required
		this.Meta.Icon = this.Icon
	} else {
		this.Meta.Setter = this.Setter
		this.Meta.Valuer = this.Valuer
		this.Meta.FormattedValuer = this.FormattedValuer
	}

	if this.ProxyTo == nil && this.Name == BASIC_META_ICON {
		this.Icon = true
		this.Meta.Icon = true
	}

	if this.EventDispatcher.GetDefinedDispatcher() == nil {
		this.EventDispatcher.SetDispatcher(this)
	}

	if this.BaseResource != nil {
		for _, cb := range this.BaseResource.Admin.onPreInitializeResourceMeta {
			cb(this)
		}
	}

	this.PreInitialize()
	this.tagsConfigure()

	if this.FieldStruct != nil {
		if injector, ok := reflect.New(indirectType(this.Typ)).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(this)
		}
	}

	if this.Virtual {
		if this.Valuer == nil {
			this.Valuer = func(i interface{}, context *core.Context) interface{} {
				return nil
			}
		}
		if this.Setter == nil {
			this.Setter = func(interface{}, *resource.MetaValue, *core.Context) error {
				return nil
			}
		}
	}

	this.Initialize()

	if this.Label != "" && this.DefaultLabel == "" && !strings.ContainsRune(this.Label, '.') {
		this.DefaultLabel = this.Label
	} else if this.DefaultLabel == "" {
		this.DefaultLabel = utils.HumanizeString(this.Name)
	}

	if this.Typ != nil {
		typ := indirectType(this.Typ)
		if typ.Kind() == reflect.Bool && this.Config == nil && ((this.Type == "" && this.Required) || this.Type == "select_one") {
			this.Type = ""
		}

		if this.FieldStruct != nil {
			if reflect.PtrTo(typ).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) {
				if typ.Kind() == reflect.Struct {
					typ = reflect.Indirect(reflect.New(typ)).Field(0).Type()
				}
			}

			// Set Meta Type
			if this.Type == "" {
				if relationship := this.FieldStruct.Relationship; relationship != nil {
					if relationship.Kind == "has_one" {
						this.Type = "single_edit"
					} else if relationship.Kind == "has_many" {
						this.Type = "collection_edit"
					} else if relationship.Kind == "belongs_to" {
						this.Type = "select_one"
					} else if relationship.Kind == "many_to_many" {
						this.Type = "select_many"
					}
				} else if _, ok := metaTypeConfigorMaps[indirectType(typ)]; !ok {
					switch typ.Kind() {
					default:
						if this.FieldStruct.TagSettings["TYPE"] == "date" {
							this.Type = "date"
						} else if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(typ.Kind().String()) {
							this.Type = "number"
						} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(typ.Kind().String()) {
							this.Type = "float"
						} else if _, ok := reflect.New(typ).Interface().(*time.Time); ok {
							this.Type = "datetime"
						} else {
							if typ.Kind() == reflect.Struct {
								if !this.Tags.Flag("SINGLE_EDIT_DISABLED") {
									this.Type = "single_edit"
								}
							} else if typ.Kind() == reflect.Slice {
								refelectType := typ.Elem()
								for refelectType.Kind() == reflect.Ptr {
									refelectType = refelectType.Elem()
								}
								if refelectType.Kind() == reflect.Struct {
									this.Type = "collection_edit"
								}
							}
						}
					}
				}
			} else {
				if relationship := this.FieldStruct.Relationship; relationship != nil {
					if (relationship.Kind == "has_one" || relationship.Kind == "has_many") && this.Meta.Setter == nil && (this.Type == "select_one" || this.Type == "select_many") {
						this.SetSetter(func(record interface{}, metaValue *resource.MetaValue, context *core.Context) error {
							reflectValue := reflect.Indirect(reflect.ValueOf(record))
							field := reflectValue.FieldByName(this.FieldName)

							if field.Kind() == reflect.Ptr {
								if field.IsNil() {
									field.Set(utils.NewValue(field.Type()).Elem())
								}

								for field.Kind() == reflect.Ptr {
									field = field.Elem()
								}
							}

							primaryKeys := utils.ToArray(metaValue.Value)
							if len(primaryKeys) > 0 {
								// set current field value to blank and replace it with new value
								field.Set(reflect.Zero(field.Type()))
								// 2020-03-31: nao havia tratamento de erro
								if err := context.DB().Where(primaryKeys).Find(field.Addr().Interface()).Error; err != nil {
									panic(err)
								}
							}

							if !aorm.ZeroIdOf(record) {
								// 2020-03-31: nao havia tratamento de erro
								if err := context.DB().Model(record).Association(this.FieldName).Replace(field.Interface()).Error(); err != nil {
									panic(err)
								}
								field.Set(reflect.Zero(field.Type()))
							}
							return nil
						})
					}
				}
			}

			// Set Meta Resource
			if this.Resource == nil {
				var typ reflect.Type
				if typ = aorm.AcceptableTypeForModelStructInterface(this.FieldStruct.Struct.Type); typ == nil {
					goto a
				}
				result := reflect.New(typ).Interface()

				if result != nil {
					var (
						res *Resource
						ok  bool
					)

					if res, ok = this.BaseResource.Resources[this.FieldStruct.Name]; !ok {
						if res, ok = this.BaseResource.GetAdmin().ResourcesByUID[utils.TypeId(typ)]; !ok {
							if res = this.BaseResource.FindResource(this.FieldStruct.Struct.Type); res == nil {
								if this.Tags.Managed() {
									if res = this.BaseResource.AddResource(&SubConfig{FieldName: this.FieldStruct.Name}, result); res == nil {
										goto a
									}
								} else if res = this.BaseResource.NewResource(&SubConfig{FieldName: this.FieldStruct.Name}, result); res == nil {
									goto a
								}
							}
						}
					}
					this.Resource = res
					this.Meta.Permission = this.Meta.Permission.Concat(res.Config.Permission)
				}
			} else if this.Config == nil && this.Resource.mounted {
				switch this.Type {
				case "select_one", "select_many":
					cfg := &SelectOneConfig{RemoteDataResource: &DataResource{}}
					cfg.Layout = BASIC_LAYOUT
					this.Config = cfg
				}
			}

			if this.Resource != nil && this.Resource != this.BaseResource {
				permission := this.Resource.Permission.Concat(this.Meta.Permission)
				this.Meta.Resource = this.Resource
				this.SetPermission(permission)
			}
		}
	}

a:
	this.FieldName = this.GetFieldName()

	if this.BaseResource.SingleEditMetas == nil {
		this.BaseResource.SingleEditMetas = make(map[string]*Meta)
	}

	if _, ok := this.BaseResource.SingleEditMetas[this.Name]; ok {
		if this.Type != "single_edit" {
			delete(this.BaseResource.SingleEditMetas, this.Name)
			this.Inline = false
		}
	} else if this.Type == "single_edit" {
		this.BaseResource.SingleEditMetas[this.Name] = this
		this.Inline = true
	}

	// call meta config's ConfigureMetaInterface
	if this.Config != nil {
		this.Config.ConfigureQorMeta(this)
	}

	// run meta configors
	if baseResource := this.BaseResource; baseResource != nil {
		if fc := baseResource.GetAdmin().metaConfigorMaps[this.Type]; fc != nil {
			fc(this)
		}
	}

	if this.Typ != nil {
		if configor, ok := metaTypeConfigorMaps[indirectType(this.Typ)]; ok {
			configor(this)
		}
	}

	// call field's ConfigureMetaInterface
	if this.FieldStruct != nil {
		switch this.FieldStruct.Struct.Type.Kind() {
		case reflect.Slice:
			i := reflect.New(indirectType(this.FieldStruct.Struct.Type.Elem())).Interface()
			if injector, ok := i.(resource.ConfigureMetaInterface); ok {
				injector.ConfigureQorMeta(this)
			}
		default:
			if injector, ok := reflect.New(this.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaInterface); ok {
				injector.ConfigureQorMeta(this)
			}
		}
	}

	if rmu, ok := this.BaseResource.Value.(ResourceMetaUpdater); ok {
		rmu.AdminMetaUpdate(this)
	}

	for _, cb := range this.BaseResource.metaUpdateCallbacks.GetAll(this.Name) {
		cb(this)
	}

	if this.Resource != nil {
		this.Resource.AddForeignMeta(this)
	}

	func(f ...func()) {
		this.afterUpdate = nil
		// call after update callbacks
		for _, f := range f {
			f()
		}
	}(this.afterUpdate...)
}

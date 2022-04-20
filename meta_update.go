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
	"github.com/go-aorm/aorm"
)

func (this *Meta) updateMeta() {
	if this.Disabled {
		return
	}

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
			Typ:              this.Typ,
			DefaultDeny:      this.DefaultDeny,
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
		this.Meta.DefaultDeny = this.DefaultDeny

		if this.Typ != nil {
			this.Meta.Typ = this.Typ
		}
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

	if this.Typ == nil {
		this.Typ = this.Meta.Typ
	}

	if this.FieldStruct != nil {
		if this.FieldStruct.IsReadOnly {
			this.ReadOnly = true
		}
	}

	this.tagsConfigure()

	if this.FieldStruct != nil && !this.Virtual {
		if injector, ok := reflect.New(indirectType(this.Typ)).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(this)
		}
	}

	for _, cb := range metaPreConfigorMaps {
		cb(this)
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

	this.Initialize(this.Virtual)

	if this.Meta.Setter == nil && (this.FieldStruct == nil || this.FieldStruct.Link != nil) {
		this.ReadOnly = true
	}

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
			if !aorm.CanFieldType(typ) && reflect.PtrTo(typ).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) {
				if typ.Kind() == reflect.Struct {
					typ = reflect.Indirect(reflect.New(typ)).Field(0).Type()
				}
			}

			// Set Meta Type
			if this.Type == "" {
				switch t := reflect.New(this.FieldStruct.Struct.Type).Interface().(type) {
				case MetaTyper:
					this.Type = t.AdminMetaType(this)
				case MetaSingleTyper:
					this.Type = t.AdminMetaType()
				case MetaRecordTyper:
					this.TypeHandler = func(meta *Meta, record interface{}, context *Context) string {
						v := reflect.ValueOf(record).Elem().FieldByIndex(this.FieldStruct.StructIndex)
						if v.Kind() != reflect.Ptr && v.CanAddr() {
							v = v.Addr()
						}
						return v.Interface().(MetaRecordTyper).AdminMetaType(meta, context)
					}
				default:
					if this.FieldStruct.IsEmbedded {
						this.Type = "single_edit"
					} else if relationship := this.FieldStruct.Relationship; relationship != nil {
						switch relationship.Kind {
						case aorm.HAS_ONE:
							this.Type = "single_edit"
						case aorm.HAS_MANY:
							this.Type = "collection_edit"
						case aorm.BELONGS_TO:
							this.Type = "select_one"
						case aorm.M2M:
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
									if !aorm.CanFieldType(typ) && !this.Tags.Flag("SINGLE_EDIT_DISABLED") {
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
				}
			} else {
				if relationship := this.FieldStruct.Relationship; relationship != nil {
					if relationship.Kind.Is(aorm.HAS_MANY, aorm.HAS_ONE) && this.Meta.Setter == nil && (this.Type == "select_one" || this.Type == "select_many") {
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
								if err := context.DB().Model(record).Association(this.FieldName).Replace(field.Interface()).Error; err != nil {
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
					if modes := this.Tags.Managed(); modes != nil {
						cfg := &Config{}
						if len(modes) != 0 {
							for _, mode := range modes {
								switch mode {
								case "RO":
									cfg.Controller = &struct {
										IndexSearchController
										ReadController
									}{}
								}
							}
						}
						res = this.BaseResource.AddResourceFieldConfig(this.FieldStruct.Name, result, cfg, true)
					} else if this.Meta.FieldStruct.IsChild || this.Meta.FieldStruct.IsEmbedded {
						res = this.BaseResource.NewResource(&SubConfig{FieldName: this.FieldStruct.Name}, result)
						res.metaEmbedded = true
						res.BaseResource = this.Meta.BaseResource.(*Resource)
					} else {
						if res, ok = this.BaseResource.Resources[this.FieldStruct.Name]; !ok {
							if res, ok = this.BaseResource.GetAdmin().ResourcesByUID[utils.TypeId(typ)]; !ok {
								if res = this.BaseResource.FindResource(this.FieldStruct.Struct.Type); res == nil {
									if res = this.BaseResource.NewResource(&SubConfig{FieldName: this.FieldStruct.Name}, result); res == nil {
										goto a
									}
									res.metaEmbedded = true
								}
							}
						}
					}
					this.Resource = res
				}
			} else if this.Config == nil && this.Resource.mounted {
				switch this.Type {
				case "select_one", "select_many":
					cfg := &SelectOneConfig{RemoteDataResource: &DataResource{}}
					cfg.Layout = BASIC_LAYOUT
					this.Config = cfg
				}
			}

			this.Meta.Resource = this.Resource
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

	for _, f := range this.updateCallbacks {
		f(this)
	}

	func(f ...func()) {
		this.afterUpdate = nil
		// call after update callbacks
		for _, f := range f {
			f()
		}
	}(this.afterUpdate...)
}

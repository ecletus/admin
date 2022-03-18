package admin

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/moisespsena-go/assetfs"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

type SelectOneConfigCallbackKey struct{}

func SetSelectOneConfigureCallback(res *Resource, cb func(cfg *SelectOneConfig)) {
	res.Data.Set(SelectOneConfigCallbackKey{}, cb)
}

type Select2SelectedItemTemplate struct {
	IconKey        string
	LabelKey       string
	LabelFormat    string
	DescriptionKey string
}

func (this Select2SelectedItemTemplate) Template() *JS {
	var s = `let v = "";`
	if this.IconKey != "" {
		s += `
	let icon = data["` + this.IconKey + `"];
	v += "<i class=\"material-icons\"";
    if (/\//.test(icon)) {
		v += "style=\"background-position-y:bottom;background-image: url('" + icon + "');background-repeat:no-repeat;background-size:contain\">";
	} else {
		v += ">" + icon;
	}
	v += "</i> ";`
	}
	if this.LabelKey != "" {
		s += `
	v += data["` + this.LabelKey + `"];`
	} else if this.LabelFormat != "" {
		s += `
	let write = function(s) { v += s };
	(function(data) {
		` + this.LabelFormat + `
	})(data);`
	}
	if this.DescriptionKey != "" {
		s += `
	v += "<small>"+data["` + this.DescriptionKey + `"]+"</small>";`
	}
	return RawJS(s + ";return v")
}

var SelectOne2ResultTemplateBasicHTMLWithIcon = RawJS(`
if (data.text) return data.text;
let icon = data.QorChooserOptions.getKey(data, data.QorChooserOptions.iconKey, data.icon || data.Icon),
	value = data.QorChooserOptions.getKey(data, data.QorChooserOptions.displayKey, data.text || data.HTML || data.html || data.value || data.Value),
	help = data.QorChooserOptions.getKey(data, data.QorChooserOptions.helpKey, data.help || data.description)
    v = "";
if (icon) {
	v += "<i class=\"material-icons\"";
    if (/\//.test(data.Icon)) {
		v += "style=\"background-position-y:bottom;background-image: url('" + icon + "');background-repeat:no-repeat;background-size:contain\">";
	} else {
		v += ">" + icon;
	}
	v += "</i> ";
}
v += value
if (help) {
	v += "<small>"+help+"</small>"
}
return $("<span>" + v + "</span>");
`)

type RemoteDataResourceConfig struct {
	Scopes []string
}

type DataResource struct {
	ResourceURL
	recordeUrl *ResourceURL
	Meta       *Meta
}

func NewDataResource(res *Resource) *DataResource {
	d := &DataResource{}
	d.Resource = res
	return d
}

func (d *DataResource) SetScheme(scheme string) *DataResource {
	d.Scheme = scheme
	return d
}

func (d *DataResource) With(f func(d *DataResource)) *DataResource {
	f(d)
	return d
}

func (d *DataResource) RecordeUrl() *ResourceURL {
	if d.recordeUrl == nil {
		d.recordeUrl = &ResourceURL{Resource: d.Resource, recorde: true}
	}
	return d.recordeUrl
}

func (d *DataResource) Dependency(dep ...interface{}) *DataResource {
	d.ResourceURL.Dependency(dep...)
	return d
}

func (d *DataResource) Filter(name string, value string) *DataResource {
	d.ResourceURL.Filter(name, value)
	return d
}

// SelectOneConfig meta configuration used for select one
type SelectOneConfig struct {
	Collection                    interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	Placeholder                   string
	AllowBlank                    bool
	DefaultCreating               bool
	SelectionTemplate             string
	Layout                        string
	Display                       string
	Scheme                        string
	SelectMode                    string // select, select_async, bottom_sheet
	PrimaryField                  string
	DisplayField                  string
	IconField                     string
	Select2ResultTemplate         *JS
	Select2SelectionTemplate      *JS
	BottomSheetSelectedTemplateJS *JS
	BottomSheetSelectedTemplate   string
	RemoteDataResource            *DataResource
	Remote                        bool
	RemoteNoCache                 bool
	RemoteURL                     string
	MakeRemoteURL                 func(*Context) string
	metaConfig
	getCollection        func(record interface{}, ctx *Context) [][]string
	Note                 string
	Basic                bool
	SelfExclude          bool
	SelfFilterParam      string
	meta                 *Meta
	BlankFormattedValuer func(ctx *Context, record interface{}) template.HTML
	EqFunc               func(a interface{}, b string) bool
}

func (cfg *SelectOneConfig) PrepareMetaContext(ctx *MetaContext, record interface{}) {
	if ctx.Context.Type.Has(INLINE, INDEX) {
		if ok := ctx.Context.Flag(resource.AutoLoadDisabled); !ok {
			ctx.Context.SetValue(resource.AutoLoadDisabled, true)
			ctx.DeferHandler(func() {
				ctx.Context.DelValue(resource.AutoLoadDisabled)
			})
		}
	}
}

func (cfg *SelectOneConfig) Eq(a interface{}, b string) bool {
	if cfg.EqFunc == nil {
		return fmt.Sprint(a) == b
	}
	return cfg.EqFunc(a, b)
}

func (cfg *SelectOneConfig) basic() {
	if cfg.Layout == "" && cfg.SelectMode != "bottom_sheet" {
		if cfg.meta.Resource.GetLayout(BASIC_LAYOUT_HTML_DESCRIPTION_WITH_ICON) != nil {
			cfg.Layout = BASIC_LAYOUT_HTML_DESCRIPTION_WITH_ICON
		} else if cfg.meta.Resource.GetLayout(BASIC_LAYOUT_HTML_WITH_ICON) != nil {
			cfg.Layout = BASIC_LAYOUT_HTML_WITH_ICON
		}
	}
	cfg.RemoteDataResource.RecordeUrl().Layout = cfg.Layout
	if cfg.SelectMode == "bottom_sheet" {
		if cfg.BottomSheetSelectedTemplate == "" {
			if cfg.DisplayField != "" {
				cfg.BottomSheetSelectedTemplate = "[[ " + cfg.DisplayField + " ]]"
			} else {
				var defaul = "[[ Value ]]"
				res := cfg.RemoteDataResource.Resource

				parse := func(t Tags) bool {
					if tmpl := t.GetString("SELECTED_TMPL"); tmpl != "" {
						defaul = tmpl
						return true
					} else if fieldName := t.GetString("SELECTED_FIELD"); fieldName != "" {
						defaul = "[[ " + fieldName + " ]]"
						return true
					} else if tmpl := t.GetString("SELECT_TMPL"); tmpl != "" {
						defaul = tmpl
						return true
					}
					return false
				}

				if !parse(cfg.meta.UITags) && res != nil {
					parse(res.UITags)
				}
				cfg.BottomSheetSelectedTemplate = defaul
			}
		}
	}
}

func (cfg *SelectOneConfig) With(f func(cfg *SelectOneConfig)) *SelectOneConfig {
	f(cfg)
	return cfg
}

func (cfg *SelectOneConfig) IsRemote() bool {
	return cfg.RemoteURL != "" || cfg.Remote || cfg.RemoteDataResource != nil
}

func (cfg *SelectOneConfig) HasDependency() bool {
	if cfg.RemoteURL != "" {
		if strings.ContainsRune(cfg.RemoteURL, '{') {
			return true
		}
	} else if cfg.RemoteDataResource != nil && cfg.RemoteDataResource.Dependencies != nil {
		return true
	}
	return false
}

// ToURLString Convert to URL string
func (cfg *SelectOneConfig) URL(context *Context) (urlS string) {
	if cfg.RemoteDataResource != nil {
		urlS = cfg.RemoteDataResource.URL(context)
	} else if cfg.MakeRemoteURL != nil {
		urlS = cfg.MakeRemoteURL(context)
	} else {
		urlS = cfg.RemoteURL
	}

	var params = url.Values{}

	if cfg.SelfExclude {
		var name string
		if cfg.SelfFilterParam == "" {
			name = "filter[exclude].Value"
		} else {
			name = cfg.SelfFilterParam
		}
		params.Add(name, "{*ID}")
	}

	if cfg.meta != nil {
		prefix := cfg.meta.Name + ":"
		for key, values := range context.Request.URL.Query() {
			if strings.HasPrefix(key, prefix) {
				key = key[len(prefix):]
				for _, v := range values {
					params.Add(key, v)
				}
			}
		}
	}

	params.Add(":no_actions", "true")

	if len(params) > 0 {
		if strings.ContainsRune(urlS, '?') {
			urlS += "&"
		} else {
			urlS += "?"
		}
		urlS += params.Encode()
	}
	return
}

// GetPlaceholder get placeholder
func (cfg *SelectOneConfig) GetPlaceholder(*Context) (template.HTML, bool) {
	return template.HTML(cfg.Placeholder), cfg.Placeholder != ""
}

// GetTemplate get template for selection template
func (cfg *SelectOneConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && cfg.SelectionTemplate != "" {
		return context.Asset(cfg.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// GetCollection get collections from select one meta
func (cfg *SelectOneConfig) GetCollection(value interface{}, context *Context) [][]string {
	if cfg.getCollection == nil {
		cfg.prepareDataSource(nil, nil, "!remote_data_selector")
	}

	if cfg.getCollection != nil {
		return cfg.getCollection(value, context)
	}
	return [][]string{}
}

func (cfg *SelectOneConfig) configure(res *Resource) (r *Resource) {
	r = res
	if cfg.IsRemote() {
		if cfg.RemoteDataResource == nil {
			cfg.RemoteDataResource = &DataResource{}
		}
		if res == nil {
			r = cfg.RemoteDataResource.Resource
		} else if cfg.RemoteDataResource.Resource == nil {
			cfg.RemoteDataResource.Resource = res
		}
		if cfg.RemoteDataResource.Layout == "" {
			cfg.RemoteDataResource.Layout = cfg.Layout
		}
		if cfg.RemoteDataResource.Display == "" {
			cfg.RemoteDataResource.Display = cfg.Display
		}
		if cfg.Basic {
			cfg.basic()
		}
		switch cfg.RemoteDataResource.Layout {
		case BASIC_LAYOUT_HTML_WITH_ICON, BASIC_LAYOUT_HTML, BASIC_LAYOUT:
			cfg.Select2ResultTemplate = SelectOne2ResultTemplateBasicHTMLWithIcon
			cfg.Select2SelectionTemplate = SelectOne2ResultTemplateBasicHTMLWithIcon
		}
	}

	if r != nil {
		tagS := r.Tags.GetString("UI_SELECT_TMPL")
		if tagS == "" {
			tagS = r.UITags.GetString("SELECT_TMPL")
		}

		if tagS != "" {
			if r.Tags.Scanner().IsTags(tagS) {
				tag := r.Tags.TagsOf(tagS)
				if selected := tag.GetString("SEL"); selected != "" {
					cfg.Select2SelectionTemplate = NewJS(selected)
				}
				if result := tag.GetString("RES"); result != "" {
					cfg.Select2ResultTemplate = NewJS(result)
				}
			} else {
				cfg.Select2ResultTemplate = NewJS(tagS)
				cfg.Select2SelectionTemplate = NewJS(tagS)
			}
		}
		if configure, ok := r.Data.Get(SelectOneConfigCallbackKey{}); ok {
			configure.(func(config *SelectOneConfig))(cfg)
		}
	}

	return
}

func (cfg *SelectOneConfig) Stringify(ctx *Context, record interface{}) string {
	res := cfg.meta.Resource
	if res == nil && cfg.RemoteDataResource != nil {
		res = cfg.RemoteDataResource.Resource
	}
	if res != nil {
		return ctx.StringifyRecord(cfg.RemoteDataResource.Resource, record)
	}
	return utils.ToString(record)
}

func (cfg *SelectOneConfig) Htmlify(ctx *Context, record interface{}) template.HTML {
	res := cfg.meta.Resource
	if res == nil && cfg.RemoteDataResource != nil {
		res = cfg.RemoteDataResource.Resource
	}
	if res != nil {
		return ctx.HtmlifyRecord(cfg.RemoteDataResource.Resource, record)
	}
	return template.HTML(utils.ToString(record))
}

// ConfigureQorMeta configure select one meta
func (cfg *SelectOneConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		if !cfg.AllowBlank && !meta.Meta.Required {
			cfg.AllowBlank = true
		}
		cfg.meta = meta
		meta.Resource = cfg.configure(meta.Resource)
		if cfg.BlankFormattedValuer != nil {
			meta.ForceShowZero = true
		}
		if cfg.IsRemote() {
			// Set FormattedValuer
			if meta.FormattedValuer == nil {
				if meta.Typ != nil && meta.Typ.Kind() == reflect.Slice {
					meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
						if record == nil {
							return nil
						}
						if record = meta.Value(context, record); record != nil {
							slice := reflect.ValueOf(record)
							ctx := GetContext(context)
							var result []string
							for l, i := slice.Len(), 0; i < l; i++ {
								result = append(result, string(ctx.HtmlifyRecord(cfg.RemoteDataResource.Resource, slice.Index(i).Interface())))
							}
							return (&FormattedValue{Record: record, Raw: record, SafeValue: strings.Join(result, ", ")}).SetNonZero()
						}
						return nil
					})
				} else {
					meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
						if record == nil {
							return nil
						}
						var v string

						value := meta.Value(context, record)
						fv := (&FormattedValue{Record: record, Raw: value}).SetNonZero()
						if value == nil {
							if cfg.BlankFormattedValuer != nil {
								fv.SafeValue = string(cfg.BlankFormattedValuer(ContextFromCoreContext(context), value))
							}
						} else {
							fv.SafeValue = string(GetContext(context).HtmlifyRecord(cfg.RemoteDataResource.Resource, value))
						}
						if fv.SafeValue != "" {
							ctx := ContextFromCoreContext(context)
							if ctx.RenderFlags.Has(CtxRenderEncode) {
								fv.Value = strip.StripTags(v)
							}
						}
						return fv
					})
				}
			}

			if cfg.RemoteDataResource != nil {
				cfg.RemoteDataResource.Meta = meta
			}
		}
		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			if meta.Typ != nil && meta.Typ.Kind() == reflect.Slice {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
					if record == nil {
						return nil
					}

					if value := meta.Value(context, record); value != nil {
						slice := reflect.ValueOf(value)
						var result []string
						for l, i := slice.Len(), 0; i < l; i++ {
							result = append(result, utils.Stringify(slice.Index(i).Interface()))
						}
						return (&FormattedValue{Record: record, Raw: value, SafeValue: strings.Join(result, ", ")}).SetNonZero()
					}
					return nil
				})
			} else if cfg.Collection != nil {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
					if record == nil {
						return nil
					}

					if value := meta.Value(context, record); value != nil {
						switch v := value.(type) {
						case string:
							items := cfg.getCollection(record, ContextFromCoreContext(context))
							for _, item := range items {
								if item[0] == v {
									return (&FormattedValue{Record: record, Raw: value, Value: item[1]}).SetNonZero()
								}
							}
							return &FormattedValue{Record: record, Raw: value, Value: v, IsZeroF: func(record, value interface{}) bool {
								return value.(string) == ""
							}}
						default:
							s := fmt.Sprint(v)
							items := cfg.getCollection(record, ContextFromCoreContext(context))
							for _, item := range items {
								if fmt.Sprint(item[0]) == s {
									return (&FormattedValue{Record: record, Raw: item[0], Value: item[1]}).SetNonZero()
								}
							}
							return &FormattedValue{Record: record, Raw: value, Value: s, Zero: s == ""}
						}
					}
					return nil
				})
			} else {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
					if record == nil {
						return nil
					}

					value := meta.Value(context, record)
					fv := &FormattedValue{Record: record, Raw: value, IsZeroF: func(record, value interface{}) bool {
						return value == nil
					}}

					if value == nil {
						if cfg.BlankFormattedValuer != nil {
							fv.SafeValue = string(cfg.BlankFormattedValuer(ContextFromCoreContext(context), value))
						} else {
							return nil
						}
					} else {
						fv.SafeValue = ContextFromCoreContext(context).Stringify(value)
					}

					if fv.SafeValue != "" {
						ctx := ContextFromCoreContext(context)
						if ctx.RenderFlags.Has(CtxRenderEncode) {
							fv.Value = strip.StripTags(fv.Value)
						}
					}
					return fv
				})
			}
		}

		cfg.prepareDataSource(meta.FieldStruct, meta.Resource, "!remote_data_selector")

		meta.Type = "select_one"
	}
}

func (cfg *SelectOneConfig) ConfigureQORAdminFilter(filter *Filter) {
	filter.Resource = cfg.configure(filter.Resource)
	var structField *aorm.StructField
	if filter.Field != nil {
		structField = filter.Field.Struct
	}
	cfg.prepareDataSource(structField, filter.Resource, "!remote_data_filter")

	if len(filter.Operations) == 0 {
		filter.Operations = []string{"eq"}
	}

	filter.Type = "select_one"
}

func (cfg *SelectOneConfig) FilterValue(filter *Filter, context *Context) interface{} {
	var (
		keyword string
	)

	arg := context.Searcher.filters[filter]
	if arg == nil || len(arg.Value.Values) == 0 {
		return nil
	}

	keyword = arg.Value.Values[0].FirstStringValue()

	if keyword != "" && cfg.RemoteDataResource != nil {
		result := cfg.RemoteDataResource.Resource.NewStruct(context.Context.Site)
		clone := context.Context.Clone()
		clone.SetRawDB(clone.DB().New())
		var err error
		if clone.ResourceID, err = cfg.RemoteDataResource.Resource.ParseID(keyword); err != nil {
			context.AddError(err)
			return nil
		}
		if cfg.RemoteDataResource.Resource.CrudScheme(clone, cfg.Scheme).FindOne(result) == nil {
			return result
		}
	}

	return keyword
}

func (cfg *SelectOneConfig) Meta() *Meta {
	return cfg.meta
}

func (cfg *SelectOneConfig) prepareDataSource(field *aorm.StructField, res *Resource, routePrefix string) {
	// Set GetCollection
	if cfg.Collection != nil {
		cfg.SelectMode = "select"

		switch cl := cfg.Collection.(type) {
		case []string:
			cfg.getCollection = func(interface{}, *Context) (results [][]string) {
				for _, value := range cl {
					results = append(results, []string{value, value})
				}
				return
			}
		case [][]string:
			cfg.getCollection = func(interface{}, *Context) [][]string {
				return cl
			}
		case func() []string:
			cfg.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl() {
					results = append(results, []string{value, value})
				}
				return
			}
		case func() [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl()
			}
		case func(*Context) []string:
			cfg.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl(context) {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(*Context) [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(context)
			}
		case func(*core.Context) []string:
			cfg.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl(context.Context) {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(*core.Context) [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(context.Context)
			}
		case func(interface{}, *core.Context) [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(record, context.Context)
			}
		case func(interface{}, *Context) [][]string:
			cfg.getCollection = cl

		case func(interface{}, *Context) []resource.IDLabeler:
			cfg.getCollection = func(record interface{}, context *Context) (res [][]string) {
				for _, item := range cl(record, context) {
					res = append(res, []string{item.GetID().String(), item.Label()})
				}
				return
			}
		case func(interface{}, *core.Context) []resource.IDLabeler:
			cfg.getCollection = func(record interface{}, context *Context) (res [][]string) {
				for _, item := range cl(record, context.Context) {
					res = append(res, []string{item.GetID().String(), item.Label()})
				}
				return
			}
		case func(*Context) []resource.IDLabeler:
			cfg.getCollection = func(record interface{}, context *Context) (res [][]string) {
				for _, item := range cl(context) {
					res = append(res, []string{item.GetID().String(), item.Label()})
				}
				return
			}
		default:
			panic(fmt.Errorf("Unsupported Collection format"))
		}
	}

	// Set GetCollection if normal select mode
	if cfg.getCollection == nil {
		if field != nil {
			switch t := reflect.New(indirectType(field.Struct.Type)).Interface().(type) {
			case SelectCollectionProvider:
				cfg.getCollection = func(interface{}, *Context) [][]string {
					return t.GetCollection()
				}
				return
			case SelectCollectionContextProvider:
				cfg.getCollection = func(_ interface{}, ctx *Context) [][]string {
					return t.GetCollection(ctx)
				}
				return
			case SelectCollectionAsyncResource:
				res = t.GetAsyncResource(cfg)
				cfg.meta.Resource = res
				cfg.meta.Meta.Resource = res
				cfg.RemoteDataResource = &DataResource{}
				cfg.RemoteDataResource.Resource = res
			default:
				typ := reflect.TypeOf(cfg.meta.BaseResource.Value)
				name := "Get" + field.Name + "Collection"
				if m, ok := typ.MethodByName(name); ok {
					index := m.Index

					// first arg is THIS
					switch m.Type.NumIn() {
					case 1:
						cfg.getCollection = func(record interface{}, ctx *Context) [][]string {
							var recordValue reflect.Value
							if record == nil {
								recordValue = reflect.New(indirectType(typ))
							} else {
								recordValue = reflect.ValueOf(record)
							}
							out := recordValue.Method(index).Call([]reflect.Value{})
							if len(out) == 2 {
								if err, ok := out[1].Interface().(error); ok && err != nil {
									ctx.AddError(err)
									return nil
								}
							}
							return out[0].Interface().([][]string)
						}
					case 3:
						cfg.getCollection = func(record interface{}, ctx *Context) [][]string {
							out := reflect.ValueOf(record).Method(index).Call([]reflect.Value{reflect.ValueOf(record), reflect.ValueOf(ctx)})
							if len(out) == 2 {
								if err, ok := out[1].Interface().(error); ok && err != nil {
									ctx.AddError(err)
									return nil
								}
							}
							return out[0].Interface().([][]string)
						}
					}

					cfg.SelectMode = "select"
					return
				} else if m, ok := typ.MethodByName(name + "MetaFactory"); ok {
					index := m.Index
					out := reflect.New(indirectType(typ)).Method(index).Call([]reflect.Value{reflect.ValueOf(cfg.meta)})
					provider := out[0].Interface().(SelectCollectionRecordContextProvider)
					cfg.getCollection = func(record interface{}, ctx *Context) [][]string {
						return provider.GetCollection(record, ctx)
					}
					cfg.SelectMode = "select"
					return
				}
			}
		}

		if cfg.RemoteDataResource == nil {
			cfg.RemoteDataResource = NewDataResource(res)
		}

		if cfg.PrimaryField == "" {
			cfg.PrimaryField = "ID"
		}

		if cfg.SelectMode == "" {
			cfg.SelectMode = "select_async"
		}

		cfg.getCollection = func(_ interface{}, context *Context) (results [][]string) {
			cloneContext := context.Clone()
			cloneContext.setResource(cfg.RemoteDataResource.Resource)
			searcher := &Searcher{Context: cloneContext}
			searcher.Scope(cfg.RemoteDataResource.Scopes...)
			searcher.Pagination.CurrentPage = -1
			searchResults, _ := searcher.Basic().ParseFindMany()
			reflectValues := reflect.Indirect(reflect.ValueOf(searchResults))

			for i := 0; i < reflectValues.Len(); i++ {
				var (
					value = reflectValues.Index(i).Interface()
					label = cloneContext.Resource.GetDefinedMeta(BASIC_META_LABEL).Value(cloneContext.Context, value)
				)
				results = append(results, []string{aorm.IdOf(value).String(), label.(string)})
			}
			return
		}
	}

	// Set GetCollection if normal select mode
	if cfg.EqFunc == nil && field != nil {
		switch t := reflect.New(indirectType(field.Struct.Type)).Interface().(type) {
		case SelectEqualer:
			cfg.EqFunc = func(a interface{}, b string) bool {
				return t.SelectOneItemEq(a, b)
			}
			return
		}
	}

	if res != nil && (cfg.SelectMode == "select_async" || cfg.SelectMode == "bottom_sheet") {
		if remoteDataResource := cfg.RemoteDataResource; remoteDataResource != nil {
			if !remoteDataResource.Resource.mounted {
				if remoteDataResource.Resource.Config.NotMount {
					remoteDataResource.Resource.MountTo(routePrefix + "!" + utils.ToParamString(field.Name))
				}
				/*remoteDataResource.Resource.params = path.Join(routePrefix, res.ToParam(), field.Name,
					fmt.Sprintf("%p", remoteDataResource.Resource))
				res.GetAdmin().RegisterResourceRouters(remoteDataResource.Resource,
					"create", "update", "read", "delete")*/
			}
		} else {
			panic(fmt.Errorf("RemoteDataResource not configured"))
		}
	}
}

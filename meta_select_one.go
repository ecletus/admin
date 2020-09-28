package admin

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/assetfs"
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
	getCollection   func(interface{}, *Context) [][]string
	Note            string
	Basic           bool
	SelfExclude     bool
	SelfFilterParam string
	meta            *Meta
}

func (cfg *SelectOneConfig) basic() {
	if cfg.Layout == "" {
		if cfg.meta.BaseResource.GetDefinedMeta(META_DESCRIPTIFY) != nil {
			cfg.Layout = BASIC_LAYOUT_HTML_DESCRIPTION_WITH_ICON
		} else {
			cfg.Layout = BASIC_LAYOUT_HTML_WITH_ICON
		}
	}
	cfg.RemoteDataResource.RecordeUrl().Layout = cfg.Layout
	if cfg.SelectMode == "bottom_sheet" {
		if cfg.BottomSheetSelectedTemplate == "" {
			if cfg.DisplayField != "" {
				cfg.BottomSheetSelectedTemplate = "[[& " + cfg.DisplayField + " ]]"
			} else  {
				var defaul = "[[& Value ]]"
				if cfg.RemoteDataResource != nil {
					if tmpl := cfg.RemoteDataResource.Resource.Tags.GetString("UI_SELECTED_TMPL"); tmpl != "" {
						defaul = tmpl
					}
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
			name = "filtersByName[exclude].Value"
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
		case "":
			cfg.RemoteDataResource.Layout = BASIC_LAYOUT_HTML_WITH_ICON
		}
	}

	if r != nil {
		if configure, ok := res.Data.Get(SelectOneConfigCallbackKey{}); ok {
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

		if cfg.IsRemote() {
			// Set FormattedValuer
			if meta.FormattedValuer == nil {
				if meta.Typ != nil && meta.Typ.Kind() == reflect.Slice {
					meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
						if record != nil {
							if record = meta.Value(context, record); record != nil {
								if record == nil {
									return ""
								}
								slice := reflect.ValueOf(record)
								ctx := GetContext(context)
								var result []string
								for l, i := slice.Len(), 0; i < l; i++ {
									result = append(result, string(ctx.HtmlifyRecord(cfg.RemoteDataResource.Resource, slice.Index(i).Interface())))
								}
								return template.HTML(strings.Join(result, ", "))
							}
						}
						return ""
					})
				} else {
					meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
						if record != nil {
							if record = meta.Value(context, record); record != nil {
								return GetContext(context).HtmlifyRecord(cfg.RemoteDataResource.Resource, record)
							}
						}
						return ""
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
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
					if record != nil {
						if record = meta.Value(context, record); record != nil {
							if record == nil {
								return ""
							}
							slice := reflect.ValueOf(record)
							var result []string
							for l, i := slice.Len(), 0; i < l; i++ {
								result = append(result, utils.Stringify(slice.Index(i).Interface()))
							}
							return template.HTML(strings.Join(result, ", "))
						}
					}
					return ""
				})
			} else if cfg.Collection != nil {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
					if value := meta.Value(context, record); value != nil {
						switch v := value.(type) {
						case string:
							items := cfg.getCollection(record, ContextFromCoreContext(context))
							for _, item := range items {
								if item[0] == v {
									return item[1]
								}
							}
							return v
						default:
							s := fmt.Sprint(v)
							items := cfg.getCollection(record, ContextFromCoreContext(context))
							for _, item := range items {
								if fmt.Sprint(item[0]) == s {
									return item[1]
								}
							}
							return s
						}
					}
					return ""
				})
			} else {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
					return ContextFromCoreContext(context).Stringify(meta.Value(context, record))
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
			searchResults, _ := searcher.Basic().FindMany()
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

type SelectCollectionProvider interface {
	GetCollection() [][]string
}

type SelectCollectionContextProvider interface {
	GetCollection(ctx *Context) [][]string
}
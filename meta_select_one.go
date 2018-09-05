package admin

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"

	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-assetfs"
)

var SelectOne2ResultTemplateBasicHTMLWithIcon = RawJS(`
if (data.text) return data.text;
var v = "";
if (data.Icon) {
	v += "<i class=\"material-icons\"";
    if (/\//.test(data.Icon)) {
		v += "style=\"background-position-y:bottom;background-image: url('" + data.Icon + "');background-repeat:no-repeat;background-size:contain\">";
	} else {
		v += ">" + data.Icon;
	}
	v += "</i> ";
}
if (data.HTML) {
	v += data.HTML;
} else if (data.Text) {
	v += data.Text;
}
return $("<span>" + v + "</span>");
`)

type RemoteDataResourceConfig struct {
	Scopes []string
}

type DataResource struct {
	ResourceURL
	recordeUrl *ResourceURL
	Dependency []interface{}
}

func NewDataResource(res *Resource) *DataResource {
	d := &DataResource{}
	d.Resource = res
	return d
}

func (d *DataResource) RecordeUrl() *ResourceURL {
	if d.recordeUrl == nil {
		d.recordeUrl = &ResourceURL{Resource: d.Resource, recorde: true}
	}
	return d.recordeUrl
}

func (d *DataResource) Filter(name string, value string) *DataResource {
	if d.Filters == nil {
		d.Filters = make(map[string]string)
	}
	d.Filters[name] = value
	return d
}

// SelectOneConfig meta configuration used for select one
type SelectOneConfig struct {
	Collection                  interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	Placeholder                 string
	AllowBlank                  bool
	DefaultCreating             bool
	SelectionTemplate           string
	Layout                      string
	Display                     string
	Scheme                      string
	SelectMode                  string // select, select_async, bottom_sheet
	PrimaryField                string
	DisplayField                string
	Select2ResultTemplate       *JS
	Select2SelectionTemplate    *JS
	BottomSheetSelectedTemplate string
	RemoteDataResource          *DataResource
	Remote                      bool
	RemoteURL                   string
	MakeRemoteURL               func(*Context) string
	metaConfig
	getCollection func(interface{}, *Context) [][]string
	Note          string
	Basic         bool
}

func (cfg *SelectOneConfig) basic() {
	if cfg.Layout == "" {
		cfg.Layout = BASIC_LAYOUT_HTML_WITH_ICON
	}
	if cfg.DisplayField == "" {
		cfg.DisplayField = "Value"
	}
	cfg.RemoteDataResource.RecordeUrl().Basic()
	if cfg.SelectMode == "bottom_sheet" {
		cfg.BottomSheetSelectedTemplate = "[[& Value ]]"
	}
}

func (cfg *SelectOneConfig) IsRemote() bool {
	return cfg.RemoteURL != "" || cfg.Remote || cfg.RemoteDataResource != nil
}

// ToURLString Convert to URL string
func (cfg *SelectOneConfig) URL(context *Context) string {
	if cfg.RemoteDataResource != nil {
		return cfg.RemoteDataResource.URL(context)
	}
	if cfg.MakeRemoteURL != nil {
		return cfg.MakeRemoteURL(context)
	}
	return cfg.RemoteURL
}

// GetPlaceholder get placeholder
func (cfg SelectOneConfig) GetPlaceholder(*Context) (template.HTML, bool) {
	return template.HTML(cfg.Placeholder), cfg.Placeholder != ""
}

// GetTemplate get template for selection template
func (cfg SelectOneConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
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

// ConfigureQorMeta configure select one meta
func (cfg *SelectOneConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		if cfg.IsRemote() {
			if cfg.RemoteDataResource == nil {
				cfg.RemoteDataResource = &DataResource{}
			}
			if cfg.RemoteDataResource.Resource == nil && meta.Resource != nil {
				cfg.RemoteDataResource.Resource = meta.Resource
			} else if cfg.RemoteDataResource.Resource != nil && meta.Resource == nil {
				meta.Resource = cfg.RemoteDataResource.Resource
			}
			if cfg.RemoteDataResource.Layout == "" {
				cfg.RemoteDataResource.Layout = cfg.Layout
			}
			if cfg.RemoteDataResource.Display == "" {
				cfg.RemoteDataResource.Display = cfg.Display
			}
			if len(cfg.RemoteDataResource.Dependency) == 0 {
				cfg.RemoteDataResource.Dependency = meta.Dependency
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

			// Set FormattedValuer
			if meta.FormattedValuer == nil {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
					if record != nil {
						record = meta.Value(context, record)
						return ContextFromQorContext(context).HtmlifyRecord(cfg.RemoteDataResource.Resource, record)
					}
					return nil
				})
			}
		}
		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
				return utils.Stringify(meta.Value(context, record))
			})
		}

		cfg.prepareDataSource(meta.FieldStruct, meta.baseResource, "!remote_data_selector")

		meta.Type = "select_one"
	}
}

func (cfg *SelectOneConfig) ConfigureQORAdminFilter(filter *Filter) {
	var structField *aorm.StructField
	if field, ok := core.FakeDB.NewScope(filter.Resource.Value).FieldByName(filter.Name); ok {
		structField = field.StructField
	}

	cfg.prepareDataSource(structField, filter.Resource, "!remote_data_filter")

	if len(filter.Operations) == 0 {
		filter.Operations = []string{"equal"}
	}
	filter.Type = "select_one"
}

func (cfg *SelectOneConfig) FilterValue(filter *Filter, context *Context) interface{} {
	var (
		prefix  = fmt.Sprintf("filters[%v].", filter.Name)
		keyword string
	)

	if metaValues, err := resource.ConvertFormToMetaValues(context.Context, context.Request, []resource.Metaor{}, prefix); err == nil {
		if metaValue := metaValues.Get("Value"); metaValue != nil {
			keyword = utils.ToString(metaValue.Value)
		}
	}

	if keyword != "" && cfg.RemoteDataResource != nil {
		result := cfg.RemoteDataResource.Resource.NewStruct(context.Context.Site)
		clone := context.Clone()
		clone.ResourceID = keyword
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
		case func() [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl()
			}
		case func(*Context) [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(context)
			}
		case func(interface{}, *core.Context) [][]string:
			cfg.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(record, context.Context)
			}
		case func() []string:
			cfg.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl() {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(*Context) []string:
			cfg.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl(context) {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(interface{}, *Context) [][]string:
			cfg.getCollection = cl
		default:
			utils.ExitWithMsg("Unsupported Collection format")
		}
	}

	// Set GetCollection if normal select mode
	if cfg.getCollection == nil {
		qorAdmin := res.GetAdmin()
		if cfg.RemoteDataResource == nil {
			if field != nil {
				fieldType := field.Struct.Type
				for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
					fieldType = fieldType.Elem()
				}
				cfg.RemoteDataResource = NewDataResource(qorAdmin.GetResourceByID(fieldType.Name()))
				if cfg.RemoteDataResource.Resource == nil {
					typInterface := reflect.New(fieldType).Interface()
					cfg.RemoteDataResource.Resource = res.AddResource(
						&SubConfig{FieldName: field.Struct.Name},
						typInterface,
						&Config{
							Param:     routePrefix + "!" + utils.ToParamString(field.Name),
							Invisible: true,
						})
				}
			}
		}

		if cfg.PrimaryField == "" {
			for _, primaryField := range cfg.RemoteDataResource.Resource.PrimaryFields {
				cfg.PrimaryField = primaryField.Name
				break
			}
		}

		if cfg.SelectMode == "" {
			cfg.SelectMode = "select_async"
		}

		cfg.getCollection = func(_ interface{}, context *Context) (results [][]string) {
			cloneContext := context.clone()
			cloneContext.setResource(cfg.RemoteDataResource.Resource)
			searcher := &Searcher{Context: cloneContext}
			searcher.Scope(cfg.RemoteDataResource.Scopes...)
			searcher.Pagination.CurrentPage = -1
			searchResults, _ := searcher.Basic().FindMany()
			reflectValues := reflect.Indirect(reflect.ValueOf(searchResults))

			for i := 0; i < reflectValues.Len(); i++ {
				value := reflectValues.Index(i).Interface()
				scope := context.DB.NewScope(value)
				label := cloneContext.Resource.GetDefinedMeta(BASIC_META_LABEL).GetValuer()(value, cloneContext.Context)
				results = append(results, []string{fmt.Sprint(scope.PrimaryKeyValue()), label.(string)})
			}
			return
		}
	}

	if res != nil && (cfg.SelectMode == "select_async" || cfg.SelectMode == "bottom_sheet") {
		if remoteDataResource := cfg.RemoteDataResource; remoteDataResource != nil {
			if !remoteDataResource.Resource.mounted {
				remoteDataResource.Resource.MountTo(routePrefix + "!" + utils.ToParamString(field.Name))
				/*remoteDataResource.Resource.params = path.Join(routePrefix, res.ToParam(), field.Name,
					fmt.Sprintf("%p", remoteDataResource.Resource))
				res.GetAdmin().RegisterResourceRouters(remoteDataResource.Resource,
					"create", "update", "read", "delete")*/
			}
		} else {
			utils.ExitWithMsg("RemoteDataResource not configured")
		}
	}
}

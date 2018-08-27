package admin

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"

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
	Resource       *Resource
	Scopes         []string
	Filters        map[string]string
	DynamicFilters func(context *Context, filters map[string]string)
	Layout         string
	Display        string
	Query          map[string]interface{}
	FormatURI      func(data *DataResource, context *Context, uri string) string
	Dependency     []interface{}
}

func NewDataResource(resource *Resource) *DataResource {
	return &DataResource{Resource: resource}
}

func (d *DataResource) Filter(name string, value string) *DataResource {
	if d.Filters == nil {
		d.Filters = make(map[string]string)
	}
	d.Filters[name] = value
	return d
}

// ToURLString Convert to URL string
func (d *DataResource) ToURLString(context *Context) string {
	var parents []string
	var query []string

	if len(d.Dependency) > 0 {
		for _, dep := range d.Dependency {
			switch dp := dep.(type) {
			case *DependencyParent:
				if len(parents) == 0 {
					parents = make([]string, d.Resource.PathLevel, d.Resource.PathLevel)
				}
				parents[dp.Meta.Resource.PathLevel] = "{" + dp.Meta.Name + "}"
			case *DependencyQuery:
				query = append(query, dp.Param+"={"+dp.Meta.Name+"}")
			}
		}
	}

	if len(parents) > 0 {
		parent := d.Resource
		for pathLevel := d.Resource.PathLevel - 1; pathLevel >= 0; pathLevel-- {
			parent = parent.ParentResource
			if parents[pathLevel] == "" {
				parents[pathLevel] = context.URLParam(parent.ParamIDName())
			}
		}
	}

	uri := d.Resource.GetContextIndexURI(context.Context, parents...)

	if d.FormatURI != nil {
		uri = d.FormatURI(d, context, uri)
	}

	if d.Layout != "" {
		query = append(query, P_LAYOUT+"="+d.Layout)
	}

	if d.Display != "" {
		query = append(query, P_DISPLAY+"="+d.Display)
	}

	for _, scope := range d.Scopes {
		query = append(query, "scopes="+scope)
	}

	for fname, fvalue := range d.Filters {
		query = append(query, "filters["+fname+"].Value="+fvalue)
	}

	if d.DynamicFilters != nil {
		dynamicFilters := make(map[string]string)
		d.DynamicFilters(context, dynamicFilters)

		for fname, fvalue := range dynamicFilters {
			query = append(query, "filters["+fname+"].Value="+fvalue)
		}
	}

	if len(query) > 0 {
		uri += "?" + strings.Join(query, "&")
	}

	return uri
}

// SelectOneConfig meta configuration used for select one
type SelectOneConfig struct {
	Collection               interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	Placeholder              string
	AllowBlank               bool
	DefaultCreating          bool
	SelectionTemplate        string
	Layout                   string
	Display                  string
	Scheme                   string
	SelectMode               string // select, select_async, bottom_sheet
	PrimaryField             string
	Select2ResultTemplate    *JS
	Select2SelectionTemplate *JS
	RemoteDataResource       *DataResource
	Remote                   bool
	RemoteURL                string
	MakeRemoteURL            func(*Context) string
	metaConfig
	getCollection func(interface{}, *Context) [][]string
	Note          string
}

func (selectOneConfig *SelectOneConfig) IsRemote() bool {
	return selectOneConfig.RemoteURL != "" || selectOneConfig.Remote || selectOneConfig.RemoteDataResource != nil
}

// ToURLString Convert to URL string
func (selectOneConfig *SelectOneConfig) ToURLString(context *Context) string {
	if selectOneConfig.RemoteDataResource != nil {
		return selectOneConfig.RemoteDataResource.ToURLString(context)
	}
	if selectOneConfig.MakeRemoteURL != nil {
		return selectOneConfig.MakeRemoteURL(context)
	}
	return selectOneConfig.RemoteURL
}

// GetPlaceholder get placeholder
func (selectOneConfig SelectOneConfig) GetPlaceholder(*Context) (template.HTML, bool) {
	return template.HTML(selectOneConfig.Placeholder), selectOneConfig.Placeholder != ""
}

// GetTemplate get template for selection template
func (selectOneConfig SelectOneConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && selectOneConfig.SelectionTemplate != "" {
		return context.Asset(selectOneConfig.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// GetCollection get collections from select one meta
func (selectOneConfig *SelectOneConfig) GetCollection(value interface{}, context *Context) [][]string {
	if selectOneConfig.getCollection == nil {
		selectOneConfig.prepareDataSource(nil, nil, "!remote_data_selector")
	}

	if selectOneConfig.getCollection != nil {
		return selectOneConfig.getCollection(value, context)
	}
	return [][]string{}
}

// ConfigureQorMeta configure select one meta
func (selectOneConfig *SelectOneConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		if selectOneConfig.IsRemote() {
			if selectOneConfig.RemoteDataResource == nil {
				selectOneConfig.RemoteDataResource = &DataResource{}
			}
			if selectOneConfig.RemoteDataResource.Resource == nil && meta.Resource != nil {
				selectOneConfig.RemoteDataResource.Resource = meta.Resource
			} else if selectOneConfig.RemoteDataResource.Resource != nil && meta.Resource == nil {
				meta.Resource = selectOneConfig.RemoteDataResource.Resource
			}
			if selectOneConfig.RemoteDataResource.Layout == "" {
				selectOneConfig.RemoteDataResource.Layout = selectOneConfig.Layout
			}
			if selectOneConfig.RemoteDataResource.Display == "" {
				selectOneConfig.RemoteDataResource.Display = selectOneConfig.Display
			}
			if len(selectOneConfig.RemoteDataResource.Dependency) == 0 {
				selectOneConfig.RemoteDataResource.Dependency = meta.Dependency
			}

			switch selectOneConfig.RemoteDataResource.Layout {
			case BASIC_LAYOUT_HTML_WITH_ICON, BASIC_LAYOUT_HTML, BASIC_LAYOUT:
				selectOneConfig.Select2ResultTemplate = SelectOne2ResultTemplateBasicHTMLWithIcon
				selectOneConfig.Select2SelectionTemplate = SelectOne2ResultTemplateBasicHTMLWithIcon
			default:
				selectOneConfig.RemoteDataResource.Layout = BASIC_LAYOUT_HTML_WITH_ICON
			}

			// Set FormattedValuer
			if meta.FormattedValuer == nil {
				meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
					if record != nil {
						record = meta.GetValuer()(record, context)
						return ContextFromQorContext(context).HtmlifyRecord(selectOneConfig.RemoteDataResource.Resource, record)
					}
					return nil
				})
			}
		}
		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
				return utils.Stringify(meta.GetValuer()(record, context))
			})
		}

		selectOneConfig.prepareDataSource(meta.FieldStruct, meta.baseResource, "!remote_data_selector")

		meta.Type = "select_one"
	}
}

func (selectOneConfig *SelectOneConfig) ConfigureQORAdminFilter(filter *Filter) {
	var structField *aorm.StructField
	if field, ok := core.FakeDB.NewScope(filter.Resource.Value).FieldByName(filter.Name); ok {
		structField = field.StructField
	}

	selectOneConfig.prepareDataSource(structField, filter.Resource, "!remote_data_filter")

	if len(filter.Operations) == 0 {
		filter.Operations = []string{"equal"}
	}
	filter.Type = "select_one"
}

func (selectOneConfig *SelectOneConfig) FilterValue(filter *Filter, context *Context) interface{} {
	var (
		prefix  = fmt.Sprintf("filters[%v].", filter.Name)
		keyword string
	)

	if metaValues, err := resource.ConvertFormToMetaValues(context.Context, context.Request, []resource.Metaor{}, prefix); err == nil {
		if metaValue := metaValues.Get("Value"); metaValue != nil {
			keyword = utils.ToString(metaValue.Value)
		}
	}

	if keyword != "" && selectOneConfig.RemoteDataResource != nil {
		result := selectOneConfig.RemoteDataResource.Resource.NewStruct(context.Context.Site)
		clone := context.Clone()
		clone.ResourceID = keyword
		if selectOneConfig.RemoteDataResource.Resource.CrudScheme(clone, selectOneConfig.Scheme).FindOne(result) == nil {
			return result
		}
	}

	return keyword
}

func (selectOneConfig *SelectOneConfig) prepareDataSource(field *aorm.StructField, res *Resource, routePrefix string) {
	// Set GetCollection
	if selectOneConfig.Collection != nil {
		selectOneConfig.SelectMode = "select"

		switch cl := selectOneConfig.Collection.(type) {
		case []string:
			selectOneConfig.getCollection = func(interface{}, *Context) (results [][]string) {
				for _, value := range cl {
					results = append(results, []string{value, value})
				}
				return
			}
		case [][]string:
			selectOneConfig.getCollection = func(interface{}, *Context) [][]string {
				return cl
			}
		case func() [][]string:
			selectOneConfig.getCollection = func(record interface{}, context *Context) [][]string {
				return cl()
			}
		case func(*Context) [][]string:
			selectOneConfig.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(context)
			}
		case func(interface{}, *core.Context) [][]string:
			selectOneConfig.getCollection = func(record interface{}, context *Context) [][]string {
				return cl(record, context.Context)
			}
		case func() []string:
			selectOneConfig.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl() {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(*Context) []string:
			selectOneConfig.getCollection = func(record interface{}, context *Context) (results [][]string) {
				for _, value := range cl(context) {
					results = append(results, []string{value, value})
				}
				return
			}
		case func(interface{}, *Context) [][]string:
			selectOneConfig.getCollection = cl
		default:
			utils.ExitWithMsg("Unsupported Collection format")
		}
	}

	// Set GetCollection if normal select mode
	if selectOneConfig.getCollection == nil {
		qorAdmin := res.GetAdmin()
		if selectOneConfig.RemoteDataResource == nil {
			if field != nil {
				fieldType := field.Struct.Type
				for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
					fieldType = fieldType.Elem()
				}
				selectOneConfig.RemoteDataResource = NewDataResource(qorAdmin.GetResourceByID(fieldType.Name()))
				if selectOneConfig.RemoteDataResource.Resource == nil {
					typInterface := reflect.New(fieldType).Interface()
					selectOneConfig.RemoteDataResource.Resource = res.AddResource(
						&SubConfig{FieldName: field.Struct.Name},
						typInterface,
						&Config{
							Param:     routePrefix + "!" + utils.ToParamString(field.Name),
							Invisible: true,
						})
				}
			}
		}

		if selectOneConfig.PrimaryField == "" {
			for _, primaryField := range selectOneConfig.RemoteDataResource.Resource.PrimaryFields {
				selectOneConfig.PrimaryField = primaryField.Name
				break
			}
		}

		if selectOneConfig.SelectMode == "" {
			selectOneConfig.SelectMode = "select_async"
		}

		selectOneConfig.getCollection = func(_ interface{}, context *Context) (results [][]string) {
			cloneContext := context.clone()
			cloneContext.setResource(selectOneConfig.RemoteDataResource.Resource)
			searcher := &Searcher{Context: cloneContext}
			searcher.Scope(selectOneConfig.RemoteDataResource.Scopes...)
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

	if res != nil && (selectOneConfig.SelectMode == "select_async" || selectOneConfig.SelectMode == "bottom_sheet") {
		if remoteDataResource := selectOneConfig.RemoteDataResource; remoteDataResource != nil {
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

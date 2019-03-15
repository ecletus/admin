package admin

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mime/multipart"

	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/moisespsena-go/aorm"
	"gopkg.in/fatih/set.v0"
)

// PaginationPageCount default pagination page count
var PaginationPageCount = 20

type scopeFunc func(db *aorm.DB, context *core.Context) *aorm.DB

// Pagination is used to hold pagination related information when rendering tables
type Pagination struct {
	Total       int
	Pages       int
	CurrentPage int
	PerPage     int
}

type ImmutableScopes struct {
	set set.Interface
}

func (is *ImmutableScopes) Has(names ...interface{}) bool {
	return is.set != nil && is.set.Has(names...)
}

func (is *ImmutableScopes) List() (items []string) {
	if is.set != nil {
		for _, v := range is.set.List() {
			items = append(items, v.(string))
		}
	}
	return
}

// Searcher is used to search results
type Searcher struct {
	*Context
	scopes        []*Scope
	filters       map[*Filter]*resource.MetaValues
	Pagination    Pagination
	CurrentScopes ImmutableScopes
	Layout        string
}

func (s *Searcher) DefaulLayout(layout ...string) {
	if s.Layout == "" {
		if len(layout) == 0 || layout[0] == "" {
			layout = []string{s.Type.Clear(DELETED).S()}
		}
		s.Layout = layout[0]
	}
}

func (s *Searcher) Basic() *Searcher {
	s.Layout = "basic"
	return s
}

func (s *Searcher) Readonly() *Searcher {
	s.Layout = "readonly"
	return s
}

func (s *Searcher) ForUpdate() *Searcher {
	s.Layout = ""
	return s
}

func (s *Searcher) IsReadonly() bool {
	return s.Layout == "readonly"
}

func (s *Searcher) clone() *Searcher {
	return &(*s)
}

// Page set current page, if current page equal -1, then show all records
func (s *Searcher) Page(num int) *Searcher {
	s.Pagination.CurrentPage = num
	return s
}

// PerPage set pre page count
func (s *Searcher) PerPage(num int) *Searcher {
	s.Pagination.PerPage = num
	return s
}

// Scope filter with defined scopes
func (s *Searcher) Scope(names ...string) *Searcher {
	newSearcher := s.clone()
	scopesSet := set.New(set.NonThreadSafe)

	for _, name := range names {
		for _, scope := range s.Scheme.scopes {
			if scope.Name == name {
				scopesSet.Add(name)

				if !scope.Default {
					newSearcher.scopes = append(newSearcher.scopes, scope)
					break
				}
			}
		}
	}

	newSearcher.CurrentScopes = ImmutableScopes{scopesSet}
	return newSearcher
}

// Filter filter with defined filtersByName, filter with columns value
func (s *Searcher) Filter(filter *Filter, values *resource.MetaValues) *Searcher {
	newSearcher := s.clone()
	if newSearcher.filters == nil {
		newSearcher.filters = map[*Filter]*resource.MetaValues{}
	}
	newSearcher.filters[filter] = values
	return newSearcher
}

// FindMany find many records based on current conditions
func (s *Searcher) FindMany() (interface{}, error) {
	context := s.parseContext()

	if context.HasError() {
		return nil, context.Errors
	}

	return s.Resource.CrudScheme(context, s.Scheme).FindManyLayoutOrDefault(s.Layout)
}

// FindOne find one record based on current conditions
func (s *Searcher) FindOne() (interface{}, error) {
	var (
		err     error
		context = s.parseContext()
		result  = s.Resource.NewStruct(s.Site)
	)

	if context.HasError() {
		return result, context.Errors
	}

	err = s.Resource.CrudScheme(context, s.Scheme).SetLayoutOrDefault(s.Layout).FindOne(result)
	return result, err
}

var filterRegexp = regexp.MustCompile(`^filtersByName\[(.*?)\]`)

func (s *Searcher) callScopes(context *core.Context) *core.Context {
	db := context.DB

	// call default scopes
	for _, scope := range s.Resource.scopes {
		if scope.Default {
			db = scope.Handler(db, s, context)
		}
	}

	// call scopes
	for _, scope := range s.scopes {
		db = scope.Handler(db, s, context)
	}

	// call filtersByName
	if s.filters != nil {
		for filter, value := range s.filters {
			if filter.Handler != nil {
				filterArgument := &FilterArgument{
					Filter:   filter,
					Scheme:   s.Scheme,
					Value:    value,
					Context:  context,
					Resource: s.Resource,
				}
				db = filter.Handler(db, filterArgument)
			}
		}
	}

	// add order by
	if orderBy := context.GetFormOrQuery("order_by"); orderBy != "" {
		if regexp.MustCompile("^[a-zA-Z_]+$").MatchString(orderBy) {
			if field, ok := db.NewScope(s.Context.Resource.Value).FieldByName(strings.TrimSuffix(orderBy, "_desc")); ok {
				if strings.HasSuffix(orderBy, "_desc") {
					db = db.Order(field.DBName+" DESC", true)
				} else {
					db = db.Order(field.DBName, true)
				}
			}
		}
	}

	context.DB = db

	// call search
	if keyword := context.GetFormOrQuery("keyword"); keyword != "" {
		if sh := s.Scheme.CurrentSearchHandler(); sh != nil {
			context.DB = sh(keyword, context)
		}
	}

	return context
}

func (s *Searcher) FilterRaw(data map[string]string) *Searcher {
	params := url.Values{}
	for key, value := range data {
		params.Add("filtersByName["+key+"].Value", value)
	}

	return s.FilterFromParams(params, nil)
}

func (s *Searcher) FilterFromParams(params url.Values, form *multipart.Form) *Searcher {
	searcher := s

	for key := range params {
		if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
			var prefix = fmt.Sprintf("filtersByName[%v].", matches[1])
			if filter, ok := s.Scheme.filtersByName[matches[1]]; ok {
				if metaValues, err := resource.ConvertFormDataToMetaValues(s.Context.Context, params, form, []resource.Metaor{}, prefix); err == nil {
					searcher = searcher.Filter(filter, metaValues)
				}
			}
		}
	}

	return searcher
}

func (s *Searcher) FilterRawPairs(args ...string) *Searcher {
	data := make(map[string]string)
	l := len(args)
	for i := 0; i < l; i += 2 {
		data[args[i]] = args[i+1]
	}
	return s.FilterRaw(data)
}

func (s *Searcher) parseContext() *core.Context {
	var (
		searcher = s.clone()
		context  = s.Context.Context.Clone()
	)

	if s.Scheme == nil {
		s.Scheme = s.Resource.Scheme
	}

	context.SetDB(context.DB.Order(s.Scheme.CurrentOrders()))
	context = s.Scheme.ApplyDefaultFilters(context)

	if context != nil && context.Request != nil {
		var query = context.Request.URL.Query()
		// parse scopes
		if scopes, ok := query["scopes"]; ok {
			searcher = searcher.Scope(scopes...)
		}
		searcher = searcher.FilterFromParams(query, context.Request.MultipartForm)

		if savingName := query.Get("filter_saving_name"); savingName != "" {
			var filters []SavedFilter
			requestURL := context.Request.URL
			requestURLQuery := context.Request.URL.Query()
			requestURLQuery.Del("filter_saving_name")
			requestURL.RawQuery = requestURLQuery.Encode()
			newFilters := []SavedFilter{{Name: savingName, URL: requestURL.String()}}
			if context.AddError(s.Admin.settings.Get("saved_filters", &filters, searcher.Context)); !context.HasError() {
				for _, filter := range filters {
					if filter.Name != savingName {
						newFilters = append(newFilters, filter)
					}
				}

				context.AddError(s.Admin.settings.Save("saved_filters", newFilters, searcher.Resource, context.CurrentUser(), searcher.Context))
			}
		}

		if savingName := query.Get("delete_saved_filter"); savingName != "" {
			var filters, newFilters []SavedFilter
			if context.AddError(s.Admin.settings.Get("saved_filters", &filters, searcher.Context)); !context.HasError() {
				for _, filter := range filters {
					if filter.Name != savingName {
						newFilters = append(newFilters, filter)
					}
				}

				context.AddError(s.Admin.settings.Save("saved_filters", newFilters, searcher.Resource, context.CurrentUser(), searcher.Context))
			}
		}
	}

	s.Scheme.PrepareContext(context)
	searcher.callScopes(context)

	db := context.DB

	// pagination
	context.DB = db.Model(s.Resource.Value).Set("qor:getting_total_count", true)
	if err := s.Resource.CrudScheme(context, s.Scheme).SetLayoutOrDefault(s.Layout).FindMany(&s.Pagination.Total); err != nil {
		context.AddError(err)
		return context
	}

	if s.Pagination.CurrentPage == 0 {
		if s.Context.Request != nil {
			if page, err := strconv.Atoi(s.Context.Request.URL.Query().Get("page")); err == nil {
				s.Pagination.CurrentPage = page
			}
		}

		if s.Pagination.CurrentPage == 0 {
			s.Pagination.CurrentPage = 1
		}
	}

	if s.Pagination.PerPage == 0 {
		if perPage, err := strconv.Atoi(s.Context.Request.URL.Query().Get("per_page")); err == nil {
			s.Pagination.PerPage = perPage
		} else if s.Resource.Config.PageCount > 0 {
			s.Pagination.PerPage = s.Resource.Config.PageCount
		} else {
			s.Pagination.PerPage = PaginationPageCount
		}
	}

	if s.Pagination.CurrentPage > 0 {
		s.Pagination.Pages = (s.Pagination.Total-1)/s.Pagination.PerPage + 1

		db = db.Limit(s.Pagination.PerPage).Offset((s.Pagination.CurrentPage - 1) * s.Pagination.PerPage)
	}

	context.DB = db

	return context
}

type filterField struct {
	Field     *FieldFilter
	Operation string
	Typ       reflect.Type
}

func (f filterField) Apply(arg interface{}) (query string, argx interface{}) {
	op := f.Operation
	fieldName := f.Field.QueryField()

	var cb = func() string {
		return fieldName + " " + op + " ?"
	}

	if f.Field.Struct.Struct.Type.Kind() == reflect.String {
		cb = func() string {
			return "UPPER(" + fieldName + ") " + op + " UPPER(?)"
		}
	} else if op == "" {
		op = "eq"
	}
	switch op {
	case "eq", "equal":
		op = "="
	case "ne":
		op = "!="
	case "btw":
		op = "BETWEEN"
		cb = func() string {
			return fieldName + " BETWEEN ? AND ?"
		}
	case "gt":
		op = ">"
	case "lt":
		op = "<"
	default:
		if f.Field.Struct.Struct.Type.Kind() == reflect.String {
			op = "LIKE"
			arg = "%" + arg.(string) + "%"
		}
	}
	return cb(), f.Field.FormatTerm(arg)
}

func filterResourceByFields(res *Resource, filterFields []filterField, keyword string, db *aorm.DB, context *core.Context) *aorm.DB {
	var (
		joinConditionsMap  = map[string][]string{}
		conditions         []string
		keywords           []interface{}
		generateConditions func(field filterField, scope *aorm.Scope)
	)

	generateConditions = func(filterfield filterField, scope *aorm.Scope) {
		field := filterfield.Field.Struct

		apply := func(kw interface{}) {
			query, arg := filterfield.Apply(kw)
			conditions = append(conditions, query)
			keywords = append(keywords, arg)
		}

		appendString := func() {
			apply(keyword)
		}

		appendInteger := func() {
			if _, err := strconv.Atoi(keyword); err == nil {
				apply(keyword)
			}
		}

		appendFloat := func() {
			if _, err := strconv.ParseFloat(keyword, 64); err == nil {
				apply(keyword)
			}
		}

		appendBool := func() {
			if value, err := strconv.ParseBool(keyword); err == nil {
				apply(value)
			}
		}

		appendTime := func() {
			if parsedTime, err := utils.ParseTime(keyword, context); err == nil {
				apply(parsedTime)
			}
		}

		appendStruct := func() {
			v := reflect.New(field.Struct.Type).Elem()
			switch v.Interface().(type) {
			case time.Time, *time.Time:
				appendTime()
			case sql.NullInt64:
				appendInteger()
			case sql.NullFloat64:
				appendFloat()
			case sql.NullString:
				appendString()
			case sql.NullBool:
				appendBool()
			default:
				// if we don't recognize the struct type, just ignore it
			}
		}

		switch field.Struct.Type.Kind() {
		case reflect.String:
			appendString()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			appendInteger()
		case reflect.Float32, reflect.Float64:
			appendFloat()
		case reflect.Bool:
			appendBool()
		case reflect.Struct, reflect.Ptr:
			appendStruct()
		default:
			apply(keyword)
		}
	}

	scope := db.NewScope(res.Value)
	for _, field := range filterFields {
		generateConditions(field, scope)
	}

	// join conditions
	if len(joinConditionsMap) > 0 {
		var joinConditions []string
		for key, values := range joinConditionsMap {
			joinConditions = append(joinConditions, fmt.Sprintf("%v %v", key, strings.Join(values, " AND ")))
		}
		db = db.Joins(strings.Join(joinConditions, " "))
	}

	// search conditions
	if len(conditions) > 0 {
		return db.Where(aorm.IQ(strings.Join(conditions, " OR ")), keywords...)
	}

	return db
}

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
	return is.set.Has(names...)
}

func (is *ImmutableScopes) List() []string {
	var items []string

	for _, v := range is.set.List() {
		items = append(items, v.(string))
	}

	return items
}

// Searcher is used to search results
type Searcher struct {
	*Context
	scopes        []*Scope
	filters       map[*Filter]*resource.MetaValues
	Pagination    Pagination
	CurrentScopes ImmutableScopes
	Layout        string
	Fragment      *Fragment
}

func (s *Searcher) DefaulLayout(layout ...string) {
	if s.Layout == "" {
		if len(layout) == 0 || layout[0] == "" {
			layout = []string{s.Type.S()}
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
		for _, scope := range s.Resource.scopes {
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

// Filter filter with defined filters, filter with columns value
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

	if s.Fragment != nil {
		context.SetDB(s.Fragment.Filter(context.GetDB()))
	}

	return s.Resource.Crud(context).FindManyLayoutOrDefault(s.Layout)
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

	err = s.Resource.Crud(context).SetLayoutOrDefault(s.Layout).FindOne(result)
	return result, err
}

var filterRegexp = regexp.MustCompile(`^filters\[(.*?)\]`)

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

	// call filters
	if s.filters != nil {
		for filter, value := range s.filters {
			if filter.Handler != nil {
				filterArgument := &FilterArgument{
					Value:    value,
					Context:  context,
					Resource: s.Resource,
				}
				db = filter.Handler(db, filterArgument)
			}
		}
	}

	// add order by
	if orderBy := context.Request.Form.Get("order_by"); orderBy != "" {
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
	var keyword string
	if keyword = context.Request.Form.Get("keyword"); keyword == "" {
		keyword = context.Request.URL.Query().Get("keyword")
	}

	if s.Scheme.SearchHandler != nil {
		context.DB = s.Scheme.SearchHandler(keyword, context)
		return context
	}

	return context
}

func (s *Searcher) FilterRaw(data map[string]string) *Searcher {
	params := url.Values{}
	for key, value := range data {
		params.Add("filters["+key+"].Value", value)
	}

	return s.FilterFromParams(params, nil)
}

func (s *Searcher) FilterFromParams(params url.Values, form *multipart.Form) *Searcher {
	searcher := s

	for key := range params {
		if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
			var prefix = fmt.Sprintf("filters[%v].", matches[1])
			if filter, ok := s.Resource.filters[matches[1]]; ok {
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
		context  = s.Resource.ApplyDefaultFilters(searcher.Context.Context)
	)

	if context != nil && context.Request != nil {
		// parse scopes
		scopes := context.Request.Form["scopes"]
		searcher = searcher.Scope(scopes...)
		searcher = searcher.FilterFromParams(context.Request.Form, context.Request.MultipartForm)
	}

	searcher.callScopes(context)

	db := context.DB

	// pagination
	context.DB = db.Model(s.Resource.Value).Set("qor:getting_total_count", true)
	if err := s.Resource.Crud(context).SetLayoutOrDefault(s.Layout).FindMany(&s.Pagination.Total); err != nil {
		context.AddError(err)
		return context
	}

	if s.Pagination.CurrentPage == 0 {
		if s.Context.Request != nil {
			if page, err := strconv.Atoi(s.Context.Request.Form.Get("page")); err == nil {
				s.Pagination.CurrentPage = page
			}
		}

		if s.Pagination.CurrentPage == 0 {
			s.Pagination.CurrentPage = 1
		}
	}

	if s.Pagination.PerPage == 0 {
		if perPage, err := strconv.Atoi(s.Context.Request.Form.Get("per_page")); err == nil {
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
	FieldName string
	Operation string
}

func filterResourceByFields(res *Resource, filterFields []filterField, keyword string, db *aorm.DB, context *core.Context) *aorm.DB {
	if keyword != "" {
		var (
			joinConditionsMap  = map[string][]string{}
			conditions         []string
			keywords           []interface{}
			generateConditions func(field filterField, scope *aorm.Scope)
		)

		generateConditions = func(filterfield filterField, scope *aorm.Scope) {
			column := filterfield.FieldName
			currentScope, nextScope := scope, scope

			if strings.Contains(column, ".") {
				for _, field := range strings.Split(column, ".") {
					column = field
					currentScope = nextScope
					if field, ok := currentScope.FieldByName(field); ok {
						if relationship := field.Relationship; relationship != nil {
							nextScope = currentScope.New(reflect.New(field.Field.Type()).Interface())
							if relationship.Kind == "many_to_many" {
								var (
									condition string
									jointable = scope.Quote(relationship.JoinTableHandler.Table(scope.DB()))
									key       = fmt.Sprintf("LEFT JOIN %v ON", jointable)
								)

								conditions := []string{}
								for index := range relationship.ForeignDBNames {
									conditions = append(conditions,
										fmt.Sprintf("%v.%v = %v.%v",
											currentScope.QuotedTableName(), scope.Quote(relationship.ForeignFieldNames[index]),
											jointable, scope.Quote(relationship.ForeignDBNames[index]),
										))
								}
								condition = strings.Join(conditions, " AND ")

								conditions = []string{}
								for index := range relationship.AssociationForeignDBNames {
									conditions = append(conditions,
										fmt.Sprintf("%v.%v = %v.%v",
											nextScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignFieldNames[index]),
											jointable, scope.Quote(relationship.AssociationForeignDBNames[index]),
										))
								}

								joinConditionsMap[key] = []string{fmt.Sprintf("%v LEFT JOIN %v ON %v", condition, nextScope.QuotedTableName(), strings.Join(conditions, " AND "))}
							} else {
								key := fmt.Sprintf("LEFT JOIN %v ON", nextScope.QuotedTableName())

								for index := range relationship.ForeignDBNames {
									if relationship.Kind == "has_one" || relationship.Kind == "has_many" {
										joinConditionsMap[key] = append(joinConditionsMap[key],
											fmt.Sprintf("%v.%v = %v.%v",
												nextScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
												currentScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
											))
									} else if relationship.Kind == "belongs_to" {
										joinConditionsMap[key] = append(joinConditionsMap[key],
											fmt.Sprintf("%v.%v = %v.%v",
												currentScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
												nextScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
											))
									}
								}
							}
						}
					}
				}
			}
			tableName := currentScope.QuotedTableName()

			appendString := func(field *aorm.Field) {
				switch filterfield.Operation {
				case "equal":
					conditions = append(conditions, fmt.Sprintf("upper(%v.%v) = upper(?)", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, keyword)
				default:
					conditions = append(conditions, fmt.Sprintf("upper(%v.%v) like upper(?)", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, "%"+keyword+"%")
				}
			}

			appendInteger := func(field *aorm.Field) {
				if _, err := strconv.Atoi(keyword); err == nil {
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, keyword)
				}
			}

			appendFloat := func(field *aorm.Field) {
				if _, err := strconv.ParseFloat(keyword, 64); err == nil {
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, keyword)
				}
			}

			appendBool := func(field *aorm.Field) {
				if value, err := strconv.ParseBool(keyword); err == nil {
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, value)
				}
			}

			appendTime := func(field *aorm.Field) {
				if parsedTime, err := utils.ParseTime(keyword, context); err == nil {
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, parsedTime)
				}
			}

			appendStruct := func(field *aorm.Field) {
				switch field.Field.Interface().(type) {
				case time.Time, *time.Time:
					appendTime(field)
					// add support for sql null fields
				case sql.NullInt64:
					appendInteger(field)
				case sql.NullFloat64:
					appendFloat(field)
				case sql.NullString:
					appendString(field)
				case sql.NullBool:
					appendBool(field)
				default:
					// if we don't recognize the struct type, just ignore it
				}
			}

			if field, ok := currentScope.FieldByName(column); ok {
				if field.IsNormal {
					switch field.Field.Kind() {
					case reflect.String:
						appendString(field)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						appendInteger(field)
					case reflect.Float32, reflect.Float64:
						appendFloat(field)
					case reflect.Bool:
						appendBool(field)
					case reflect.Struct, reflect.Ptr:
						appendStruct(field)
					default:
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				} else if relationship := field.Relationship; relationship != nil {
					switch relationship.Kind {
					case "select_one", "select_many":
						for _, foreignFieldName := range relationship.ForeignFieldNames {
							generateConditions(filterField{
								FieldName: strings.Join([]string{field.Name, foreignFieldName}, "."),
								Operation: filterfield.Operation,
							}, currentScope)
						}
					case "belongs_to":
						for _, foreignFieldName := range relationship.ForeignFieldNames {
							generateConditions(filterField{
								FieldName: foreignFieldName,
								Operation: filterfield.Operation,
							}, currentScope)
						}
					case "many_to_many":
						for _, foreignFieldName := range relationship.ForeignFieldNames {
							generateConditions(filterField{
								FieldName: strings.Join([]string{field.Name, foreignFieldName}, "."),
								Operation: filterfield.Operation,
							}, currentScope)
						}
					}
				}
			} else {
				// context.AddError(fmt.Errorf("filter `%v` is not supported", column))
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
			return db.Where(strings.Join(conditions, " OR "), keywords...)
		}
	}

	return db
}

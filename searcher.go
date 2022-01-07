package admin

import (
	"strconv"
	"strings"

	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

type SearchTermHandler = func(searcher *Searcher, db *aorm.DB, keyword string) (_ *aorm.DB, err error)

// Searcher is used to search results
type Searcher struct {
	*Context
	scopes        []*Scope
	filters       map[*Filter]*FilterArgument
	Filters       map[uintptr]*FilterArgument
	Pagination    Pagination
	Aggregations  CountAggregationScopes
	CurrentScopes ImmutableScopes
	Finder        *Finder
	Orders        []interface{}
	one           bool
	Keyword       string
}

func (this Searcher) clone() *Searcher {
	this.Context = this.Context.Clone()
	return &this
}

// Page set current page, if current page equal -1, then show all records
func (this *Searcher) Page(num int) *Searcher {
	this.Pagination.CurrentPage = num
	return this
}

// PerPage set pre page count
func (this *Searcher) PerPage(num int) *Searcher {
	this.Pagination.PerPage = num
	return this
}

func (this *Searcher) ParseContext(finder ...*Finder) (err error) {
	this.Context.Context = this.Context.Context.Clone()
	ctx := this.Context.Context

	var callNamedSearcherHandlers = func(registrators NamedSearcherHandlersRegistrator) {
		oldContext := this.Context.Context
		this.Context.Context = ctx
		CallNamedSearcherHandlers(this, registrators)
		this.Context.Context = oldContext
	}

	for _, this.Finder = range finder {
	}
	if this.Finder == nil {
		this.Finder = &Finder{}
	}

	if this.Scheme == nil {
		this.Scheme = this.Resource.Scheme
	}

	if len(this.Orders) == 0 {
		this.Orders = this.Scheme.CurrentOrders()
	}

	callNamedSearcherHandlers(this.Scheme.PrepareSearchHandlers)

	if ctx, err = this.Scheme.ApplyDefaultFilters(this.Context); err != nil {
		return
	}

	if !this.Finder.RequestParserDisabled && ctx != nil && ctx.ResourceID == nil && ctx.Request != nil {
		var query = ctx.Request.URL.Query()
		// parse scopes
		if scopes, ok := query["scope[]"]; ok {
			this.Scope(scopes...)
		}
		this.FilterFromParams(query, ctx.Request.MultipartForm)
		this.DefaultFilters()

		if savingName := query.Get("filter_saving_name"); savingName != "" {
			var filters []SavedFilter
			requestURL := ctx.Request.URL
			requestURLQuery := ctx.Request.URL.Query()
			requestURLQuery.Del("filter_saving_name")
			requestURL.RawQuery = requestURLQuery.Encode()
			newFilters := []SavedFilter{{Name: savingName, URL: requestURL.String()}}
			if ctx.AddError(this.Admin.settings.Get("saved_filters", &filters, this.Context)); !ctx.HasError() {
				for _, filter := range filters {
					if filter.Name != savingName {
						newFilters = append(newFilters, filter)
					}
				}

				if err = this.Admin.settings.Save("saved_filters", newFilters, this.Resource, ctx.CurrentUser(), this.Context); err != nil {
					return
				}
			}
		}

		if savingName := query.Get("delete_saved_filter"); savingName != "" {
			var filters, newFilters []SavedFilter
			if ctx.AddError(this.Admin.settings.Get("saved_filters", &filters, this.Context)); !ctx.HasError() {
				for _, filter := range filters {
					if filter.Name != savingName {
						newFilters = append(newFilters, filter)
					}
				}

				if err = this.Admin.settings.Save("saved_filters", newFilters, this.Resource, ctx.CurrentUser(), this.Context); err != nil {
					return
				}
			}
		}
	}

	var db = ctx.DB()

	if this.Resource.IsSingleton() || ctx.ResourceID != nil || this.one {
		db = db.Limit(1)
	} else {
		if db, err = this.callFilters(db, ctx); err != nil {
			return
		}

		// call search
		if !this.Finder.RequestParserDisabled {
			if keyword := ctx.GetFormOrQuery("id"); keyword != "" {
				var ID aorm.ID
				if ID, err = this.Resource.ParseID(keyword); err != nil {
					return
				}
				db = db.Where(ID)
			} else if keyword := ctx.GetFormOrQuery("keyword"); keyword != "" {
				this.Keyword = keyword
				var ok bool
				if keyword[0] == '#' {
					var ID aorm.ID
					if ID, err = this.Resource.ParseID(keyword[1:]); err == nil {
						db = db.Where(ID)
						ok = true
					}
				}
				if !ok {
					if sh := this.Scheme.CurrentSearchHandler(); sh != nil {
						if db, err = sh(this, db, keyword); err != nil {
							return
						}
					}
				}
			}
		}

		if !this.Finder.Unlimited {
			// pagination
			this.Pagination.UnlimitedEnabled = this.Resource.Config.UnlimitedPageCount
			ctx.SetRawDB(db.Model(this.Resource.Value))
			this.Aggregations = CountAggregationScopes{}
			callNamedSearcherHandlers(this.Scheme.SearchCountHandlers)
			callNamedSearcherHandlers(this.Scheme.CountAggregationsHandlers)

			if this.Finder.Count == nil {
				countAggregations := aorm.NewRecordCountResult(&this.Pagination.Total, aorm.CountAggregations{})
				for i, agg := range this.Aggregations {
					for j, aggr := range *agg {
						if aggr.Record == nil {
							aggr.Record = aggr.Resource.New()
						}
						instance := aorm.InstanceOf(aggr.Record)
						for k, agg := range aggr.Aggregations {
							countAggregations.Aggregations[[3]interface{}{i, j, k}] = &aorm.AggregationClause{
								agg.Query,
								agg.QueryFunc,
								aorm.NewFieldScanner(instance.FieldsMap[agg.FieldName]),
								agg.Embed,
							}
						}
					}
				}
				if err = this.Resource.CrudScheme(ctx, this.Scheme).Count(countAggregations); err != nil {
					return
				}
			} else {
				if this.Pagination.Total, err = this.Finder.Count(this); err != nil {
					return err
				}
			}

			if this.Pagination.CurrentPage == 0 {
				if this.Context.Request != nil {
					if page, err := strconv.Atoi(this.Context.Request.URL.Query().Get("page")); err == nil {
						this.Pagination.CurrentPage = page
					}
				}

				if this.Pagination.CurrentPage == 0 {
					this.Pagination.CurrentPage = 1
				}
			}

			if this.Pagination.PerPage == 0 {
				if perPage, err := strconv.Atoi(this.Context.Request.URL.Query().Get("per_page")); err == nil {
					this.Pagination.PerPage = perPage
				} else if this.Resource.Config.PageCount > 0 {
					this.Pagination.PerPage = this.Resource.Config.PageCount
				} else {
					this.Pagination.PerPage = PaginationPageCount
				}
			}

			if this.Pagination.PerPage < 0 {
				if this.Pagination.UnlimitedEnabled {
					this.Pagination.PerPage = -1
				} else if this.Resource.Config.PageCount > 0 {
					this.Pagination.PerPage = this.Resource.Config.PageCount
				} else {
					this.Pagination.PerPage = PaginationPageCount
				}
			}

			if this.Pagination.CurrentPage > 0 {
				this.Pagination.Pages = (this.Pagination.Total-1)/this.Pagination.PerPage + 1
				if this.Finder.Limit == nil {
					db = db.Limit(this.Pagination.PerPage).Offset((this.Pagination.CurrentPage - 1) * this.Pagination.PerPage)
				} else {
					db = this.Finder.Limit(this)
				}
			}

			// exclude
			if exclude := ctx.Request.URL.Query()["exclude"]; len(exclude) > 0 {
				var ids []aorm.ID

				for _, exclude := range exclude {
					if exclude == "" {
						continue
					}
					for _, value := range strings.Split(exclude, ":") {
						if id, err := this.Resource.ParseID(value); err != nil {
							ctx.AddError(err)
						} else {
							ids = append(ids, id)
						}
					}
				}

				ctx.ExcludeResourceID = ids
			}
		}

		if !this.Finder.RequestParserDisabled {
			// add order by
			if orderBy := ctx.GetFormOrQuery("order_by"); orderBy != "" {
				if match := reOrderBy.FindAllStringSubmatch(orderBy, -1); len(match) > 0 {
					fieldPath, order := match[0][1], match[0][4]
					if fpq := this.Context.Resource.ModelStruct.FieldPathQueryOf(fieldPath); fpq != nil {
						if order == "desc" {
							fpq.Sufix(" DESC")
						}
						this.Orders = []interface{}{fpq}
					} else if m := this.Context.Resource.GetDefinedMeta(fieldPath); m != nil {
						if this.Scheme.IsSortableMeta(m.Name) {
							this.Orders = []interface{}{m.Sort(this, aorm.ToOrder(order != "desc"))}
						}
					}
				}
			}
		}

		if len(this.Orders) > 0 {
			db = db.Order(this.Orders, true)
		}
	}

	ctx.SetRawDB(db)
	callNamedSearcherHandlers(this.Scheme.SearchFindHandlers)

	if this.Finder.FindMany == nil {
		this.Finder.FindMany = DefaultFindMany
	}

	if this.Finder.FindOne == nil {
		this.Finder.FindOne = DefaultFindOne
	}

	return
}

var (
	DefaulSearchCrudFactory = func(s *Searcher) *resource.CRUD {
		crud := s.Resource.CrudScheme(s.Context.Context, s.Scheme)
		crud.SetDB(crud.DB().Opt(aorm.OptPreloadTagNames(strings.ToUpper(s.Type.String()))))
		return crud
	}
	DefaultFindMany = func(s *Searcher) (interface{}, error) {
		return DefaulSearchCrudFactory(s).FindManyLayoutOrDefault(s.Layout)
	}
	DefaultFindOne = func(s *Searcher) (interface{}, error) {
		result := s.Resource.NewStruct(s.Site)
		if err := DefaulSearchCrudFactory(s).SetLayoutOrDefault(s.Layout).FindOne(result); err != nil {
			return nil, err
		}
		return result, nil
	}
)

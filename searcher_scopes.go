package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"

	"regexp"

	"gopkg.in/fatih/set.v0"
)

var reOrderBy = regexp.MustCompile(`^(\w+(\.\w+)*)(:(desc))?$`)

type scopeFunc func(db *aorm.DB, context *core.Context) *aorm.DB

type ImmutableScopes struct {
	set set.Interface
}

func (is *ImmutableScopes) Has(names ...interface{}) bool {
	return is.set != nil && is.set.Has(names...)
}

func (is *ImmutableScopes) HasAny(names ...interface{}) bool {
	if is.set == nil {
		return false
	}
	for _, name := range names {
		if is.set.Has(name) {
			return true
		}
	}
	return false
}

func (is *ImmutableScopes) List() (items []string) {
	if is.set != nil {
		for _, v := range is.set.List() {
			items = append(items, v.(string))
		}
	}
	return
}

// Scope filter with defined scopes
func (this *Searcher) Scope(names ...string) {
	scopesSet := set.New(set.NonThreadSafe)
	for _, scope := range this.scopes {
		scopesSet.Add(scope.Name)
	}

	for _, name := range names {
		for _, scope := range this.Scheme.MustGetScopes() {
			if scope.Name == name && !scope.Default && !scopesSet.Has(scope.Name) {
				this.scopes = append(this.scopes, scope)
				scopesSet.Add(name)
			}
		}
	}
	this.CurrentScopes.set = scopesSet
}

func (this *Searcher) GetScopes(advanced bool) (res []string) {
	for _, s := range this.scopes {
		if s.Advanced(this.Context) == advanced {
			res = append(res, s.Name)
		}
	}
	return
}

func (this *Searcher) HasScopes(advanced bool) bool {
	for _, s := range this.scopes {
		if s.Advanced(this.Context) == advanced {
			return true
		}
	}
	return false
}

func (this *Searcher) CountScopes(advanced bool) (i int) {
	for _, s := range this.scopes {
		if s.Advanced(this.Context) == advanced {
			i++
		}
	}
	return
}

func (this *Searcher) callFilters(db *aorm.DB, context *core.Context) (_ *aorm.DB, err error) {
	// call default scopes
	scopes := this.Scheme.MustGetScopes()
defaults:
	for _, scope := range scopes {
		if scope.Default {
			if scope.Group != "" {
				for _, s := range this.scopes {
					if s.Group == scope.Group {
						continue defaults
					}
				}
			}
			db = scope.Handler(db, this, context)
		}
	}

	// call scopes
	for _, scope := range this.scopes {
		db = scope.Handler(db, this, context)
	}

	this.Filters = map[uintptr]*FilterArgument{}

	// call filter
	if this.filters != nil {
		for filter, filterArgument := range this.filters {
			if filter.Handler != nil {
				filterArgument.Context = context
				if filter.Valuer != nil {
					if filterArgument.GoValue, err = filter.Valuer(filterArgument); err != nil {
						return
					}
				} else if v := filterArgument.Value.Get("Value"); v != nil {
					filterArgument.GoValue = v.FirstStringValue()
				}
				this.Filters[filter.ID] = filterArgument
				db = filter.Handler(db, filterArgument)
			}
		}
	}
	return db, nil
}

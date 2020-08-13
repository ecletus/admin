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
		for _, scope := range this.Scheme.scopes {
			if scope.Name == name && !scope.Default && !scopesSet.Has(scope.Name) {
				this.scopes = append(this.scopes, scope)
				scopesSet.Add(name)
			}
		}
	}
}

func (this *Searcher) callScopes(db *aorm.DB, context *core.Context) (_ *aorm.DB, err error) {
	// call default scopes
defaults:
	for _, scope := range this.Scheme.scopes {
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

	// call filtersByName
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
				db = filter.Handler(db, filterArgument)
			}
		}
	}
	return db, nil
}

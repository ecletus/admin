package admin

import (
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/core"
)

// Scope register scope for qor resource
func (res *Resource) Scope(scope *Scope) {
	if scope.Label == "" {
		scope.Label = scope.Name
	}
	res.scopes = append(res.scopes, scope)
}

// Scope scope definiation
type Scope struct {
	Name    string
	Label   string
	Group   string
	Visible func(context *Context) bool
	Handler func(*aorm.DB, *Searcher, *core.Context) *aorm.DB
	Default bool
}

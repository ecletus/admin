package admin

import (
	"github.com/aghape/core"
	"github.com/moisespsena-go/aorm"
)

// Scope register scope for qor resource
func (s *Scheme) Scope(scope *Scope) {
	if scope.Label == "" {
		scope.Label = scope.Name
	}
	s.scopes = append(s.scopes, scope)
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

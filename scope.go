package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

// Scope register scope for qor resource
func (this *Scheme) Scope(scope *Scope) {
	if scope.Label == "" {
		scope.Label = scope.Name
	}
	this.scopes = append(this.scopes, scope)
}

// ScopeGroup register scopes into group for resource
func (this *Scheme) ScopeGroup(groupName string, scope ...*Scope) {
	for _, scope := range scope {
		scope.Group = groupName
		this.Scope(scope)
	}
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

func NewScope(name string, label string, handler func(*aorm.DB, *Searcher, *core.Context) *aorm.DB, defaul ...bool) *Scope {
	var d bool
	for _, d = range defaul {
	}
	return &Scope{Name: name, Label: label, Handler: handler, Default: d}
}

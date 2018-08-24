package admin

import (
	"github.com/aghape/core/resource"
	"github.com/aghape/roles"
)

type Layout struct {
	resource.Layout
	Parent        *Layout
	Metas         []string
	MetaNames     []*resource.MetaName
	MetasFunc     func(res *Resource, context *Context, recorde interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName)
	MetaNamesFunc func(res *Resource, context *Context, recorde interface{}, roles ...roles.PermissionMode) []string
	MetaAliases   map[string]*resource.MetaName
}

func (l *Layout) MetaNameDiscovery(key string) *resource.MetaName {
	for l != nil {
		if l.MetaAliases != nil {
			if name, ok := l.MetaAliases[key]; ok {
				return name
			}
		}
		l = l.Parent
	}
	return nil
}

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

func (l *Layout) SetMetaNames(names ...interface{}) {
	var mnames []*resource.MetaName
	for _, name := range names {
		switch nt := name.(type) {
		case *resource.MetaName:
			mnames = append(mnames, nt)
		case []*resource.MetaName:
			mnames = append(mnames, nt...)
		case []string:
			for _, nameString := range nt {
				mnames = append(mnames, &resource.MetaName{nameString, nameString})
			}
		case string:
			mnames = append(mnames, &resource.MetaName{nt, nt})
		}
	}
	l.MetaNames = mnames
}

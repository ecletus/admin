package admin

import (
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/aghape/roles"
	"github.com/moisespsena-go/aorm"
)

type Layout struct {
	*resource.Layout
	Resource         *Resource
	Parent           *Layout
	Metas            []string
	MetaNames        []*resource.MetaName
	MetasFunc        func(res *Resource, context *Context, recorde interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName)
	MetaNamesFunc    func(res *Resource, context *Context, recorde interface{}, roles ...roles.PermissionMode) []string
	MetaAliases      map[string]*resource.MetaName
	NotIndexRenderID bool
	MetaID           string
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

func (l *Layout) SetMetaNames(names ...interface{}) *Layout {
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
	res := l.Resource
	ms := res.FakeScope.GetModelStruct()
	var columns []interface{}

	for _, f := range ms.PrimaryFields {
		columns = append(columns, aorm.IQ("{}."+f.DBName))
	}

	for _, metaName := range l.MetaNames {
		meta := res.GetMeta(metaName.Name)
		if meta.DB != nil {
			dbName := &aorm.Alias{Expr: "(" + meta.DB.Expr + ")"}
			if meta.DB.Name != "" {
				dbName.Name = utils.ToParamString(meta.DB.Name)
			} else if meta.FieldName != "" {
				dbName.Name = utils.ToParamString(meta.FieldName)
			} else {
				dbName.Name = utils.ToParamString(meta.Name)
			}
			columns = append(columns, dbName)
		} else if meta.FieldName != "" {
			dbName := ms.StructFieldsByName[meta.FieldName].DBName
			columns = append(columns, aorm.IQ("{}."+dbName))
		}
	}
	l.Select(columns...)
	return l
}

func (l *Layout) Prepare(crud *resource.CRUD) *resource.CRUD {
	return l.Layout.Prepare(crud)
}

func (res *Resource) Layout(name string, layout resource.LayoutInterface) {
	if res.registered {
		if l, ok := layout.(*Layout); ok {
			l.Resource = res
			l.SetMetaNames(l.MetaNames)
		}
	}
	res.Resource.Layout(name, layout)
}

func (res *Resource) initializeLayouts() {
	for _, l := range res.Layouts {
		if l, ok := l.(*Layout); ok {
			l.Resource = res
			if l.MetaNames != nil {
				l.SetMetaNames(l.MetaNames)
			}
		}
	}
}

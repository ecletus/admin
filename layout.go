package admin

import (
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
	"github.com/go-aorm/aorm"
	"github.com/pkg/errors"
)

type Layout struct {
	*resource.Layout
	SectionsLayout   string
	SectionsProvider string

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

func (l *Layout) GetSectionsProvider(ctx *Context) *SchemeSectionsProvider {
	if l.SectionsProvider != "" {
		if l.SectionsLayout != "" {
			old := ctx.SectionLayout
			defer func() {
				ctx.SectionLayout = old
			}()
			ctx.SectionLayout = l.SectionsLayout
		}
	}
	return ctx.Scheme.GetSectionsProvider(ctx, ctx.Type, l.SectionsProvider)
}

func (l *Layout) GetSections(res *Resource, context *Context, recorde interface{}, f *SectionsFilter) Sections {
	var (
		names []string
		f2    *SectionsFilter
	)
	if l.MetaNamesFunc != nil {
		names = l.MetaNamesFunc(res, context, recorde)
	}
	if names == nil {
		names = l.Metas
	}
	f2 = new(SectionsFilter).SetInclude(names...)
	if f != nil && len(f.Exclude) > 0 {
		for name := range f.Exclude {
			delete(f2.Include, name)
		}
	}
	return l.GetSectionsProvider(context).
		MustContextSections(&SectionsContext{Ctx: context, Record: recorde}).
		Filter(f2)
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

	res := l.Resource
	var columns []interface{}

	for _, f := range res.ModelStruct.PrimaryFields {
		columns = append(columns, aorm.IQ("{}."+f.DBName))
	}

	if mnames != nil && len(mnames) != len(l.Metas) {
		panic(errors.New("names len isn't equals to defined metas"))
	}

	l.MetaNames = mnames

	for _, metaName := range l.MetaNames {
		meta := res.GetMeta(metaName.Name)
		if meta.FieldStruct != nil && meta.FieldStruct.IsPrimaryKey {
			continue
		}
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
			dbName := res.ModelStruct.FieldsByName[meta.FieldName].DBName
			columns = append(columns, aorm.IQ("{}."+dbName))
		}
	}
	l.Select(columns...)
	return l
}

func (l *Layout) Prepare(crud *resource.CRUD) *resource.CRUD {
	return l.Layout.Prepare(crud)
}

func (this *Resource) Layout(name string, layout resource.LayoutInterface) {
	if this.initialized {
		if l, ok := layout.(*Layout); ok {
			l.Resource = this
			l.SetMetaNames(l.MetaNames)
		}
	}
	this.Resource.Layout(name, layout)
}

func (this *Resource) initializeLayouts() {
	for _, l := range this.Layouts {
		if l, ok := l.(*Layout); ok {
			l.Resource = this
			if l.MetaNames != nil {
				l.SetMetaNames(l.MetaNames)
			}
		}
	}
}

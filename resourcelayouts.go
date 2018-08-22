package admin

import (
	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/roles"
)

func configureDefaultLayouts(res *Resource) {
	defaultLayout := &Layout{
		Layout: resource.Layout{res.Value, res.CallFindManyHandler, res.CallFindOneHandler, nil},
		MetasFunc: func(res *Resource, context *Context, record interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
			metas = res.ConvertSectionToMetas(res.allowedSections(record, res.IndexAttrs(), context, roles...))
			return
		}}
	res.Layout(DEFAULT_LAYOUT, defaultLayout)
	res.Layout("index", &Layout{
		Layout: defaultLayout.Layout,
		MetasFunc: func(res *Resource, context *Context, record interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
			metas = res.ConvertSectionToMetas(res.allowedSections(record, res.IndexAttrs(), context, roles...))
			return
		}})
	res.Layout("show", &Layout{
		Layout: defaultLayout.Layout,
		MetasFunc: func(res *Resource, context *Context, record interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
			metas = res.ConvertSectionToMetas(res.allowedSections(record, res.ShowAttrs(), context, roles...))
			return
		}})
	res.Layout("edit", &Layout{
		Layout: defaultLayout.Layout,
		MetasFunc: func(res *Resource, context *Context, record interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
			metas = res.ConvertSectionToMetas(res.allowedSections(record, res.EditAttrs(), context, roles...))
			return
		}})

	configureDefaultBasicLayouts(res, defaultLayout)
}

func configureDefaultBasicLayouts(res *Resource, defaultLayout *Layout) {
	res.SetMeta(&Meta{Name: BASIC_META_ID, Valuer: func(r interface{}, context *core.Context) interface{} {
		if b, ok := r.(resource.BasicValue); ok {
			return b.BasicID()
		}
		if b, ok := r.(interface {
			GetID() string
		}); ok {
			return b.GetID()
		}
		if b, ok := r.(interface {
			GetID() int64
		}); ok {
			return b.GetID()
		}
		return nil
	}})

	res.SetMeta(&Meta{Name: BASIC_META_LABEL, Valuer: func(r interface{}, context *core.Context) interface{} {
		var label string
		if b, ok := r.(interface {
			BasicLabel() string
		}); ok {
			label = b.BasicLabel()
		} else if b, ok := r.(interface {
			Stringify() string
		}); ok {
			label = b.Stringify()
		}
		return label
	}})

	res.SetMeta(&Meta{Name: BASIC_META_ICON, Valuer: func(r interface{}, context *core.Context) interface{} {
		var icon string
		if b, ok := r.(interface {
			BasicIcon() string
		}); ok {
			icon = b.BasicIcon()
		} else if b, ok := r.(interface {
			GetIcon() string
		}); ok {
			icon = b.GetIcon()
		}
		return icon
	}})

	res.SetMeta(&Meta{Name: BASIC_META_HTML, Valuer: func(r interface{}, context *core.Context) interface{} {
		html := context.Htmlify(r)
		return html
	}})

	res.MetaAliases[BASIC_META_ID] = &resource.MetaName{Name: "ID"}
	res.MetaAliases[BASIC_META_LABEL] = &resource.MetaName{Name: "Text"}
	res.MetaAliases[BASIC_META_HTML] = &resource.MetaName{Name: "HTML"}
	res.MetaAliases[BASIC_META_ICON] = &resource.MetaName{Name: "Icon"}

	var (
		metaNames             = []string{BASIC_META_ID, BASIC_META_LABEL}
		metaNamesWithIcon     = append(metaNames, BASIC_META_ICON)
		metaHTMLNames         = []string{BASIC_META_ID, BASIC_META_HTML}
		metaHTMLNamesWithIcon = append(metaHTMLNames, BASIC_META_ICON)
	)

	res.Layout(BASIC_LAYOUT, &Layout{
		Layout: defaultLayout.Layout,
		Metas:  metaNames,
	})

	res.Layout(BASIC_LAYOUT_WITH_ICON, &Layout{
		Layout: defaultLayout.Layout,
		Metas:  metaNamesWithIcon,
	})

	res.Layout(BASIC_LAYOUT_HTML, &Layout{
		Layout: defaultLayout.Layout,
		Metas:  metaHTMLNames,
	})

	res.Layout(BASIC_LAYOUT_HTML_WITH_ICON, &Layout{
		Layout: defaultLayout.Layout,
		Metas:  metaHTMLNamesWithIcon,
	})
}

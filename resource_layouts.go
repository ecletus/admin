package admin

import (
	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/aghape/roles"
)

func configureDefaultLayouts(res *Resource) {
	defaultLayout := &Layout{
		Layout: &resource.Layout{StructValue: resource.NewStructValue(res.Value)},
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
		return res.GetKey(r)
	}})

	res.SetMeta(&Meta{Name: BASIC_META_LABEL, Valuer: func(r interface{}, context *core.Context) interface{} {
		if b, ok := r.(resource.BasicValue); ok {
			return b.BasicLabel()
		}
		return utils.Stringify(r)
	}})

	res.SetMeta(&Meta{Name: BASIC_META_ICON, Valuer: func(r interface{}, context *core.Context) interface{} {
		switch rt := r.(type) {
		case resource.IconGetter:
			return rt.GetIcon()
		case resource.IconContextGetter:
			return rt.GetIcon(context)
		case resource.BasicValue:
			return rt.BasicIcon()
		default:
			return ""
		}
	}})

	res.SetMeta(&Meta{Name: BASIC_META_HTML, Valuer: func(r interface{}, context *core.Context) interface{} {
		html := context.Htmlify(r)
		return html
	}})

	res.MetaAliases[BASIC_META_ID] = &resource.MetaName{Name: "ID"}
	res.MetaAliases[BASIC_META_LABEL] = &resource.MetaName{Name: "Title"}
	res.MetaAliases[BASIC_META_HTML] = &resource.MetaName{Name: "Title"}
	res.MetaAliases[BASIC_META_ICON] = &resource.MetaName{Name: "Icon"}

	var (
		metaNames             = []string{BASIC_META_ID, BASIC_META_LABEL}
		metaNamesWithIcon     = append(metaNames, BASIC_META_ICON)
		metaHTMLNames         = []string{BASIC_META_ID, BASIC_META_HTML}
		metaHTMLNamesWithIcon = append(metaHTMLNames, BASIC_META_ICON)
	)

	basicLayout := resource.NewBasicLayout()

	res.Layout(BASIC_LAYOUT, &Layout{
		Layout: basicLayout,
		Metas:  metaNames,
	})

	res.Layout(BASIC_LAYOUT_WITH_ICON, &Layout{
		Layout: basicLayout,
		Metas:  metaNamesWithIcon,
	})

	res.Layout(BASIC_LAYOUT_HTML, &Layout{
		Layout: basicLayout,
		Metas:  metaHTMLNames,
	})

	res.Layout(BASIC_LAYOUT_HTML_WITH_ICON, &Layout{
		Layout: basicLayout,
		Metas:  metaHTMLNamesWithIcon,
	})
}

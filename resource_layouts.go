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
	res.SetMeta(&Meta{
		Name:  BASIC_META_ID,
		Label: I18NGROUP + ".basic_metas.ID",
		Valuer: func(r interface{}, context *core.Context) interface{} {
			return res.GetKey(r)
		}})

	res.SetMeta(&Meta{
		Name:  BASIC_META_LABEL,
		Label: I18NGROUP + ".basic_metas.Label",
		Valuer: func(r interface{}, context *core.Context) interface{} {
			if b, ok := r.(resource.BasicLabel); ok {
				return b.BasicLabel()
			}
			return utils.StringifyContext(r, context)
		}})

	res.SetMeta(&Meta{
		Name:  BASIC_META_ICON,
		Label: I18NGROUP + ".basic_metas.Icon",
		Type:  "image_url",
		Valuer: func(r interface{}, context *core.Context) interface{} {
			switch rt := r.(type) {
			case resource.IconGetter:
				return rt.GetIcon()
			case resource.IconContextGetter:
				return rt.GetIcon(context)
			case resource.BasicIcon:
				return rt.BasicIcon()
			default:
				return ""
			}
		}})

	res.SetMeta(&Meta{
		Name:  BASIC_META_HTML,
		Label: I18NGROUP + ".basic_metas.Label",
		Valuer: func(r interface{}, ctx *core.Context) interface{} {
			html := utils.HtmlifyContext(r, ctx)
			if html == "" {
				value := res.GetMeta(BASIC_META_LABEL).Value(ctx, r)
				html = utils.HtmlifyContext(value, ctx)
			}
			return html
		}})

	res.MetaAliases[BASIC_META_ID] = &resource.MetaName{Name: "ID"}
	res.MetaAliases[BASIC_META_LABEL] = &resource.MetaName{Name: "Value"}
	res.MetaAliases[BASIC_META_HTML] = &resource.MetaName{Name: "Value"}
	res.MetaAliases[BASIC_META_ICON] = &resource.MetaName{Name: "Icon"}

	var (
		metaNames             = []string{BASIC_META_ID, BASIC_META_LABEL}
		metaNamesWithIcon     = append([]string{BASIC_META_ICON}, metaNames...)
		metaHTMLNames         = []string{BASIC_META_ID, BASIC_META_HTML}
		metaHTMLNamesWithIcon = append([]string{BASIC_META_ICON}, metaHTMLNames...)
	)

	basicLayout := resource.NewBasicLayout()

	res.Layout(BASIC_LAYOUT, &Layout{
		Layout:           basicLayout,
		Metas:            metaNames,
		NotIndexRenderID: true,
	})

	res.Layout(BASIC_LAYOUT_WITH_ICON, &Layout{
		Layout:           basicLayout,
		Metas:            metaNamesWithIcon,
		NotIndexRenderID: true,
		MetaID:           BASIC_META_ID,
	})

	res.Layout(BASIC_LAYOUT_HTML, &Layout{
		Layout:           basicLayout,
		Metas:            metaHTMLNames,
		NotIndexRenderID: true,
		MetaID:           BASIC_META_ID,
	})

	res.Layout(BASIC_LAYOUT_HTML_WITH_ICON, &Layout{
		Layout:           basicLayout,
		Metas:            metaHTMLNamesWithIcon,
		NotIndexRenderID: true,
		MetaID:           BASIC_META_ID,
	})
}

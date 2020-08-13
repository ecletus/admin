package admin

import (
	"fmt"
	"reflect"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"

	"github.com/ecletus/roles"
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
		Name:             BASIC_META_ID,
		DefaultInvisible: true,
		Label:            I18NGROUP + ".basic_metas.ID",
		EncodedName:      "ID",
		Valuer: func(r interface{}, context *core.Context) interface{} {
			return res.GetKey(r)
		}})

	res.SetMeta(&Meta{
		Name:             BASIC_META_LABEL,
		DefaultInvisible: true,
		EncodedName:      "Label",
		Label:            I18NGROUP + ".basic_metas.Label",
		Valuer: func(r interface{}, context *core.Context) interface{} {
			if b, ok := r.(resource.BasicLabel); ok {
				return b.BasicLabel()
			} else if s, ok := r.(fmt.Stringer); ok {
				return s.String()
			} else if meta := res.GetDefinedMeta(META_STRINGIFY); meta != nil {
				return meta.Value(context, r).(string)
			}
			return utils.StringifyContext(r, context)
		}})

	res.SetMeta(&Meta{
		Name:             BASIC_META_ICON,
		DefaultInvisible: true,
		EncodedName:      "Icon",
		Label:            I18NGROUP + ".basic_metas.Icon",
		Type:             "image_url",
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
		Name:             BASIC_META_HTML,
		EncodedName:      "Label",
		Label:            I18NGROUP + ".basic_metas.Label",
		DefaultInvisible: true,
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

	{
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

	{
		var descriptionValuer func(ctx *core.Context, r interface{}) string

		if res.Config.DescriprionGetter != nil {
			descriptionValuer = res.Config.DescriprionGetter
		} else if _, ok := res.Value.(resource.DescriptionGetter); ok {
			descriptionValuer = func(_ *core.Context, r interface{}) string {
				return r.(resource.DescriptionGetter).GetDescription()
			}
		} else {
			for _, fieldName := range DescriptionFields {
				if field, ok := res.ModelStruct.FieldsByName[fieldName]; ok && field.Struct.Type.Kind() == reflect.String {
					descriptionValuer = func(_ *core.Context, r interface{}) string {
						return reflect.Indirect(reflect.ValueOf(r)).FieldByIndex(field.StructIndex).String()
					}
					break
				}
			}
		}

		if descriptionValuer == nil {
			return
		}

		res.DescriptionValuer = descriptionValuer

		var (
			metaNames             = append(metaNames, META_DESCRIPTIFY)
			metaNamesWithIcon     = append(metaNamesWithIcon, META_DESCRIPTIFY)
			metaHTMLNames         = append(metaHTMLNames, META_DESCRIPTIFY)
			metaHTMLNamesWithIcon = append(metaHTMLNamesWithIcon, META_DESCRIPTIFY)
			layout                = resource.NewBasicDescriptionLayout()
		)

		res.SetMeta(&Meta{
			Name:             META_DESCRIPTIFY,
			EncodedName:      "Description",
			Label:            I18NGROUP + ".metas.Descriptify",
			DefaultInvisible: true,
			Valuer: func(r interface{}, ctx *core.Context) interface{} {
				if g, ok := r.(resource.DescriptionGetter); ok {
					return g.GetDescription()
				}
				return res.DescriptionValuer(ctx, r)
			}})

		res.Layout(BASIC_LAYOUT_DESCRIPTION, &Layout{
			Layout:           layout,
			Metas:            metaNames,
			NotIndexRenderID: true,
		})

		res.Layout(BASIC_LAYOUT_DESCRIPTION_WITH_ICON, &Layout{
			Layout:           layout,
			Metas:            metaNamesWithIcon,
			NotIndexRenderID: true,
			MetaID:           BASIC_META_ID,
		})

		res.Layout(BASIC_LAYOUT_HTML_DESCRIPTION, &Layout{
			Layout:           layout,
			Metas:            metaHTMLNames,
			NotIndexRenderID: true,
			MetaID:           BASIC_META_ID,
		})

		res.Layout(BASIC_LAYOUT_HTML_DESCRIPTION_WITH_ICON, &Layout{
			Layout:           layout,
			Metas:            metaHTMLNamesWithIcon,
			NotIndexRenderID: true,
			MetaID:           BASIC_META_ID,
		})
	}
}

package admin

func DisableMetaOnPrint(res *Resource, names ...string) {
	do := func() {
		res.EachMetas(func(m *Meta) {
			m.NewEnabled(func(old MetaEnabled, recorde interface{}, ctx *Context, meta *Meta, readOnly bool) bool {
				if ctx.Type.Has(PRINT) {
					return false
				}
				return old(recorde, ctx, meta, readOnly)
			})
		}, names...)
	}
	if res.Config.NotMount {
		res.PostInitialize(do)
	} else {
		res.PostMetasSetup(do)
	}
}

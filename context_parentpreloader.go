package admin

type CheckLoaderForUpdater interface {
	IsLoadForUpdate(ctx *Context) bool
}

func (this *Context) ParentPreload(flag ParentPreloadFlag) (ctx *Context) {
	if this.Resource.Config.ParentPreload.Has(flag) {
		var (
			parentRes = this.Resource.ParentResource
			parent    = parentRes.New()
			DB        = this.DB().Unscoped()
		)
		this.ParentResourceID[len(this.ParentResourceID)-1].SetTo(parent)
		if err := DB.ModelStruct(parentRes.ModelStruct, parent).First(parent).Error; err != nil {
			this.AddError(err)
			return this
		}
		this.ParentRecord[len(this.ParentRecord)-1] = parent
	}

	ctx = this

	if len(ctx.ParentRecord) > 0 {
		prs := ctx.ParentResource
		for i, p := range ctx.ParentRecord {
			ctx = ctx.CreateChild(prs[i], p)
			if ctx.ResourceID == nil {
				ctx.ResourceID = this.ParentResourceID[i]
			}
		}
		ctx = ctx.CreateChild(this.Resource, this.ResourceRecord)
		ctx.ParentRecord = this.ParentRecord
		ctx.ParentResourceID = this.ParentResourceID
		ctx.ParentResource = this.ParentResource
		ctx.ResourceID = this.ResourceID
	}

	ctx.Type = this.Type
	return
}

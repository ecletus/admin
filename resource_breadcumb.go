package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

type ResourceCrumber struct {
	Prev     core.Breadcrumber
	Resource *Resource
	ParentID []aorm.ID
	Parent   []*Resource
	ID       aorm.ID
}

type ResourceCrumb struct {
	core.Breadcrumb
	Resource *Resource
	ParentID []aorm.ID
	ID       aorm.ID
}

func (this *ResourceCrumber) NewCrumb(ctx *core.Context, recorde bool) core.Breadcrumb {
	uri := this.Resource.GetContextIndexURI(ctx, this.ParentID...)
	crumb := &ResourceCrumb{
		Resource: this.Resource,
		ParentID: this.ParentID,
	}
	if recorde {
		crumb.ID = this.ID
		_ = ctx.WithDB(func(ctx *core.Context) {
			ctx.SetRawDB(ctx.DB().Unscoped())
			model, err := this.Resource.Crud(ctx).FindOneBasic(this.ID)

			if err != nil {
				if aorm.IsRecordNotFoundError(err) {
					ctx.AddError(err)
					crumb = nil
					return
				} else {
					panic(err)
				}
			}
			if !this.Resource.Config.Singleton {
				uri += "/" + this.ID.String()
			}
			crumb.Breadcrumb = core.NewBreadcrumb(uri, model.BasicLabel(), model.BasicIcon())
		})
	} else if this.Resource.Config.Singleton {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, this.Resource.SingularLabelKey())
	} else {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, this.Resource.PluralLabelKey())
	}
	return crumb
}

func (this *ResourceCrumber) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb) {
	crumbs = append(crumbs, this.NewCrumb(ctx, false))
	if !this.Resource.Config.Singleton && this.ID != nil {
		if crumb := this.NewCrumb(ctx, true); crumb != nil {
			crumbs = append(crumbs, crumb)
		}
	}
	return
}

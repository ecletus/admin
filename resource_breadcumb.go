package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
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

type StaticCrumbValuer interface {
	AdminStaticCrumbValue(ctx *core.Context, record bool) resource.BasicValuer
}

func (this *ResourceCrumber) NewCrumb(ctx *core.Context, record bool) core.Breadcrumb {
	uri := this.Resource.GetContextIndexURI(ctx, this.ParentID...)
	crumb := &ResourceCrumb{
		Resource: this.Resource,
		ParentID: this.ParentID,
	}
	if record {
		crumb.ID = this.ID
		_ = ctx.WithDB(func(ctx *core.Context) {
			ctx.SetRawDB(ctx.DB().Unscoped())
			var (
				model resource.BasicValuer
				err   error
			)
			if scv, ok := this.Resource.Value.(StaticCrumbValuer); ok {
				model = scv.AdminStaticCrumbValue(ctx, record)
			} else if model, err = this.Resource.Crud(ctx).FindOneBasic(this.ID); err != nil {
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

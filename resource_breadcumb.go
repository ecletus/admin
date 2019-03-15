package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

type ResourceCrumber struct {
	Prev     core.Breadcrumber
	Resource *Resource
	ParentID []string
	ID       string
}

type ResourceCrumb struct {
	core.Breadcrumb
	Resource *Resource
	ParentID []string
	ID       string
}

func (r *ResourceCrumber) NewCrumb(ctx *core.Context, recorde bool) core.Breadcrumb {
	uri := r.Resource.GetContextIndexURI(ctx, r.ParentID...)
	crumb := &ResourceCrumb{
		Resource: r.Resource,
		ParentID: r.ParentID,
	}
	if recorde {
		crumb.ID = r.ID
		clone := ctx.Clone()

		if HasDeletedUrlQuery(ctx.Request.URL.Query()) {
			clone.SetDB(clone.DB.Unscoped())
		}
		model, err := r.Resource.Crud(clone).FindOneBasic(r.ID)

		if err != nil {
			if aorm.IsRecordNotFoundError(err) {
				return nil
			}
			panic(err)
		}

		if !r.Resource.Config.Singleton {
			uri += "/" + r.ID
		}
		crumb.Breadcrumb = core.NewBreadcrumb(uri, model.BasicLabel(), model.BasicIcon())
	} else if r.Resource.Config.Singleton {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, r.Resource.SingularLabelKey())
	} else {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, r.Resource.PluralLabelKey())
	}
	return crumb
}

func (r *ResourceCrumber) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb) {
	crumbs = append(crumbs, r.NewCrumb(ctx, false))
	if !r.Resource.Config.Singleton && r.ID != "" {
		if crumb := r.NewCrumb(ctx, true); crumb != nil {
			crumbs = append(crumbs, crumb)
		}
	}
	return
}

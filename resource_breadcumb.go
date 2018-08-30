package admin

import (
	"github.com/aghape/core"
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
		uri = uri + "/" + r.ID
		model, err := r.Resource.CrudDB(ctx.DB).FindOneBasic(r.ID)

		if err != nil {
			panic(err)
		}

		uri = uri + "/" + r.ID
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
	if r.ID != "" && !r.Resource.Config.Singleton {
		crumbs = append(crumbs, r.NewCrumb(ctx, true))
	}
	return
}

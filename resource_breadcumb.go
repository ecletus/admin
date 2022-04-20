package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
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
	AdminStaticCrumbValue(ctx *core.Context, res *Resource, id aorm.ID) (v resource.BasicValuer, err error)
}

func (this *ResourceCrumber) NewCrumb(ctx *core.Context, record bool) (_ core.Breadcrumb, err error) {
	uri := this.Resource.GetContextIndexURI(ContextFromContext(ctx), this.ParentID...)
	crumb := &ResourceCrumb{
		Resource: this.Resource,
		ParentID: this.ParentID,
	}
	if record {
		crumb.ID = this.ID
		_ = ctx.WithDB(func(ctx *core.Context) {
			if scv, ok := this.Resource.Value.(StaticCrumbValuer); ok {
				var model resource.BasicValuer
				if model, err = scv.AdminStaticCrumbValue(ctx, this.Resource, this.ID); err != nil {
					return
				}
				crumb.Breadcrumb = core.NewBreadcrumb(uri, model.BasicLabel(), model.BasicIcon())
			} else {
				crumb.Breadcrumb = core.NewBreadcrumb(uri, this.ID.String(), "")
			}
			if !this.Resource.Config.Singleton {
				uri += "/" + this.ID.String()
			}
		})
	} else if this.Resource.Config.Singleton {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, this.Resource.SingularLabelKey())
	} else {
		crumb.Breadcrumb = core.NewBreadcrumb(uri, this.Resource.PluralLabelKey())
	}
	return crumb, nil
}

func (this *ResourceCrumber) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb, _ error) {
	c, err := this.NewCrumb(ctx, false)
	if err != nil {
		return nil, err
	}
	crumbs = append(crumbs, c)
	if !this.Resource.Config.Singleton && this.ID != nil {
		if crumb, err := this.NewCrumb(ctx, true); err != nil {
			return nil, err
		} else if crumb != nil {
			crumbs = append(crumbs, crumb)
		}
	}
	return
}

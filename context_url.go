package admin

import (
	"path"

	"github.com/go-aorm/aorm"
)

// RoutePrefix return route prefix of resource
func (this *Resource) RoutePrefix() string {
	var params string
	for this.ParentResource != nil {
		params = path.Join(this.ParentResource.ToParam(), this.ParentResource.ParamIDPattern(), params)
		this = this.ParentResource
	}
	return params
}

// URLFor generate url for resource value
//     context.URLFor(&Product{})
//     context.URLFor(&Product{ID: 111})
//     context.URLFor(productResource)
func (this *Context) URLFor(value interface{}, resources ...*Resource) string {
	if admin, ok := value.(*Admin); ok {
		return this.Path(admin.Config.MountPath)
	} else if urler, ok := value.(interface {
		URL(*Context) string
	}); ok {
		return urler.URL(this)
	} else if res, ok := value.(*Resource); ok {
		return res.GetContextIndexURI(this)
	} else {
		if len(resources) == 0 {
			return ""
		}

		res := resources[0]
		if res.Config.Singleton {
			return res.GetIndexLink(this, this.ParentResourceID)
		}

		/*
			var (
				ids []aorm.ID
				self = this
				i = len(res.Parents)
			)

			for ;i > 0;i++ {
				if pres := self.Resource; pres == res.Parents[i] {
					ids = append(ids, self.ResourceID)
					self = self.Parent
				}
			}
		*/

		uri := res.GetLink(this, value, this.Context, this.ParentResourceID)
		return uri
	}
	return this.Path("")
}

// URLFor generate url for resource value
//     context.URLFor(&Product{})
//     context.URLFor(&Product{ID: 111})
//     context.URLFor(productResource)
func (this *Context) TopURLFor(value interface{}, resources ...*Resource) string {
	var res *Resource
	if len(resources) == 0 {
		res = this.Resource
	} else {
		res = resources[0]
	}
	res = res.Top()

	if value == nil {
		return this.Path(res.GetIndexURI(this))
	}
	if vs, ok := value.(string); ok {
		if vs == "" {
			return this.Path(res.GetIndexURI(this))
		}
		if id, err := res.ParseID(vs); err != nil {
			return "[[parse id failed: " + err.Error() + "]]"
		} else {
			return this.Path(res.GetURI(this, id))
		}
	} else if vs, ok := value.(aorm.ID); ok {
		if vs == nil {
			return this.Path(res.GetIndexURI(this))
		}
		return this.Path(res.GetURI(this, vs))
	}
	return this.Path(res.URLFor(this, value))
}

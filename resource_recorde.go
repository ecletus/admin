package admin

import (
	"fmt"
	"strings"

	"github.com/ecletus/core/utils"
)

type ResourceRecorde struct {
	Context  *Context
	Resource *Resource
	Recorde  interface{}
	Recordes []interface{}
}

func NewResourceRecorde(context *Context, resource *Resource, recorde ...interface{}) *ResourceRecorde {
	r := &ResourceRecorde{Context: context, Resource: resource}
	if len(recorde) == 1 {
		r.Recorde = recorde[0]
	} else {
		r.Recordes = recorde
	}
	return r
}

func (rr *ResourceRecorde) Count() int {
	if rr.Recorde != nil {
		return 1
	}
	return len(rr.Recordes)
}

func (rr *ResourceRecorde) String() string {
	r := rr.Context.Ts(rr.Resource.SingularLabelKey(), rr.Resource.Name)
	if rr.Resource.Config.Singleton {
		return r
	}
	if rr.Recorde != nil {
		r += " (" + utils.Stringify(rr.Recorde) + ")"
	} else if len(rr.Recordes) > 0 {
		var ss []string
		for i, recorde := range rr.Recordes {
			ss = append(ss, fmt.Sprintf("%d: %s", i+1, utils.Stringify(recorde)))
		}
		r += " {" + strings.Join(ss, ", ") + "}"
	}
	return r
}

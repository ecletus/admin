package admin

import (
	"github.com/aghape/core/utils"

	"github.com/moisespsena-go/xroute"
)

func (res *Resource) RegisterDefaultRouters(actions ...string) {
	res.Controller.AppendDefaultActions(actions...)
	res.Controller.RegisterDefaultRouters()
}

func (res *Resource) InitRoutes() *xroute.Mux {
	if res.Config.Singleton {
		for param, subRes := range res.ResourcesByParam {
			r := subRes.InitRoutes()
			pattern := "/" + param
			res.Router.Mount(pattern, r)
		}
	} else {
		for param, subRes := range res.ResourcesByParam {
			r := subRes.InitRoutes()
			pattern := "/" + param
			res.ObjectRouter.Mount(pattern, r)
		}
		res.Router.Mount("/"+res.ParamIDPattern(), res.ObjectRouter)
	}
	return res.Router
}

func (res *Resource) MountTo(param string) *Resource {
	config := &(*res.Config)
	if config.Sub != nil {
		config.Sub = &(*config.Sub)
	}
	nmp := utils.NamifyString(param)
	config.Name += nmp
	config.Param = param
	config.ID += nmp
	config.NotMount = false
	config.Invisible = true
	return res.admin.AddResource(res.Value, config)
}
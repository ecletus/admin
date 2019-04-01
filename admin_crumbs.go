package admin

import (
	"strings"

	"github.com/ecletus/core"
)

func (admin *Admin) LoadCrumbs(rh *RouteHandler, ctx *Context, patterns ...string) {
	resourceParam := strings.TrimSuffix(patterns[0][1:], "/*")
	res := admin.GetResourceByParam(resourceParam)

	if res == nil {
		return
	}
	patterns = patterns[1:]

	resCrumber := &ResourceCrumber{Resource: res}
	var lastScheme *Scheme
	crubers := []core.Breadcrumber{resCrumber}

	for i, l := 0, len(patterns); i < l; i++ {
		// skip slash
		p := patterns[i][1:]
		if p == "" {
			continue
		}

		if res.Config.Singleton {
			if res.HasKey() && resCrumber.ID == "" {
				if primaryFiels := res.GetPrimaryFields(); len(primaryFiels) > 0 {
					recorde := res.New()
					err := ctx.Site.GetSystemDB().DB.Model(res.Value).Select(primaryFiels[0].DBName).First(recorde).Error
					if err != nil {
						panic(err)
					}
					key := res.GetKey(recorde)
					resCrumber.ID = key
					ctx.RouteContext.URLParams.Add(res.ParamIDName(), key)
				}
			} else {
				for _, m := range *res.menus {
					switch p {
					case m.RelativePath[1:], m.RelativePath[1:] + "/*":
						if subRes := m.Resource; subRes != nil {
							if subRes.Config.Singleton && subRes.HasKey() {
								resCrumber = &ResourceCrumber{Resource: subRes, ParentID: append(resCrumber.ParentID, resCrumber.ID), ID: resCrumber.ID}
							} else {
								resCrumber = &ResourceCrumber{Resource: subRes, ParentID: append(resCrumber.ParentID, resCrumber.ID)}
							}
							res = subRes
							crubers = append(crubers, resCrumber)
						}
					}
				}
			}
		} else {
			// id pattern
			idPattern := res.ParamIDPattern()
			if strings.HasPrefix(p, idPattern) {
				resCrumber.ID = ctx.URLParam(res.ParamIDName())
			} else {
				subRes := res.GetResourceByParam(strings.TrimSuffix(p, "/*"))
				if subRes == nil {
					schemeName := strings.Replace(strings.TrimSuffix(p, ".json"), "/", ".", -1)
					if scheme, ok := res.GetSchemeOk(schemeName); ok {
						crubers = append(crubers, scheme)
						lastScheme = scheme
					}
				} else {
					if subRes.Config.Singleton && subRes.HasKey() {
						resCrumber = &ResourceCrumber{Resource: subRes, ParentID: append(resCrumber.ParentID, resCrumber.ID), ID: resCrumber.ID}
					} else {
						resCrumber = &ResourceCrumber{Resource: subRes, ParentID: append(resCrumber.ParentID, resCrumber.ID)}
					}
					res = subRes
					crubers = append(crubers, resCrumber)
				}
			}
		}
	}

	if resCrumber != nil {
		ctx.setResourceFromCrumber(resCrumber)
		if lastScheme != nil && lastScheme.Resource == resCrumber.Resource {
			ctx.Scheme = lastScheme
		}
	}

	appender := ctx.Breadcrumbs()

	for _, crumber := range crubers {
		appender.Append(crumber.Breadcrumbs(ctx.Context)...)
	}
	return
}

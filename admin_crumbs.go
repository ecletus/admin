package admin

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ecletus/core"
)

func (this *Admin) LoadCrumbs(_ *RouteHandler, ctx *Context, patterns ...string) {
	resourceParam := strings.TrimSuffix(patterns[0][1:], "/*")
	res := this.GetResourceByParam(resourceParam)

	if res == nil {
		return
	}
	patterns = patterns[1:]

	resCrumber := &ResourceCrumber{Resource: res}
	var lastScheme *Scheme
	crubers := []core.Breadcrumber{resCrumber}

	var unescape = func(v string) string {
		if v2, err := url.PathUnescape(v); err != nil {
			return ""
		} else {
			return v2
		}
	}

	for i, l := 0, len(patterns); i < l; i++ {
		// skip slash
		p := patterns[i][1:]
		if p == "" {
			continue
		}

		if res.Config.Singleton {
			if res.HasKey() && resCrumber.ID == nil {
				if primaryFiels := res.GetPrimaryFields(); len(primaryFiels) > 0 {
					recorde := res.New()
					err := ctx.Site.GetSystemDB().DB.Model(res.Value).Select(primaryFiels[0].DBName).First(recorde).Error
					if err != nil {
						panic(err)
					}
					key := res.GetKey(recorde)
					resCrumber.ID = key
					ctx.RouteContext.URLParams.Add(res.ParamIDName(), key.String())
				}
			} else {
				var ok bool
				if res.itemMenus != nil {
				menu_loop:
					for _, m := range res.itemMenus {
						switch p {
						case m.URI:
							if subRes := m.Resource; subRes != nil {
								if subRes.Config.Singleton && subRes.HasKey() {
									resCrumber = &ResourceCrumber{
										Resource: subRes,
										Parent:   append(resCrumber.Parent, resCrumber.Resource),
										ParentID: append(resCrumber.ParentID, resCrumber.ID),
										ID:       resCrumber.ID,
									}
								} else {
									resCrumber = &ResourceCrumber{
										Resource: subRes,
										Parent:   append(resCrumber.Parent, resCrumber.Resource),
										ParentID: append(resCrumber.ParentID, resCrumber.ID),
									}
								}
								res = subRes
								crubers = append(crubers, resCrumber)
								ok = true
								break menu_loop
							}
						}
					}
				}

				if !ok {
					for _, action := range res.Actions {
						if p == action.Name {
							ok = true
							crubers = append(crubers, core.BreadCrumberFunc(func(ctx *core.Context) ([]core.Breadcrumb, error) {
								return []core.Breadcrumb{core.NewBreadcrumb("", ctx.TtS(action), "")}, nil
							}))
							if action.Resource != nil {
								resCrumber = &ResourceCrumber{Resource: action.Resource}
							}
							break
						}
					}
				}
			}
		} else {
			// id pattern
			idPattern := res.ParamIDPattern()
			if strings.HasPrefix(p, idPattern) {
				var err error
				resCrumber.ID, err = res.ParseID(unescape(ctx.URLParam(res.ParamIDName())))
				if err != nil {
					http.Error(ctx.Writer, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				cleanedP := strings.TrimSuffix(p, "/*")
				subRes := res.GetResourceByParam(cleanedP)
				if subRes == nil {
					for param, sub := range res.ResourcesByParam {
						if strings.HasPrefix(cleanedP, param+"/") {
							subRes = sub
							patterns[i] = patterns[i][len(param)+1:]
							patterns = append(patterns[0:i], append([]string{"/" + param + "/*"}, patterns[i:]...)...)
							l++
							break
						}
					}
				}
				if subRes == nil {
					schemeName := strings.Replace(strings.TrimSuffix(p, ".json"), "/", ".", -1)
					if scheme, ok := res.GetSchemeOk(schemeName); ok {
						crubers = append(crubers, scheme)
						lastScheme = scheme
					}
				} else {
					if subRes.Config.Singleton && subRes.HasKey() {
						resCrumber = &ResourceCrumber{
							Resource: subRes,
							Parent:   append(resCrumber.Parent, resCrumber.Resource),
							ParentID: append(resCrumber.ParentID, resCrumber.ID),
							ID:       resCrumber.ID,
						}
					} else {
						resCrumber = &ResourceCrumber{
							Resource: subRes,
							Parent:   append(resCrumber.Parent, resCrumber.Resource),
							ParentID: append(resCrumber.ParentID, resCrumber.ID),
						}
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
		if scheme := ctx.Request.URL.Query().Get("!scheme"); scheme != "" {
			if s := ctx.Resource.GetScheme(scheme); s != nil {
				ctx.Scheme = s
			}
		}
	}

	appender := ctx.Breadcrumbs()

	if len(crubers) > 0 {
		if rc, ok := crubers[0].(*ResourceCrumber); ok {
			if rc.Resource.defaultMenu != nil && len(rc.Resource.defaultMenu.Ancestors) > 0 {
				var (
					m         = this.GetMenu(rc.Resource.defaultMenu.Ancestors[0])
					parts     = strings.Split(rc.Resource.Param, "/")
					ancestors = rc.Resource.defaultMenu.Ancestors[1:]
				)
				if m.Name != "" {
					appender.Append(core.NewBreadcrumb(ctx.Path(m.Name), string(ctx.Tt(m))))
				} else {
					appender.Append(core.NewBreadcrumb(ctx.Path(parts[0]), string(ctx.Tt(m))))
				}
				if len(parts) > 1 {
					parts = parts[1 : len(parts)-1] // exclude first and self resource
				} else {
					parts = parts[1:]
				}
				if l := len(parts); l > 0 && l >= len(ancestors) {
					for i := range ancestors {
						if m = m.GetMenu(parts[i]); m == nil {
							break
						} else {
							appender.Append(core.NewBreadcrumb(ctx.Path(parts[i]), string(ctx.Tt(m))))
						}
					}
				}
			}
		}
	}

	for _, crumber := range crubers {
		if !ctx.HasError() {
			crumbs, err := crumber.Breadcrumbs(ctx.Context)
			if err != nil {
				ctx.AddError(err)
				return
			}
			appender.Append(crumbs...)
		}
	}

	return
}

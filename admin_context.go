package admin

import (
	"net/http"

	"github.com/ecletus/core"
	"github.com/ecletus/render"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/logging"
)

// NewContext new admin context
func (this *Admin) NewContext(args ...interface{}) (c *Context) {
	if len(args) == 0 {
		return this.NewContext(&core.Context{})
	}
	var req *http.Request
	for i, arg := range args {
		switch ctx := arg.(type) {
		case *core.Site:
			return this.NewContext(ctx.NewContext())
		case *core.Context:
			c = &Context{Context: ctx}
		case http.ResponseWriter:
			if req == nil {
				req = args[i+1].(*http.Request)
			}
			_, coreCtx := this.ContextFactory.NewContextFromRequestPair(ctx, req, this.Config.MountPath)
			c = &Context{Context: coreCtx}
		case *http.Request:
			req = ctx
		}
	}

	if c != nil {
		var metaPath []string
		c.metaPath = &metaPath
		if c.Context == nil {
			_, c.Context = this.ContextFactory.NewContextFromRequestPair(c.Writer, c.Request, this.Config.MountPath)
			c.Request = c.Context.Request
		}
		c.Settings = map[string]interface{}{}
		c.Admin = this
		c.Context.SetValue(CONTEXT_KEY, c)

		if c.Request != nil {
			if v := c.Request.URL.Query().Get(P_DISPLAY); v != "" {
				c.Display = v
			}

			c.RequestLayout = c.Request.Header.Get("X-Layout")
		}
		c.Context.Init()

		for _, cb := range this.NewContextCallbacks {
			cb(c)
		}

		c.Yield = c.defaultYield
		var (
			err         error
			siteAssetFS assetfs.Interface
		)
		if siteAssetFS, err = c.Site.SystemStorage().AssetFS(); err != nil {
			logging.Tee(c.Site.Log, log).Warningf("get site assetfs failed: %s", err)
			siteAssetFS = assetfs.FakeFileSystem()
		}
		c.SiteTemplateFS = siteAssetFS.NameSpace("admin/templates")
		c.SiteStaticFS = siteAssetFS.NameSpace("admin/static")

		var templatesStack = make(PathStack, 0, 0)
		c.templatesStack = &templatesStack
		c.PageHandlers = &render.PageHandlers{}
	}
	return
}

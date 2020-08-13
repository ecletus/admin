package admin

import (
	"github.com/ecletus/auth"
	"github.com/ecletus/common"
	"github.com/ecletus/core"
)

// Auth is an auth interface that used to qor admin
// If you want to implement an authorization gateway for admin interface, you could implement this interface, and set it to the admin with `admin.SetAuth(auth)`
type Auth interface {
	GetCurrentUser(*Context) (common.User, error)
	LoginURL(*Context) string
	LogoutURL(*Context) string
	ProfileURL(c *Context) string
	Auth() *auth.Auth
	IsSuperAdmin(ctx *Context) (ok bool, err error)
	core.Permissioner
}

package admin_test

import (
	"net/http/httptest"

	"github.com/ecletus/admin"
	. "github.com/ecletus/admin/tests/dummy"
	"github.com/go-aorm/aorm"
)

var (
	server *httptest.Server
	db     *aorm.DB
	Admin  *admin.Admin
)

func init() {
	Admin = NewDummyAdmin()
	db = Admin.Config.DB
	server = httptest.NewServer(Admin.NewServeMux("/admin"))
}

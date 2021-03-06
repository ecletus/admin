package dummy

import (
	"fmt"

	"github.com/ecletus/admin"
	"github.com/ecletus/media"
	"github.com/ecletus/core"
	"github.com/ecletus/core/test/utils"
)

// NewDummyAdmin generate admin for dummy app
func NewDummyAdmin(keepData ...bool) *admin.Admin {
	var (
		db     = utils.TestDB()
		models = []interface{}{&User{}, &CreditCard{}, &Address{}, &Language{}, &Profile{}, &Phone{}, &Company{}}
		Admin  = admin.New(&core.NewConfig(db))
	)

	media.RegisterCallbacks(db)

	for _, value := range models {
		if len(keepData) == 0 {
			db.DropTableIfExists(value)
		}
		db.AutoMigrate(value)
	}

	Admin.AddResource(&Company{})
	Admin.AddResource(&Language{}, &admin.Config{Name: "语种 & 语言", Priority: -1})
	user := Admin.AddResource(&User{})
	user.Meta(&admin.Meta{
		Name: "CreditCard",
		Type: "single_edit",
	})
	user.Meta(&admin.Meta{
		Name: "Languages",
		Type: "select_many",
		Collection: func(resource interface{}, context *core.Context) (results [][]string) {
			if languages := []Language{}; !context.DB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{fmt.Sprint(language.ID), language.Name})
				}
			}
			return
		},
	})

	return Admin
}

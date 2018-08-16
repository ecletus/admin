package admin

import (
	"github.com/aghape/db"
)

type Plugin struct {
	db.DisDBNames
}

func (p *Plugin) OnRegister() {
	p.DBOnInitGorm(func(e *db.GormDBEvent) {
		DB := e.DB
		if DB.Callback().Query().Get("qor_admin:composite_primary_key") == nil {
			DB.Callback().Query().Before("gorm:query").Register("qor_admin:composite_primary_key", compositePrimaryKeyQueryCallback)
		}

		if DB.Callback().RowQuery().Get("qor_admin:composite_primary_key") == nil {
			DB.Callback().RowQuery().Before("gorm:row_query").Register("qor_admin:composite_primary_key", compositePrimaryKeyQueryCallback)
		}
	})
}

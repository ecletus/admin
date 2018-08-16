package admin

import (
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/aghape"
	"github.com/aghape/roles"
)

type SubConfig struct {
	Parent          *Resource
	ParentFieldName string
	ParentField     string
	FieldName       string
	Filters         []func(context *qor.Context, db *aorm.DB) *aorm.DB
	RawFieldFilter  map[string]interface{}
}

// Config admin config struct
type Config struct {
	Sub           *SubConfig
	Prefix        string
	Param         string
	Name          string
	PluralName    string
	Menu          []string
	Permission    *roles.Permission
	Themes        []ThemeInterface
	Displays      map[string]DisplayInterface
	Priority      int
	Singleton     bool
	Invisible     bool
	PageCount     int
	ID            string
	DisableFormID bool
	NotMount      bool
	Setup         func(res *Resource)
	MenuEnabled   func(menu *Menu, ctx *Context) bool
}

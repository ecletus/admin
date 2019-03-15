package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/aorm"
)

type SubConfig struct {
	Parent          *Resource
	ParentFieldName string
	ParentField     string
	FieldName       string
	Filters         []func(context *core.Context, db *aorm.DB) *aorm.DB
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
	Controller    interface{}
	LabelKey      string
}

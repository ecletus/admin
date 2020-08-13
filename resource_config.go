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
	Relation        *aorm.Relationship
	Filters         []*DBFilter
	RawFieldFilter  map[string]interface{}
}

// Config admin config struct
type Config struct {
	Dialect                 aorm.Dialector
	Sub                     *SubConfig
	Prefix                  string
	Param                   string
	Name                    string
	PluralName              string
	Menu                    []string
	Permission              *roles.Permission
	Themes                  []ThemeInterface
	Displays                map[string]DisplayInterface
	Priority                int
	Singleton               bool
	Invisible               bool
	PageCount               int
	UnlimitedPageCount      bool
	ID                      string
	DisableFormID           bool
	NotMount                bool
	Virtual                 bool
	Setup                   func(res *Resource)
	Setups                  []func(res *Resource)
	MenuEnabled             func(menu *Menu, ctx *Context) bool
	Controller              interface{}
	LabelKey                string
	DisableParentJoin       bool
	Duplicated              func(uid string, res *Resource)
	ViewControllerFactory   func(controller interface{}) interface{}
	ActionControllerFactory func(action *Action) ActionController
	DescriprionGetter       func(ctx *core.Context, record interface{}) string

	// apenas cria o resource, nao registra em nada, não monta
	Alone       bool

	ModelStruct *aorm.ModelStruct
}

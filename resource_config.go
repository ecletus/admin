package admin

import (
	"github.com/ecletus/roles"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
)

type SubConfig struct {
	Parent              *Resource
	ParentFieldName     string
	ParentField         string
	FieldName           string
	Relation            *resource.ParentRelationship
	Filters             []*DBFilter
	RawFieldFilter      map[string]interface{}
	MountAsItemDisabled bool
	ParentFilter        func(ctx *core.Context, db *aorm.DB, parentKey aorm.ID) (_ *aorm.DB, err error)
}

type CreateWizardConfig struct {
	Value  interface{}
	Config *Config
}

// Config admin config struct
type Config struct {
	Dialect            aorm.Dialector
	Sub                *SubConfig
	Prefix             string
	Param              string
	Name               string
	SubmitLabelKey     string
	SubmitLabel        string
	PluralName         string
	Menu               []string
	Permission         *roles.Permission
	Themes             []ThemeInterface
	Displays           map[string]DisplayInterface
	Priority           int
	Singleton          bool
	Invisible          bool
	PageCount          int
	UnlimitedPageCount bool
	UID, ID            string
	DisableFormID      bool
	NotMount           bool
	Virtual            bool
	// Protected disable registration into global registers
	Protected               bool
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
	RecordUriHandler        func(ctx *Context, record interface{}, parentKeys ...aorm.ID) string
	IndexUriHandler         func(ctx *Context, parentKeys ...aorm.ID) string

	Deletable            func(ctx *Context, record interface{}) bool
	DeleteableDB         func(ctx *Context, db *aorm.DB) *aorm.DB
	BulkDeletionDisabled bool

	Wizard *Wizard

	// apenas cria o resource, nao registra em nada, não monta
	Alone bool

	ModelStruct *aorm.ModelStruct

	CreateWizards []*CreateWizardConfig

	CreationAllowed func(ctx *Context) bool

	ParentPreload ParentPreloadFlag

	DefaultSectionsLayout string
}

func (this *Config) PrependSetup(f ...func(res *Resource)) {
	if this.Setup != nil {
		f = append(f, this.Setup)
		this.Setup = nil
	}
	this.Setups = append(f, this.Setups...)
}

func (this *Config) AppendSetup(f ...func(res *Resource)) {
	this.Setups = append(this.Setups, f...)
}

var DefaultUndeletedFields = func(ctx *Context) (m map[string]interface{}) {
	m = map[string]interface{}{}
	m[aorm.SoftDeleteFieldDeletedAt] = nil

	if _, ok := ctx.Resource.ModelStruct.FieldsByName[aorm.SoftDeleteFieldDeletedByID]; ok {
		m[aorm.SoftDeleteFieldDeletedByID] = nil
	}
	return
}

const (
	_ ParentPreloadFlag = iota << 1
	ParentPreloadIndex
	ParentPreloadNew
	ParentPreloadCreate
	ParentPreloadShow
	ParentPreloadEdit
	ParentPreloadUpdate
	ParentPreloadDelete
	ParentPreloadAction
)

type ParentPreloadFlag uint8

func (this ParentPreloadFlag) Has(flag ParentPreloadFlag) bool {
	return (this & flag) != 0
}

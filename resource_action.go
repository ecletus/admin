package admin

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/maps"

	"github.com/ecletus/roles"

	"github.com/ecletus/core/utils"
)

func (this *Resource) OnActionAdded(f func(action *Action)) {
	this.ActionAddedCallbacks = append(this.ActionAddedCallbacks, f)

	for _, action := range this.Actions {
		f(action)
	}
}

// Action register action for qor resource
func (this *Resource) Action(action *Action) *Action {
	if action.Name == "" {
		action.Name = utils.ToParamString(action.Label)
	} else if action.LabelKey == "" && action.Label == "" {
		action.Label = utils.HumanizeStringU(action.Name)
	}

	for _, a := range this.Actions {
		if a.Name == action.Name {
			if action.Label != "" {
				a.Label = action.Label
			}

			if action.Method != "" {
				a.Method = action.Method
			}

			if action.URL != nil {
				a.URL = action.URL
			}

			if action.URLOpenType != "" {
				a.URLOpenType = action.URLOpenType
			}

			if action.Visible != nil {
				a.Visible = action.Visible
			}

			if action.RecordAvailable != nil {
				a.RecordAvailable = action.RecordAvailable
			}

			if action.Available != nil {
				a.Available = action.Available
			}

			if action.Handler != nil {
				a.Handler = action.Handler
			}

			if len(action.Modes) != 0 {
				a.Modes = action.Modes
			}

			if action.Resource != nil {
				a.Resource = action.Resource
			}

			if action.Permission != nil {
				a.Permission = action.Permission
			}

			if action.PermissionMode != "" {
				a.PermissionMode = action.PermissionMode
			}

			if action.SkipRecordLoad {
				a.SkipRecordLoad = true
			}

			*action = *a
			return a
		}
	}

	if action.Label == "" {
		action.Label = utils.HumanizeString(action.Name)
	}

	if action.BaseResource == nil {
		action.BaseResource = this
	}

	if action.Method == "" {
		if action.ReadOnly() || action.URL != nil {
			action.Method = http.MethodGet
		} else {
			action.Method = http.MethodPut
		}
	}

	if action.URLOpenType == "" {
		if action.Resource != nil {
			action.URLOpenType = "bottomsheet"
		} else if action.Method == http.MethodGet {
			action.URLOpenType = "_blank"
		}
	}

	if action.Resource != nil {
		if action.Permission == nil && action.BaseResource.Config.Permission != nil {
			action.Resource.Permission = action.BaseResource.Config.Permission
		} else {
			action.Resource.Permission = roles.AllowAny(roles.Anyone)
		}

		for name := range action.Resource.ModelStruct.ChildrenByName {
			action.Resource.Meta(&Meta{
				Name:       name,
				Permission: roles.AllowAny(roles.Anyone),
			}).Resource.Permission = roles.AllowAny(roles.Anyone)
		}

	}

	if action.PermissionMode == "" {
		if action.ReadOnly() {
			action.PermissionMode = roles.Read
		} else {
			action.PermissionMode = roles.Update
			switch strings.ToUpper(action.Method) {
			case http.MethodPost:
				action.PermissionMode = roles.Create
			case http.MethodDelete:
				action.PermissionMode = roles.Delete
			case http.MethodGet:
				action.PermissionMode = roles.Read
			}
		}
	}

	if action.Permissioner == nil {
		action.Permissioner = this
	}

	this.Actions = append(this.Actions, action)

	// Register Actions into Router
	{
		var actionController ActionController
		if this.Config.ActionControllerFactory != nil {
			actionController = this.Config.ActionControllerFactory(action)
		} else {
			actionController = &ActionControl{action: action}
		}

		routeConfig := &RouteConfig{
			Permissioner:   action,
			PermissionMode: action.PermissionMode,
			CrumbsLoaderFunc: func(rh *RouteHandler, ctx *Context, pattern ...string) {
				this.Admin.LoadCrumbs(rh, ctx, pattern...)
			},
		}
		actionParam := "/" + action.ToParam()

		h := func() *RouteHandler {
			return NewHandler(actionController.Action, routeConfig)
		}

		if action.ReadOnly() && action.Method == http.MethodGet {
			this.ItemRoutes.Add(actionParam, &RouteNode{Action: action}).Get(h())
		} else if action.Resource != nil || action.Handler != nil {
			bulkPattern := "/!action" + actionParam
			if action.Resource != nil {
				if !this.IsSingleton() {
					// Bulk Action
					node := this.Routes.Add(bulkPattern, &RouteNode{Action: action}).Post(h())

					if !action.ReadOnly() {
						node.Get(NewHandler(actionController.Action, routeConfig))
						node.Put(NewHandler(actionController.Action, routeConfig))
					}
				}
				if !action.ItemDisabled {
					// Single Resource Action
					node := this.ItemRoutes.Add(actionParam, &RouteNode{Action: action}).Post(h())

					if !action.ReadOnly() {
						node.Get(h())
						node.Put(h())
					}
				}
			} else if action.Handler != nil {
				if !this.IsSingleton() {
					node := this.Routes.Add(bulkPattern, &RouteNode{Action: action}).Handle(action.Method, h())
					if action.EmptyBulkAllowed && action.Method != http.MethodGet {
						node.Get(h())
					}
				}
				// Single Resource action
				this.ItemRoutes.Add(actionParam, &RouteNode{Action: action}).Handle(action.Method, h())
			}
		}
	}

	for _, cb := range this.ActionAddedCallbacks {
		cb(action)
	}

	return action
}

// GetAction get defined action
func (this *Resource) GetAction(name string) *Action {
	for _, action := range this.Actions {
		if action.Name == name {
			return action
		}
	}
	return nil
}

// ActionArgument action argument that used in handle
type ActionArgument struct {
	Action              *Action
	PrimaryValues       []aorm.ID
	Context             *Context
	Argument            interface{}
	Record              interface{}
	Data                interface{}
	SkipDefaultResponse bool
	successMessage      string
}

func (this *ActionArgument) Success(message string) {
	this.successMessage = message
}

func (this *ActionArgument) Successf(format string, args ...interface{}) {
	this.successMessage = fmt.Sprintf(format, args...)
}

type ActionType int

const (
	ActionPrimary ActionType = iota
	ActionDefault
	ActionInfo
	ActionDanger
	ActionSuperDanger
)

var ActionTypeName = []string{
	"primary",
	"default",
	"info",
	"danger",
	"super_danger",
}

type Actions []*Action

func (this Actions) Sort() Actions {
	sort.Slice(this, func(i, j int) bool {
		a, b := this[i], this[j]
		if a.Type == b.Type {
			return a.Name < b.Name
		}
		if a.Type < b.Type {
			return true
		}
		return false
	})
	return this
}

const (
	ActionFormNew ActionFormType = iota + 1
	ActionFormShow
	ActionFormEdit
)

type ActionFormType uint8

// Action action definiation
type Action struct {
	Name        string
	Label       string
	LabelKey    string
	MdlIcon     string
	Method      string
	URL         func(record interface{}, context *Context, args ...interface{}) string
	URLOpenType string

	FindRecord func(s *Searcher) (rec interface{}, err error)

	Available,
	IndexVisible func(context *Context) bool

	Visible,
	RecordAvailable func(record interface{}, context *Context) bool

	Handler,
	ShowHandler,
	SetupArgument func(arg *ActionArgument) error
	Modes          []string
	BaseResource   *Resource
	Resource       *Resource
	Permission     *roles.Permission
	Permissioner   Permissioner
	Type           ActionType
	PermissionMode roles.PermissionMode
	FormType       ActionFormType
	ReturnURL,
	RefreshURL func(record interface{}, context *Context) string
	// executes and disable now
	One bool
	// allow executes with empty bulk records
	EmptyBulkAllowed bool
	ItemDisabled     bool // only on bulk mode
	//
	TargetWindow      bool
	PassCurrentParams bool
	SkipRecordLoad    bool

	Data          maps.Map
	TemplatePaths []string
	States        []*ActionState
}

func (this *Action) ReadOnly() bool {
	return this.FormType == ActionFormShow
}

func (this *Action) State(state ...*ActionState) {
	this.States = append(this.States, state...)
}

func (this *Action) GetData() maps.Interface {
	if this.Data == nil {
		this.Data = make(maps.Map)
	}
	return &this.Data
}

func (this *Action) ConfigSet(key, value interface{}) {
	this.Data.Set(key, value)
}

func (this *Action) ConfigGet(key interface{}) (value interface{}, ok bool) {
	return this.Data.Get(key)
}

func (this *Action) GetTemplatePaths() (pths []string) {
	if len(this.TemplatePaths) > 0 {
		return this.TemplatePaths
	}

	for _, pth := range this.BaseResource.GetTemplatePaths() {
		pths = append(pths, pth+"/actions/"+this.Name)
	}
	pths = append(pths, path.Join(strings.TrimSuffix(this.BaseResource.PkgPath, "/models"), "actions/"+this.Name))
	pths = append(pths, "actions/"+this.Name)
	return
}

func (this *Action) TypeName() string {
	return ActionTypeName[this.Type]
}

// ToParam used to register routes for actions
func (this *Action) ToParam() string {
	return utils.ToParamString(this.Name)
}

// IsAllowed check if current user has permission to view the action
func (this *Action) IsAllowed(context *Context, records ...interface{}) bool {
	if len(records) == 0 {
		if this.Available != nil && !this.Available(context) {
			return false
		}
		if this.IndexVisible != nil && !this.IndexVisible(context) {
			return false
		}
	} else {
		if this.RecordAvailable != nil {
			for _, record := range records {
				if !this.RecordAvailable(record, context) {
					return false
				}
			}
		}
		if this.Visible != nil {
			for _, record := range records {
				if !this.Visible(record, context) {
					return false
				}
			}
		}
	}

	if context.Roles.Has(roles.Anyone) {
		return true
	}

	if this.Permission != nil {
		return context.HasPermission(this, this.PermissionMode)
	}

	return context.HasPermission(this, this.PermissionMode)
}

// AdminHasContextPermission check if current user has permission for the action
func (this *Action) AdminHasPermission(mode roles.PermissionMode, context *Context) (perm roles.Perm) {
	if this.Permission != nil {
		return this.Permission.HasPermission(context, mode, context.Roles.Interfaces()...)
	}
	return this.Permissioner.AdminHasPermission(mode, context)
}

func (this *Action) GetLabelPair() ([]string, string) {
	var (
		keys = this.GetLabelKeys()
	)

	if this.Label != "" {
		return keys, this.Label
	}

	return keys, this.Name
}

func (this *Action) GetLabelKeys() []string {
	if this.LabelKey != "" {
		return []string{this.LabelKey}
	}

	return []string{this.BaseResource.I18nPrefix + ".actions." + this.Name, I18NGROUP + ".actions." + this.Name}
}

// FindSelectedRecords find selected records when run bulk actions
func (this *ActionArgument) FindSelectedRecords() []interface{} {
	if this.Record != nil {
		return []interface{}{this.Record}
	}
	var (
		context   = this.Context
		res       = context.Resource
		records   = []interface{}{}
		sqls      []string
		sqlParams []interface{}
	)

	if len(this.PrimaryValues) == 0 {
		return records
	}

	clone := context.Clone()
	for _, primaryValue := range this.PrimaryValues {
		primaryQuerySQL, primaryParams, err := resource.IdToPrimaryQuery(context.Context, res, false, primaryValue)
		if err != nil {
			context.AddError(err)
			return nil
		}
		sqls = append(sqls, primaryQuerySQL)
		sqlParams = append(sqlParams, primaryParams...)
	}

	if len(sqls) > 0 {
		clone.SetRawDB(clone.DB().Where(strings.Join(sqls, " OR "), sqlParams...))
	}

	results, err := clone.ParseFindMany()
	if err != nil {
		context.AddError(err)
		return nil
	}

	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	return records
}

func (this *Action) MessageKey(key string) string {
	return this.BaseResource.I18nPrefix + ".actions_messages." + this.ToParam() + "." + key
}

package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/roles"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
)

// Action register action for qor resource
func (this *Resource) Action(action *Action) *Action {
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
		if action.ReadOnly || action.URL != nil {
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

	if action.PermissionMode == "" {
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

		if action.ReadOnly {
			this.ItemRouter.Get(actionParam, NewHandler(actionController.Action, routeConfig))
		} else if action.Resource != nil || action.Handler != nil {
			bulkPattern := "/!action" + actionParam
			if action.Resource != nil {
				if !this.IsSingleton() {
					// Bulk Action
					this.Router.Get(bulkPattern, NewHandler(actionController.Action, routeConfig))
					this.Router.Put(bulkPattern, NewHandler(actionController.Action, routeConfig))
					this.Router.Post(bulkPattern, NewHandler(actionController.Action, routeConfig))
				}
				// Single Resource Action
				this.ItemRouter.Get(actionParam, NewHandler(actionController.Action, routeConfig))
				this.ItemRouter.Put(actionParam, NewHandler(actionController.Action, routeConfig))
				this.ItemRouter.Post(actionParam, NewHandler(actionController.Action, routeConfig))
			} else if action.Handler != nil {
				if !this.IsSingleton() {
					// Bulk Action
					this.Router.HandleMethod(action.Method, bulkPattern, NewHandler(actionController.Action, routeConfig))
				}
				// Single Resource action
				this.ItemRouter.HandleMethod(action.Method, actionParam, NewHandler(actionController.Action, routeConfig))
			}
		}
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

// Action action definiation
type Action struct {
	Name           string
	Label          string
	LabelKey       string
	MdlIcon        string
	Method         string
	URL            func(record interface{}, context *Context, args ...interface{}) string
	URLOpenType    string
	Available      func(context *Context) bool
	IndexVisible   func(context *Context) bool
	Visible        func(record interface{}, context *Context) bool
	Handler        func(argument *ActionArgument) error
	SetupArgument  func(argument *ActionArgument) error
	Modes          []string
	BaseResource   *Resource
	Resource       *Resource
	Permission     *roles.Permission
	Permissioner   core.Permissioner
	Type           ActionType
	PermissionMode roles.PermissionMode
	ReadOnly       bool
	ReturnURL      func(record interface{}, context *Context) string
	RefreshURL     func(record interface{}, context *Context) string
	// executes and disable now
	One bool
	// allow executes with empty bulk records
	EmptyBulkAllowed bool
	//
	TargetWindow      bool
	PassCurrentParams bool
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
		if this.IndexVisible != nil && !this.IndexVisible(context) {
			return false
		}
	} else if this.Visible != nil {
		for _, record := range records {
			if !this.Visible(record, context) {
				return false
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

// HasContextPermission check if current user has permission for the action
func (this *Action) HasPermission(mode roles.PermissionMode, context *core.Context) (perm roles.Perm) {
	if this.Permission != nil {
		return this.Permission.HasPermission(context, mode, context.Roles.Interfaces()...)
	}
	return this.Permissioner.HasPermission(mode, context)
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
func (actionArgument *ActionArgument) FindSelectedRecords() []interface{} {
	var (
		context   = actionArgument.Context
		res       = context.Resource
		records   = []interface{}{}
		sqls      []string
		sqlParams []interface{}
	)

	if len(actionArgument.PrimaryValues) == 0 {
		return records
	}

	clone := context.Clone()
	for _, primaryValue := range actionArgument.PrimaryValues {
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

	results, err := clone.FindMany()
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
	return this.BaseResource.I18nPrefix+".actions_messages."+this.ToParam()+"."+key
}
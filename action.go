package admin

import (
	"reflect"
	"strings"

	"github.com/aghape/core"
	"github.com/aghape/core/utils"
	"github.com/aghape/roles"
)

// Action register action for qor resource
func (res *Resource) Action(action *Action) *Action {
	for _, a := range res.Actions {
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

			*action = *a
			return a
		}
	}

	if action.Label == "" {
		action.Label = utils.HumanizeString(action.Name)
	}

	if action.Method == "" {
		if action.URL != nil {
			action.Method = "GET"
		} else {
			action.Method = "PUT"
		}
	}

	if action.URLOpenType == "" {
		if action.Resource != nil {
			action.URLOpenType = "bottomsheet"
		} else if action.Method == "GET" {
			action.URLOpenType = "_blank"
		}
	}

	res.Actions = append(res.Actions, action)

	// Register Actions into Router
	{
		actionController := &Controller{Admin: res.GetAdmin(), action: action}

		if action.Resource != nil || action.Handler != nil {
			routeConfig := &RouteConfig{Permissioner: action, PermissionMode: roles.Update}
			actionParam := "/" + action.ToParam()
			bulkPattern := "/!action" + actionParam

			if action.Resource != nil {
				// Bulk Action
				res.Router.Get(bulkPattern, NewHandler(actionController.Action, routeConfig))
				// Single Resource Action
				res.ObjectRouter.Get(actionParam, NewHandler(actionController.Action, routeConfig))
			} else if action.Handler != nil {
				// Bulk Action
				res.Router.Put(bulkPattern, NewHandler(actionController.Action, routeConfig))
				// Single Resource action
				res.ObjectRouter.Put(actionParam, NewHandler(actionController.Action, routeConfig))
			}
		}
	}

	return action
}

// GetAction get defined action
func (res *Resource) GetAction(name string) *Action {
	for _, action := range res.Actions {
		if action.Name == name {
			return action
		}
	}
	return nil
}

// ActionArgument action argument that used in handle
type ActionArgument struct {
	PrimaryValues       []string
	Context             *Context
	Argument            interface{}
	SkipDefaultResponse bool
}

type ActionType int

const (
	ActionDefault        ActionType = iota
	ActionPrimary
	ActionInfo
	ActionDanger
)

var ActionTypeName = []string{
	"default",
	"primary",
	"info",
	"danger",
}

// Action action definiation
type Action struct {
	Name        string
	Label       string
	Method      string
	URL         func(record interface{}, context *Context, args ...interface{}) string
	URLOpenType string
	Available   func(context *Context) bool
	Visible     func(record interface{}, context *Context) bool
	Handler     func(argument *ActionArgument) error
	Modes       []string
	Resource    *Resource
	Permission  *roles.Permission
	Type       ActionType
}

func (action *Action) TypeName() string{
	return ActionTypeName[action.Type]
}

// ToParam used to register routes for actions
func (action Action) ToParam() string {
	return utils.ToParamString(action.Name)
}

// IsAllowed check if current user has permission to view the action
func (action Action) IsAllowed(mode roles.PermissionMode, context *Context, records ...interface{}) bool {
	if action.Visible != nil {
		for _, record := range records {
			if !action.Visible(record, context) {
				return false
			}
		}
	}

	if action.Permission != nil {
		return action.HasPermission(mode, context.Context)
	}

	if context.Resource != nil {
		return context.Resource.HasPermission(mode, context.Context)
	}
	return true
}

// HasPermission check if current user has permission for the action
func (action Action) HasPermission(mode roles.PermissionMode, context *core.Context) bool {
	if action.Permission != nil {
		var roles = []interface{}{}
		for _, role := range context.Roles {
			roles = append(roles, role)
		}
		return action.Permission.HasPermission(mode, roles...)
	}

	return true
}

// FindSelectedRecords find selected records when run bulk actions
func (actionArgument *ActionArgument) FindSelectedRecords() []interface{} {
	var (
		context   = actionArgument.Context
		resource  = context.Resource
		records   = []interface{}{}
		sqls      []string
		sqlParams []interface{}
	)

	if len(actionArgument.PrimaryValues) == 0 {
		return records
	}

	clone := context.clone()
	for _, primaryValue := range actionArgument.PrimaryValues {
		primaryQuerySQL, primaryParams := resource.ToPrimaryQueryParams(primaryValue)
		sqls = append(sqls, primaryQuerySQL)
		sqlParams = append(sqlParams, primaryParams...)
	}

	if len(sqls) > 0 {
		clone.DB = clone.DB.Where(strings.Join(sqls, " OR "), sqlParams...)
	}
	results, _ := clone.FindMany()

	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	return records
}

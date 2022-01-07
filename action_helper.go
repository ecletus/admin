package admin

import (
	"net/http"
)

func ActionIndexBulkExport(res *Resource, action *Action, handler func(arg *ActionArgument) (err error)) *Action {
	action.EmptyBulkAllowed = true
	action.TargetWindow = true
	action.Modes = []string{"index"}
	action.Method = http.MethodPost
	action.FormType = ActionFormShow

	if action.Name == "" {
		action.Name = "bulk_export"
		action.Label = "Export"
	}
	if action.LabelKey == "" {
		action.LabelKey = res.I18nPrefix + ".actions." + action.Name
	}
	action.Handler = handler
	return res.Action(action)
}

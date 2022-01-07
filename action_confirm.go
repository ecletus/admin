package admin

type (
	Confirm struct{}
)

func ActionConfirm(Admin *Admin, action *Action) *Action {
	action.Resource = Admin.NewResource(&Confirm{})
	action.Resource.TemplatePath = PKG + "/models/confirm"
	return action
}
func ActionConfirmNotUndo(Admin *Admin, action *Action, setup ...func(arg *ActionArgument) error) *Action {
	action.Resource = Admin.NewResource(&Confirm{})
	action.Resource.TemplatePath = PKG + "/models/confirm_not_undo"
	action.SetupArgument = func(arg *ActionArgument) (err error) {
		arg.Argument = &Confirm{}
		for _, s := range setup {
			if err = s(arg); err != nil {
				return
			}
		}
		return
	}
	return action
}

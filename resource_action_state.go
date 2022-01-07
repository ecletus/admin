package admin

type ActionState struct {
	Name         string
	Value        string
	CssClass     string
	CssClassFunc func(arg *ActionArgument) string
	Label        string
	LabelFunc    func(arg *ActionArgument) string
	ValueFunc    func(arg *ActionArgument) string
	Enabled      func(arg *ActionArgument) bool
	Handler      func(arg *ActionArgument, val string, render func()) error
}

func (this *ActionState) IsEnabled(arg *ActionArgument) bool {
	if this.Enabled != nil {
		return this.Enabled(arg)
	}
	return true
}

func (this *ActionState) GetCssClass(arg *ActionArgument) string {
	if this.CssClassFunc != nil {
		return this.CssClassFunc(arg)
	}
	return this.CssClass
}

func (this *ActionState) GetLabel(arg *ActionArgument) string {
	if this.LabelFunc != nil {
		return this.LabelFunc(arg)
	}
	return this.Label
}

func (this *ActionState) GetValue(arg *ActionArgument) string {
	if this.ValueFunc != nil {
		return this.ValueFunc(arg)
	}
	return this.Value
}

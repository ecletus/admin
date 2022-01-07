package admin

type UcState struct {
	Name         string
	CssClass     string
	CssClassFunc func(ctx *Context) string
	Label        string
	LabelFunc    func(ctx *Context) string
	Enabled      func(ctx *Context) bool
	Handler      func(ctx *Context, messages *[]string, render func()) error
}

func (this *UcState) IsEnabled(arg *Context) bool {
	if this.Enabled != nil {
		return this.Enabled(arg)
	}
	return true
}

func (this *UcState) GetCssClass(arg *Context) string {
	if this.CssClassFunc != nil {
		return this.CssClassFunc(arg)
	}
	return this.CssClass
}

func (this *UcState) GetLabel(arg *Context) string {
	if this.LabelFunc != nil {
		return this.LabelFunc(arg)
	}
	if this.Label != "" {
		return this.Label
	}
	return this.Name
}

type UcStates struct {
	CreateStates []*UcState
	UpdateStates []*UcState
}

func (this *UcStates) CreateState(state ...*UcState) {
	this.CreateStates = append(this.CreateStates, state...)
}

func (this *UcStates) UpdateState(state ...*UcState) {
	this.UpdateStates = append(this.UpdateStates, state...)
}

package admin

func (this *Context) DefaulLayout(layout ...string) {
	if this.Layout == "" {
		if len(layout) == 0 || layout[0] == "" {
			layout = []string{this.Type.Clear(DELETED).S()}
		}
		this.Layout = layout[0]
	}
}

func (this *Context) Basic() *Context {
	this.Layout = "basic"
	return this
}

func (this *Context) Readonly() *Context {
	this.Layout = "readonly"
	return this
}

func (this *Context) ForUpdate() *Context {
	this.Layout = ""
	return this
}

func (this *Context) IsReadonly() bool {
	return this.Layout == "readonly"
}

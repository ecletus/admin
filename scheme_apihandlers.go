package admin

type ApiHandler struct {
	Ext     string
	Handler func(ctx *Context)
}

type ApiHandlerRegistrator interface {
	InheritsFrom(registrator ...ApiHandlerRegistrator)
	Add(handler ...*ApiHandler)
	Get(ext string) *ApiHandler
	Each(state map[string]*ApiHandler, cb func(f *ApiHandler) (err error)) (err error)
}

type ApiHandlerRegister struct {
	handlers     []*ApiHandler
	byExt        map[string]*ApiHandler
	inheritsFrom []ApiHandlerRegistrator
}

func (this *ApiHandlerRegister) InheritsFrom(registrator ...ApiHandlerRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *ApiHandlerRegister) Add(handler ...*ApiHandler) {
	this.handlers = append(this.handlers, handler...)
	if this.byExt == nil {
		this.byExt = make(map[string]*ApiHandler, len(handler))
	}
	for _, f := range handler {
		this.byExt[f.Ext] = f
	}
}

func (this *ApiHandlerRegister) Get(ext string) (h *ApiHandler) {
	if this.byExt != nil {
		if h = this.byExt[ext]; h != nil {
			return
		}
	}
	for _, parent := range this.inheritsFrom {
		if h = parent.Get(ext); h != nil {
			return
		}
	}
	return
}

func (this *ApiHandlerRegister) Each(state map[string]*ApiHandler, cb func(f *ApiHandler) (err error)) (err error) {
	if this.byExt != nil {
		var ok bool
		for ext, f := range this.byExt {
			if _, ok = state[ext]; !ok {
				state[ext] = f
				if err = cb(f); err == ErrStopIteration {
					return nil
				} else if err != nil {
					return
				}
			}
		}
	}
	for _, parent := range this.inheritsFrom {
		if err = parent.Each(state, cb); err == ErrStopIteration {
			return nil
		} else if err != nil {
			return
		}
	}
	return nil
}

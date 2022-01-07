package admin

type NamedSearcherHandler struct {
	Name    string
	Handler func(s *Searcher)
}

type NamedSearcherHandlersRegistrator interface {
	InheritsFrom(registrator ...NamedSearcherHandlersRegistrator)
	Add(handler ...*NamedSearcherHandler)
	Get(name string) *NamedSearcherHandler
	Each(cb func(handler *NamedSearcherHandler) (err error), state ...map[string]*NamedSearcherHandler) (err error)
}

type NamedSearcherHandlersRegistry struct {
	handlers     []*NamedSearcherHandler
	byName       map[string]*NamedSearcherHandler
	inheritsFrom []NamedSearcherHandlersRegistrator
}

func (this *NamedSearcherHandlersRegistry) InheritsFrom(registrator ...NamedSearcherHandlersRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *NamedSearcherHandlersRegistry) Add(handler ...*NamedSearcherHandler) {
	this.handlers = append(this.handlers, handler...)
	if this.byName == nil {
		this.byName = make(map[string]*NamedSearcherHandler, len(handler))
	}
	for _, s := range handler {
		this.byName[s.Name] = s
	}
}

func (this *NamedSearcherHandlersRegistry) Each(cb func(handler *NamedSearcherHandler) (err error), state ...map[string]*NamedSearcherHandler) (err error) {
	var s map[string]*NamedSearcherHandler
	for _, s = range state {
	}
	if s == nil {
		s = map[string]*NamedSearcherHandler{}
	}

	if this.byName != nil {
		var ok bool
		for name, f := range this.byName {
			if _, ok = s[name]; !ok {
				s[name] = f
				if err = cb(f); err == ErrStopIteration {
					return nil
				} else if err != nil {
					return
				}
			}
		}
	}
	for _, parent := range this.inheritsFrom {
		if err = parent.Each(cb, s); err == ErrStopIteration {
			return nil
		} else if err != nil {
			return
		}
	}
	return nil
}

func (this *NamedSearcherHandlersRegistry) Get(name string) (f *NamedSearcherHandler) {
	if this.byName != nil {
		if f = this.byName[name]; f != nil {
			return
		}
	}
	for _, parent := range this.inheritsFrom {
		if f = parent.Get(name); f != nil {
			return
		}
	}
	return
}

func GetNamedSearcherHandlers(this NamedSearcherHandlersRegistrator, names ...string) (handlers []*NamedSearcherHandler, err error) {
	ok := func(f *NamedSearcherHandler) bool {
		return true
	}

	if len(names) > 0 {
		only := map[string]interface{}{}
		for _, name := range names {
			only[name] = nil
		}
		ok = func(f *NamedSearcherHandler) (ok bool) {
			_, ok = only[f.Name]
			return
		}
	}
	err = this.Each(func(f *NamedSearcherHandler) (err error) {
		if ok(f) {
			handlers = append(handlers, f)
		}
		return nil
	})
	return
}

func MustGetNamedSearcherHandlers(this NamedSearcherHandlersRegistrator, names ...string) (handlers []*NamedSearcherHandler) {
	var err error
	if handlers, err = GetNamedSearcherHandlers(this, names...); err != nil {
		panic(err)
	}
	return
}

func CallNamedSearcherHandlers(s *Searcher, this NamedSearcherHandlersRegistrator) {
	this.Each(func(handler *NamedSearcherHandler) (err error) {
		handler.Handler(s)
		return nil
	})
}

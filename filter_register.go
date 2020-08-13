package admin

import (
	"errors"
	"sync/atomic"
)

var ErrStopIteration = errors.New("stop iteration")
var filterIndex uint32

type FilterRegistrator interface {
	InheritsFrom(registrator ...FilterRegistrator)
	AddFilter(filter ...*Filter)
	Get(name string) *Filter
	Each(state map[string]*Filter, cb func(f *Filter) (err error)) (err error)
	EachDefaults(cb func(f *Filter))
}

type FilterRegister struct {
	filters      []*Filter
	byName       map[string]*Filter
	inheritsFrom []FilterRegistrator
}

func (this *FilterRegister) InheritsFrom(registrator ...FilterRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *FilterRegister) AddFilter(filter ...*Filter) {
	this.filters = append(this.filters, filter...)
	if this.byName == nil {
		this.byName = make(map[string]*Filter, len(filter))
	}
	for _, f := range filter {
		this.byName[f.Name] = f
		f.index = atomic.AddUint32(&filterIndex, 1)
	}
}

func (this *FilterRegister) Each(state map[string]*Filter, cb func(f *Filter) (err error)) (err error) {
	if this.byName != nil {
		var ok bool
		for name, f := range this.byName {
			if _, ok = state[name]; !ok {
				state[name] = f
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

func (this *FilterRegister) Get(name string) (f *Filter) {
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

func (this *FilterRegister) EachDefaults(cb func(f *Filter)) {
	this.Each(map[string]*Filter{}, func(f *Filter) (err error) {
		if f.HandleEmpty {
			cb(f)
		}
		return nil
	})
	return
}

package admin

import (
	"errors"
	"sync/atomic"

	hash_key_gen "github.com/unapu-go/hash-key-gen"
)

var ErrStopIteration = errors.New("stop iteration")
var filterIndex uint32

type FilterRegistrator interface {
	InheritsFrom(registrator ...FilterRegistrator)
	AddFilter(filter ...*Filter)
	Get(id uintptr) *Filter
	GetByName(name string) *Filter
	Each(state map[string]*Filter, cb func(f *Filter) (err error)) (err error)
	EachDefaults(cb func(f *Filter))
}

type FilterRegister struct {
	filters      []*Filter
	byID         map[uintptr]*Filter
	inheritsFrom []FilterRegistrator
}

func (this *FilterRegister) InheritsFrom(registrator ...FilterRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *FilterRegister) AddFilter(filter ...*Filter) {
	this.filters = append(this.filters, filter...)
	if this.byID == nil {
		this.byID = make(map[uintptr]*Filter, len(filter))
	}
	for _, f := range filter {
		if f.ID == 0 {
			f.ID = hash_key_gen.OfString(f.Name)
		}
		this.byID[f.ID] = f
		f.index = atomic.AddUint32(&filterIndex, 1)
	}
}

func (this *FilterRegister) Each(state map[string]*Filter, cb func(f *Filter) (err error)) (err error) {
	if this.byID != nil {
		var ok bool
		for _, f := range this.byID {
			if _, ok = state[f.Name]; !ok {
				state[f.Name] = f
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

func (this *FilterRegister) Get(id uintptr) (f *Filter) {
	if this.byID != nil {
		if f = this.byID[id]; f != nil {
			return
		}
	}
	for _, parent := range this.inheritsFrom {
		if f = parent.Get(id); f != nil {
			return
		}
	}
	return
}

func (this *FilterRegister) GetByName(name string) (f *Filter) {
	for _, f := range this.filters {
		if f.Name == name {
			return f
		}
	}
	for _, parent := range this.inheritsFrom {
		if f = parent.GetByName(name); f != nil {
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

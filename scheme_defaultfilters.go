package admin

import (
	"github.com/moisespsena-go/aorm"
)

type DBFilterFunc func(context *Context, db *aorm.DB) (DB *aorm.DB, err error)

type DBFilter struct {
	Name    string
	Handler DBFilterFunc
}

type DBFilterRegistrator interface {
	InheritsFrom(registrator ...DBFilterRegistrator)
	AddFilter(filter ...*DBFilter)
	Get(name string) *DBFilter
	Each(state map[string]*DBFilter, cb func(f *DBFilter) (err error)) (err error)
}

type DBFilterRegister struct {
	filters      []*DBFilter
	byName       map[string]*DBFilter
	inheritsFrom []DBFilterRegistrator
}

func (this *DBFilterRegister) InheritsFrom(registrator ...DBFilterRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *DBFilterRegister) AddFilter(filter ...*DBFilter) {
	this.filters = append(this.filters, filter...)
	if this.byName == nil {
		this.byName = make(map[string]*DBFilter, len(filter))
	}
	for _, f := range filter {
		if f.Name != "" {
			this.byName[f.Name] = f
		}
	}
}

func (this *DBFilterRegister) Each(state map[string]*DBFilter, cb func(f *DBFilter) (err error)) (err error) {
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

func (this *DBFilterRegister) Get(name string) (f *DBFilter) {
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

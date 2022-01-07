package admin

import (
	"github.com/ecletus/core/utils"
)

type ScopeRegistrator interface {
	InheritsFrom(registrator ...ScopeRegistrator)
	Add(scope ...*Scope)
	Get(name string) *Scope
	Each(cb func(scope *Scope) (err error), state ...map[string]*Scope) (err error)
}

type ScopeRegister struct {
	scopes       []*Scope
	byName       map[string]*Scope
	inheritsFrom []ScopeRegistrator
}

func (this *ScopeRegister) InheritsFrom(registrator ...ScopeRegistrator) {
	this.inheritsFrom = append(this.inheritsFrom, registrator...)
}

func (this *ScopeRegister) Add(scope ...*Scope) {
	this.scopes = append(this.scopes, scope...)
	if this.byName == nil {
		this.byName = make(map[string]*Scope, len(scope))
	}
	for _, s := range scope {
		if s.Label == "" {
			if s.Name == "" {
				panic("ScopeRegister.Add: unamed scope")
			}
			s.Label = utils.HumanizeString(s.Name)
		} else if s.Name == "" {
			s.Label = utils.ToParamString(s.Label)
		}
		this.byName[s.Name] = s
	}
}

func (this *ScopeRegister) Each(cb func(scope *Scope) (err error), state ...map[string]*Scope) (err error) {
	var s map[string]*Scope
	for _, s = range state {
	}
	if s == nil {
		s = map[string]*Scope{}
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

func (this *ScopeRegister) Get(name string) (f *Scope) {
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

// Scope register scope for qor resource
func (this *Scheme) Scope(scope *Scope) {
	if scope.ScopeConfig == nil {
		scope.ScopeConfig = &ScopeConfig{}
	}
	if scope.Label == "" {
		scope.Label = utils.HumanizeString(scope.Name)
	} else if scope.Name == "" {
		scope.Name = utils.ToParamString(scope.Label)
	}
	scope.BaseResource = this.Resource
	this.Scopes.Add(scope)
}

func (this *Scheme) GetScope(name string) *Scope {
	return this.Scopes.Get(name)
}

func (this *Scheme) MustGetScopes() (scopes []*Scope) {
	var err error
	if scopes, err = this.GetScopes(); err != nil {
		panic(err)
	}
	return
}

func (this *Scheme) GetScopes() (scopes []*Scope, err error) {
	ok := func(f *Scope) bool {
		return true
	}
	err = this.Scopes.Each(func(f *Scope) (err error) {
		if ok(f) {
			scopes = append(scopes, f)
		}
		return nil
	})
	return
}

package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

func (this *Resource) OnBeforeCreate(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_CREATE)
}

func (this *Resource) OnAfterCreate(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_CREATE)
}

func (this *Resource) OnBeforeUpdate(f ...func(ctx *core.Context, old, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Old(), e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_UPDATE)
}

func (this *Resource) OnAfterUpdate(f ...func(ctx *core.Context, old, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Old(), e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_UPDATE)
}

func (this *Resource) OnBeforeDelete(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_DELETE)
}

func (this *Resource) OnAfterDelete(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_DELETE)
}

func (this *Resource) OnBeforeFindOne(f ...func(ctx *core.Context) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_FIND_ONE)
}

func (this *Resource) OnAfterFindOne(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_FIND_ONE)
}

func (this *Resource) OnBeforeFindMany(f ...func(ctx *core.Context) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_FIND_MANY)
}

func (this *Resource) OnAfterFindMany(f ...func(ctx *core.Context, record interface{}) error) {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_FIND_MANY)
}

func (this *Resource) OnBeforeFind(f ...func(ctx *core.Context) error) {
	this.OnBeforeFindMany(f...)
	this.OnBeforeFindOne(f...)
}

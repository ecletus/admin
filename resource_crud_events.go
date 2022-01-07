package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

func (this *Resource) OnBeforeSave(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	this.OnBeforeCreate(f...)
	this.OnBeforeUpdate(func(ctx *core.Context, old, record interface{}) (err error) {
		for _, f := range f {
			if err = f(ctx, record); err != nil {
				return
			}
		}
		return
	})
	return this
}

func (this *Resource) OnBeforeCreate(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_CREATE)
	return this
}

func (this *Resource) OnAfterCreate(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_CREATE)
	return this
}

func (this *Resource) OnBeforeUpdate(f ...func(ctx *core.Context, old, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Old(), e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_UPDATE)
	return this
}

func (this *Resource) OnAfterUpdate(f ...func(ctx *core.Context, old, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Old(), e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_UPDATE)
	return this
}

func (this *Resource) OnBeforeDelete(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_DELETE)
	return this
}

func (this *Resource) OnAfterDelete(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_DELETE)
	return this
}

func (this *Resource) OnBeforeFindOne(f ...func(ctx *core.Context) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_FIND_ONE)
	return this
}

func (this *Resource) OnAfterFindOne(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_FIND_ONE)
	return this
}

func (this *Resource) OnBeforeFindMany(f ...func(ctx *core.Context) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_FIND_MANY)
	return this
}

func (this *Resource) OnAfterFindMany(f ...func(ctx *core.Context, record interface{}) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context, e.Result()); err != nil {
				return
			}
		}
		return
	}, resource.AFTER|resource.E_DB_ACTION_FIND_MANY)
	return this
}

func (this *Resource) OnBeforeFind(f ...func(ctx *core.Context) error) *Resource {
	this.OnBeforeFindMany(f...)
	this.OnBeforeFindOne(f...)
	return this
}

func (this *Resource) OnBeforeCount(f ...func(ctx *core.Context) error) *Resource {
	_ = this.OnDBActionE(func(e *resource.DBEvent) (err error) {
		for _, f := range f {
			if err = f(e.Context); err != nil {
				return
			}
		}
		return
	}, resource.BEFORE|resource.E_DB_ACTION_COUNT)
	return this
}

func (this *Resource) OnBeforeFindAndCount(f ...func(ctx *core.Context) error) *Resource {
	this.OnBeforeFind(f...)
	this.OnBeforeCount(f...)
	return this
}

func (this *Resource) OnBeforeFindManyAndCount(f ...func(ctx *core.Context) error) *Resource {
	this.OnBeforeFindMany(f...)
	this.OnBeforeCount(f...)
	return this
}

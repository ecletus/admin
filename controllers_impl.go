package admin

import (
	"errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
)

// Search

type SearchController struct{}

func (SearchController) Search(context *Context) (result interface{}) {
	var err error
	if result, err = context.ParseFindMany(); err != nil {
		context.AddError(err)
	}
	return
}

// Index
type IndexController struct{}

func (IndexController) Index(context *Context) (result interface{}) {
	if context.ValidateLayout() {
		var err error
		result, err = context.ParseFindMany()
		context.AddError(err)
	}
	return
}

type IndexSearchController struct {
	IndexController
	SearchController
}

// Create

type CreateController struct{}

func (CreateController) New(context *Context) interface{} {
	return context.Resource.NewStruct(context.Site)
}

func (CreateController) Create(context *Context, recorde interface{}) {
	context.AddError(context.WithTransaction(func() error {
		return context.Crud().Create(recorde)
	}))
}

// Read

func ReadControllerRead(context *Context) (recorde interface{}) {
	var err error
	res := context.Resource

	if res.Config.Singleton && !res.HasKey() {
		recorde = res.NewStruct(context.Site)
		var ctx *core.Context
		if ctx, err = res.ApplyDefaultFilters(context); err == nil {
			ctx.WithDB(func(ctx *core.Context) {
				err = context.Crud(ctx).FindOne(recorde)
			}, ctx.DB().Opt(aorm.OptPreloadTagNames(context.Type.String())))
		}
	} else {
		if context.Type.Has(DELETED) {
			var (
				qnt   int
				query string
				key   []interface{}
			)
			if query, key, err = resource.IdToPrimaryQuery(context.Context, res, false, context.ResourceID); err != nil {
				context.AddError(err)
				return
			}
			err := context.DB().Model(res.Value).
				Where("_.deleted_at IS NOT NULL").
				Where(query, key...).
				Count(&qnt).Error
			if err != nil {
				context.AddError(err)
				return
			}
			context.SetRawDB(context.DB().Unscoped())
		}
		recorde, err = context.FindOne()
	}
	if err != nil {
		if !aorm.IsRecordNotFoundError(err) {
			context.AddError(err)
		}
		recorde = nil
	}
	return
}

type ReadController struct {
	Record     interface{}
	RecordFunc func(context *Context) interface{}
}

func (this ReadController) Read(context *Context) (recorde interface{}) {
	if this.RecordFunc != nil {
		return this.RecordFunc(context)
	}
	if this.Record != nil {
		return this.Record
	}
	return ReadControllerRead(context)
}

// Update

type UpdateController struct{}

func (UpdateController) Update(ctx *Context, recorde interface{}, old ...interface{}) {
	ctx.AddError(ctx.WithTransaction(func() (err error) {
		if err = ctx.Crud().Update(recorde, old...); err != nil || ctx.HasError() || ctx.DecoderExcludes == nil {
			return
		}
		return
	}))
}

// Delete

type DeleteController struct {
}

func (DeleteController) Delete(context *Context, recorde interface{}) {
	context.AddError(context.WithTransaction(func() error {
		if ctx, err := context.Resource.ApplyDefaultFilters(context); err != nil {
			return err
		} else {
			return context.Crud(ctx).Delete(recorde)
		}
	}))
}

type DeleteBulkController struct {
	DeleteController
}

func (DeleteBulkController) DeleteBulk(context *Context, recorde ...interface{}) {
	context.AddError(context.WithTransaction(func() error {
		ctx, err := context.Resource.ApplyDefaultFilters(context)
		if err != nil {
			return err
		}
		var qnt uint32
		for _, recorde := range recorde {
			if recorde == nil {
				continue
			}
			ctx := ctx.Clone()
			ctx.ResourceID = context.Resource.GetKey(recorde)
			if err = context.Crud(ctx).Delete(recorde); err != nil {
				break
			}
			qnt++
		}
		if qnt == 0 {
			return aorm.ErrRecordNotFound
		}
		return err
	}))
}

type RestoreController struct {
}

func (RestoreController) DeletedIndex(context *Context) (result interface{}) {
	if context.Scheme.SchemeName != A_DELETED_INDEX {
		context.AddError(errors.New("unexpected scheme"))
		return
	}
	if context.ValidateLayout() {
		context.WithDB(func(context *Context) {
			var err error
			result, err = context.ParseFindMany()
			context.AddError(err)
		})
	}
	return
}

func (RestoreController) Restore(context *Context, key ...aorm.ID) {
	context.AddError(context.WithTransaction(func() error {
		return context.Resource.Restore(context, key...)
	}))
}

type IndexReadController struct {
	IndexController
	ReadController
}

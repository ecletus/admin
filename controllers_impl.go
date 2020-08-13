package admin

import (
	"errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

// Search

type SearchController struct{}

func (SearchController) Search(context *Context) (result interface{}) {
	var err error
	if result, err = context.FindMany(); err != nil {
		context.AddError(err)
	}
	return
}

// Index
type IndexController struct{}

func (IndexController) Index(context *Context) (result interface{}) {
	if context.ValidateLayout() {
		var err error
		result, err = context.FindMany()
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
	context.WithTransaction(func() {
		context.AddError(context.Crud().Create(recorde))
	})
}

// Read

func ReadControllerRead(context *Context) (recorde interface{}) {
	var err error
	res := context.Resource

	if res.Config.Singleton && !res.HasKey() {
		recorde = res.NewStruct(context.Site)
		var ctx *core.Context
		if ctx, err = res.ApplyDefaultFilters(context.Context); err == nil {
			err = context.Crud(ctx).FindOne(recorde)
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
				Where("deleted_at IS NOT NULL").
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
	Recorde     interface{}
	RecordeFunc func(context *Context) interface{}
}

func (this ReadController) Read(context *Context) (recorde interface{}) {
	if this.RecordeFunc != nil {
		return this.RecordeFunc(context)
	}
	if this.Recorde != nil {
		return this.Recorde
	}
	return ReadControllerRead(context)
}

// Update

type UpdateController struct{}

func (UpdateController) Update(context *Context, recorde interface{}, old ...interface{}) {
	context.WithTransaction(func() {
		context.AddError(context.Crud().Update(recorde, old...))
	})
}

// Delete

type DeleteController struct {
}

func (DeleteController) Delete(context *Context, recorde interface{}) {
	context.WithTransaction(func() {
		if ctx, err := context.Resource.ApplyDefaultFilters(context.Context); err != nil {
			context.AddError(err)
		} else {
			context.AddError(context.Crud(ctx).Delete(recorde))
		}
	})
}

type DeleteBulkController struct {
	DeleteController
}

func (DeleteBulkController) DeleteBulk(context *Context, recorde ...interface{}) {
	context.WithTransaction(func() {
		ctx, err := context.Resource.ApplyDefaultFilters(context.Context)
		if err != nil {
			context.AddError(err)
			return
		}
		for _, recorde := range recorde {
			ctx := ctx.Clone()
			ctx.ResourceID = context.Resource.GetKey(recorde)
			context.AddError(context.Crud(ctx).Delete(recorde))
			if context.HasError() {
				break
			}
		}
	})
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
			result, err = context.FindMany()
			context.AddError(err)
		})
	}
	return
}

func (RestoreController) Restore(context *Context, key ...aorm.ID) {
	context.WithTransaction(func() {
		context.Resource.Restore(context, key...)
	})
}

type IndexReadController struct {
	IndexController
	ReadController
}
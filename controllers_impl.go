package admin

import (
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

// Search

type SearchController struct{}

func (SearchController) Search(context *Context) interface{} {
	items, err := context.FindMany()
	if err != nil {
		context.AddError(err)
	}
	return items
}

// Index

type IndexController struct{}

func (IndexController) Index(context *Context) interface{} {
	var result interface{}
	if context.ValidateLayout() {
		var err error
		result, err = context.FindMany()
		context.AddError(err)
	}
	return result
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
		err = context.Crud(res.ApplyDefaultFilters(context.Context)).FindMany(recorde)
	} else {
		if context.Type.Has(DELETED) {
			var qnt int
			query, key := resource.StringToPrimaryQuery(res, context.ResourceID)
			err := context.DB.
				Table(res.FakeScope.TableName()).
				Where("deleted_at IS NOT NULL").
				Where(query, key...).
				Count(&qnt).Error
			if err != nil {
				context.AddError(err)
				return
			}
			context.SetDB(context.DB.Unscoped())
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

type ReadController struct{}

func (ReadController) Read(context *Context) (recorde interface{}) {
	return ReadControllerRead(context)
}

// Update

type UpdateController struct{}

func (UpdateController) Update(context *Context, recorde interface{}) {
	context.WithTransaction(func() {
		context.AddError(context.Crud().Update(recorde))
	})
}

// Delete

type DeleteController struct {
}

func (DeleteController) Delete(context *Context, recorde interface{}) {
	context.WithTransaction(func() {
		ctx := context.Resource.ApplyDefaultFilters(context.Context)
		context.AddError(context.Crud(ctx).Delete(recorde))
	})
}

type DeleteBulkController struct {
	DeleteController
}

func (DeleteBulkController) DeleteBulk(context *Context, recorde ...interface{}) {
	context.WithTransaction(func() {
		ctx := context.Clone()
		ctx = context.Resource.ApplyDefaultFilters(context.Context)
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

func (RestoreController) DeletedIndex(context *Context) interface{} {
	var result interface{}
	if context.ValidateLayout() {
		var err error
		context.SetDB(context.DB.Where(aorm.IQ("{}.deleted_at IS NOT NULL")).Unscoped())
		result, err = context.FindMany()
		context.AddError(err)
	}
	return result
}

func (RestoreController) Restore(context *Context, key ...string) {
	context.WithTransaction(func() {
		context.Resource.Restore(context, key...)
	})
}

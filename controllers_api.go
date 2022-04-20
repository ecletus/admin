package admin

import "github.com/go-aorm/aorm"

type ControllerSearcher interface {
	Search(context *Context) (recordes interface{})
}

type ControllerIndex interface {
	Index(context *Context) (recordes interface{})
}

type ControllerCreator interface {
	New(context *Context) interface{}
	Create(context *Context, recorde interface{})
}

type ControllerReader interface {
	Read(context *Context) (recorde interface{})
}

type ControllerUpdater interface {
	Update(context *Context, recorde interface{}, old ...interface{})
}

type ControllerDeleter interface {
	Delete(context *Context, recorde interface{})
}

type ControllerBulkDeleter interface {
	ControllerDeleter
	DeleteBulk(context *Context, recorde ...interface{})
}

type ControllerCruder interface {
	ControllerCreator
	ControllerReader
	ControllerUpdater
	ControllerBulkDeleter
}

type ControllerCrudIndexer interface {
	ControllerCruder
	ControllerIndex
}

type ControllerCrudSearcher interface {
	ControllerCruder
	ControllerSearcher
}

type ControllerCrudSearchIndexer interface {
	ControllerSearcher
	ControllerIndex
}

type ControllerRestorer interface {
	DeletedIndex(context *Context) (recordes interface{})
	Restore(context *Context, key ...aorm.ID)
}

type ControllerSetuper interface {
	SetupContext(ctx *Context) (err error)
}

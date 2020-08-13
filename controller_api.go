package admin

type MainControllerIndexer interface {
	Index(context *Context)
}

type MainControllerShower interface {
	Show(context *Context)
}

type MainControllerFormCreator interface {
	New(context *Context)
}

type MainControllerCreator interface {
	Create(context *Context)
}

type MainControllerFormUpdater interface {
	Edit(context *Context)
}

type MainControllerUpdater interface {
	Update(context *Context)
}

type MainControllerDeleter interface {
	Delete(context *Context)
}

type MainControllerBulkDeleter interface {
	BulkDelete(context *Context)
}

type MainControllerSearcher interface {
	Search(context *Context)
}

type MainControllerRestorer interface {
	Restore(context *Context)
}

type MainControllerDeletedIndexer interface {
	DeletedIndex(context *Context)
}


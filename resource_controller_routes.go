package admin

import (
	"github.com/moisespsena-go/xroute"
)

func (rc *ResourceController) RegisterDefaultSingletonRouters() {
	vc := rc.ViewController
	res := rc.Resource

	if rc.Readable() {
		var readHandler = vc.Handlers.Require(VA_READ)

		readHandler.Path = P_SINGLETON_READ

		res.Router.Api(func(router xroute.Router) {
			router.Get(P_SINGLETON_READ, readHandler)
		})

		if rc.Updatable() {
			var (
				updateForm    = vc.Handlers.Require(VA_UPDATE_FORM)
				updateHandler = vc.Handlers.Require(VA_UPDATE)
			)

			updateForm.Path = P_SINGLETON_UPDATE_FORM
			updateHandler.Path = P_SINGLETON_UPDATE

			res.Router.Api(func(router xroute.Router) {
				router.Get(P_SINGLETON_UPDATE_FORM, updateForm)
				router.Put(P_SINGLETON_UPDATE, updateHandler)
			})
		}
	} else if rc.Updatable() {
		var (
			updateForm    = vc.Handlers.Require(VA_UPDATE_FORM)
			updateHandler = vc.Handlers.Require(VA_UPDATE)
		)

		updateForm.Path = P_SINGLETON_READ
		updateHandler.Path = P_SINGLETON_READ

		res.Router.Api(func(router xroute.Router) {
			router.Get(P_SINGLETON_READ, updateForm)
			router.Put(P_SINGLETON_READ, updateHandler)
		})
	} else if rc.Creatable() {
		var (
			createForm    = vc.Handlers.Require(VA_CREATE_FROM)
			createHandler = vc.Handlers.Require(VA_CREATE)
		)

		createForm.Path = P_SINGLETON_READ
		createHandler.Path = P_SINGLETON_READ

		res.Router.Get(P_SINGLETON_READ, createForm)
		res.Router.Api(func(router xroute.Router) {
			router.Post(P_SINGLETON_READ, createHandler)
		})
	}
}

func (rc *ResourceController) RegisterDefaultNormalRouters() {
	vc := rc.ViewController
	res := rc.Resource

	if rc.Creatable() {
		var (
			createForm    = vc.Handlers.Require(VA_CREATE_FROM)
			createHandler = vc.Handlers.Require(VA_CREATE)
		)

		createForm.Path = P_NEW_FORM
		createHandler.Path = P_NEW

		res.Router.Get(P_NEW_FORM, createForm)
		res.Router.Api(func(router xroute.Router) {
			router.Post(P_NEW, createHandler)
		})
	}

	if rc.Readable() {
		var readHandler = vc.Handlers.Require(VA_READ)

		readHandler.Path = P_OBJ_READ

		res.ObjectRouter.Api(func(router xroute.Router) {
			router.Get(P_OBJ_READ, readHandler)
		})
	}

	if rc.Updatable() {
		var (
			updateForm    = vc.Handlers.Require(VA_UPDATE_FORM)
			updateHandler = vc.Handlers.Require(VA_UPDATE)
		)

		updateForm.Path = P_OBJ_UPDATE_FORM
		updateHandler.Path = P_OBJ_UPDATE

		res.ObjectRouter.Api(func(router xroute.Router) {
			router.Get(P_OBJ_UPDATE_FORM, updateForm)
			router.Put(P_OBJ_UPDATE, updateHandler)
		})
	}

	if rc.Deletable() {
		res.ObjectRouter.Delete(P_OBJ_DELETE, vc.Handlers.Require(VA_DELETE))
	}

	if rc.BulkDeletable() {
		res.Router.Post(P_BULK_DELETE, vc.Handlers.Require(VA_BULK_DELETE))
	}

	if rc.Restorable() {
		res.Router.Put(P_RESTORE, vc.Handlers.Require(VA_RESTORE))
		res.Router.Get(P_DELETED_INDEX, vc.Handlers.Require(VA_DELETED_INDEX))
	}

	if rc.Indexable() {
		var indexHandler = vc.Handlers.Require(VA_INDEX)

		indexHandler.Path = P_INDEX

		res.Router.Api(func(router xroute.Router) {
			router.Get(P_INDEX, indexHandler)
		})
	}

	if rc.Searchable() {
		var searchHandler = vc.Handlers.Require(VA_SEARCH)

		searchHandler.Path = P_SEARCH

		res.Router.Api(func(router xroute.Router) {
			router.Get(P_SEARCH, searchHandler)
		})
	}
}

package admin

import (
	"fmt"

	"github.com/ecletus/roles"
)

// View Actions (VA)
const (
	VA_CREATE        = A_CREATE
	VA_CREATE_FROM   = VA_CREATE + "_form"
	VA_READ          = A_READ
	VA_UPDATE        = A_UPDATE
	VA_UPDATE_FORM   = VA_UPDATE + "_form"
	VA_DELETE        = A_DELETE
	VA_BULK_DELETE   = A_BULK_DELETE
	VA_INDEX         = A_INDEX
	VA_SEARCH        = A_SEARCH
	VA_RESTORE       = A_RESTORE
	VA_DELETED_INDEX = A_DELETED_INDEX
)

type Handlers map[string]*RouteHandler

func (handlers Handlers) Require(key string) (handler *RouteHandler) {
	var ok bool
	if handler, ok = handlers[key]; !ok {
		panic(fmt.Errorf("Handler %q does not exists", key))
	}
	return
}

type ResourceViewController struct {
	Controller *Controller
	Handlers   Handlers
}

func (c *ResourceViewController) GetHandler(name string) (handler *RouteHandler, ok bool) {
	if c.Handlers != nil {
		handler, ok = c.Handlers[name]
	}
	return
}

func (c *ResourceViewController) HasHandler(name string) (ok bool) {
	if c.Handlers != nil {
		_, ok = c.Handlers[name]
	}
	return
}

func (c *ResourceViewController) SetHandler(name string, handler *RouteHandler) *ResourceViewController {
	if c.Handlers == nil {
		c.Handlers = make(Handlers)
	}
	c.Handlers[name] = handler
	return c
}

func (vc *ResourceViewController) CreateHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_CREATE)
	return h
}

func (vc *ResourceViewController) ReadHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_READ)
	return h
}

func (vc *ResourceViewController) UpdateHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_UPDATE)
	return h
}

func (vc *ResourceViewController) DeleteHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_DELETE)
	return h
}

func (vc *ResourceViewController) BulkDeleteHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_BULK_DELETE)
	return h
}

func (vc *ResourceViewController) RestoreHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_RESTORE)
	return h
}

func (vc *ResourceViewController) DeletedIndexHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_DELETED_INDEX)
	return h
}

func (vc *ResourceViewController) IndexHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_INDEX)
	return h
}

func (vc *ResourceViewController) SearchHandler() *RouteHandler {
	h, _ := vc.GetHandler(VA_SEARCH)
	return h
}

func (vc *ResourceViewController) InitDefaultHandlers(c *ResourceController) {
	if vc.Handlers == nil {
		vc.Handlers = make(Handlers)
	}

	res := c.Resource

	if c.IsReader() {
		readConfig := &RouteConfig{PermissionMode: roles.Read, Resource: res}

		if !vc.HasHandler(VA_READ) {
			vc.Handlers[VA_READ] = NewHandler(vc.Controller.Show, readConfig).WithName(VA_READ)
		}

		if !res.Config.Singleton {
			if c.IsIndexer() && !vc.HasHandler(VA_INDEX) {
				vc.Handlers[VA_INDEX] = NewHandler(vc.Controller.Index, readConfig).WithName(VA_INDEX)
			}

			if c.IsSearcher() && !vc.HasHandler(VA_SEARCH) {
				vc.Handlers[VA_SEARCH] = NewHandler(vc.Controller.Search, &RouteConfig{
					PermissionMode: roles.Read,
					Resource: res,
				}).WithName(VA_SEARCH)
			}
		}
	}

	if c.IsUpdater() {
		updateConfig := &RouteConfig{PermissionMode: roles.Update, Resource: res}

		if !vc.HasHandler(VA_UPDATE_FORM) {
			vc.Handlers[VA_UPDATE_FORM] = NewHandler(vc.Controller.Edit, updateConfig).WithName(VA_UPDATE_FORM)
		}

		if !vc.HasHandler(VA_UPDATE) {
			vc.Handlers[VA_UPDATE] = NewHandler(vc.Controller.Update, updateConfig).WithName(VA_UPDATE)
		}
	}

	if c.IsCreator() {
		createConfig := &RouteConfig{PermissionMode: roles.Create, Resource: res}

		if !vc.HasHandler(VA_CREATE_FROM) {
			vc.Handlers[VA_CREATE_FROM] = NewHandler(vc.Controller.New, createConfig).WithName(VA_CREATE_FROM)
		}

		if !vc.HasHandler(VA_CREATE) {
			vc.Handlers[VA_CREATE] = NewHandler(vc.Controller.Create, createConfig).WithName(VA_CREATE)
		}
	}

	if c.IsDeleter() {
		if !vc.HasHandler(VA_DELETE) {
			vc.Handlers[VA_DELETE] = NewHandler(vc.Controller.Delete, &RouteConfig{
				PermissionMode: roles.Delete,
				Resource: res,
			}).WithName(VA_DELETE)
		}
	}

	if c.IsBulkDeleter() {
		if !vc.HasHandler(VA_BULK_DELETE) {
			vc.Handlers[VA_BULK_DELETE] = NewHandler(vc.Controller.BulkDelete, &RouteConfig{
				PermissionMode: roles.Delete,
				Resource: res,
			}).WithName(VA_BULK_DELETE)
		}
	}

	if c.IsRestorer() {
		if !vc.HasHandler(VA_RESTORE) {
			vc.Handlers[VA_RESTORE] = NewHandler(vc.Controller.Restore, &RouteConfig{
				PermissionMode: roles.Update,
				Resource: res,
			}).WithName(VA_RESTORE)
		}

		if !vc.HasHandler(VA_DELETED_INDEX) {
			vc.Handlers[VA_DELETED_INDEX] = NewHandler(vc.Controller.DeletedIndex, &RouteConfig{
				PermissionMode: roles.Read,
				Resource: res,
			}).WithName(VA_DELETED_INDEX)
		}
	}
}

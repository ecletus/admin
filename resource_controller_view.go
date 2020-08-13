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

type ResourceViewControllerBuilder struct {
	ResourceController *ResourceControllerBuilder
	Controller         interface{}
	Handlers           Handlers
}

func (this *ResourceViewControllerBuilder) GetHandler(name string) (handler *RouteHandler, ok bool) {
	if this.Handlers != nil {
		handler, ok = this.Handlers[name]
	}
	return
}

func (this *ResourceViewControllerBuilder) HasHandler(name string) (ok bool) {
	if this.Handlers != nil {
		_, ok = this.Handlers[name]
	}
	return
}

func (this *ResourceViewControllerBuilder) SetHandler(name string, handler *RouteHandler) *ResourceViewControllerBuilder {
	if this.Handlers == nil {
		this.Handlers = make(Handlers)
	}
	this.Handlers[name] = handler
	return this
}

func (this *ResourceViewControllerBuilder) CreateHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_CREATE)
	return h
}

func (this *ResourceViewControllerBuilder) ReadHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_READ)
	return h
}

func (this *ResourceViewControllerBuilder) UpdateHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_UPDATE)
	return h
}

func (this *ResourceViewControllerBuilder) DeleteHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_DELETE)
	return h
}

func (this *ResourceViewControllerBuilder) BulkDeleteHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_BULK_DELETE)
	return h
}

func (this *ResourceViewControllerBuilder) RestoreHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_RESTORE)
	return h
}

func (this *ResourceViewControllerBuilder) DeletedIndexHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_DELETED_INDEX)
	return h
}

func (this *ResourceViewControllerBuilder) IndexHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_INDEX)
	return h
}

func (this *ResourceViewControllerBuilder) SearchHandler() *RouteHandler {
	h, _ := this.GetHandler(VA_SEARCH)
	return h
}

func (this *ResourceViewControllerBuilder) FormCreator() (ctrl MainControllerFormCreator) {
	if this.ResourceController.IsCreator() {
		ctrl, _ = this.Controller.(MainControllerFormCreator)
	}
	return
}

func (this *ResourceViewControllerBuilder) Creator() (ctrl MainControllerCreator) {
	if this.ResourceController.IsCreator() {
		ctrl, _ = this.Controller.(MainControllerCreator)
	}
	return
}

func (this *ResourceViewControllerBuilder) Shower() (ctrl MainControllerShower) {
	if this.ResourceController.IsReader() {
		ctrl, _ = this.Controller.(MainControllerShower)
	}
	return
}

func (this *ResourceViewControllerBuilder) Indexer() (ctrl MainControllerIndexer) {
	if this.ResourceController.IsIndexer() {
		ctrl, _ = this.Controller.(MainControllerIndexer)
	}
	return
}

func (this *ResourceViewControllerBuilder) Searcher() (ctrl MainControllerSearcher) {
	if this.ResourceController.IsSearcher() {
		ctrl, _ = this.Controller.(MainControllerSearcher)
	}
	return
}

func (this *ResourceViewControllerBuilder) FormUpdater() (ctrl MainControllerFormUpdater) {
	if this.ResourceController.IsUpdater() {
		ctrl, _ = this.Controller.(MainControllerFormUpdater)
	}
	return
}

func (this *ResourceViewControllerBuilder) Updater() (ctrl MainControllerUpdater) {
	if this.ResourceController.IsUpdater() {
		ctrl, _ = this.Controller.(MainControllerUpdater)
	}
	return
}

func (this *ResourceViewControllerBuilder) Deleter() (ctrl MainControllerDeleter) {
	if this.ResourceController.IsDeleter() {
		ctrl, _ = this.Controller.(MainControllerDeleter)
	}
	return
}

func (this *ResourceViewControllerBuilder) BulkDeleter() (ctrl MainControllerBulkDeleter) {
	if this.ResourceController.IsBulkDeleter() {
		ctrl, _ = this.Controller.(MainControllerBulkDeleter)
	}
	return
}

func (this *ResourceViewControllerBuilder) Restorer() (ctrl MainControllerRestorer) {
	if this.ResourceController.IsRestorer() {
		ctrl, _ = this.Controller.(MainControllerRestorer)
	}
	return
}

func (this *ResourceViewControllerBuilder) DeletedIndexer() (ctrl MainControllerDeletedIndexer) {
	if this.ResourceController.IsRestorer() {
		ctrl, _ = this.Controller.(MainControllerDeletedIndexer)
	}
	return
}
func (this *ResourceViewControllerBuilder) IsCreator() (ok bool) {
	_, ok = this.Controller.(ControllerCreator)
	return ok
}

func (this *ResourceViewControllerBuilder) IsReader() (ok bool) {
	_, ok = this.Controller.(ControllerReader)
	return
}

func (this *ResourceViewControllerBuilder) IsUpdater() (ok bool) {
	_, ok = this.Controller.(ControllerUpdater)
	return
}

func (this *ResourceViewControllerBuilder) IsDeleter() bool {
	return this.Deleter() != nil
}

func (this *ResourceViewControllerBuilder) IsBulkDeleter() bool {
	return this.BulkDeleter() != nil
}

func (this *ResourceViewControllerBuilder) IsRestorer() bool {
	return this.Restorer() != nil
}

func (this *ResourceViewControllerBuilder) IsIndexer() bool {
	return this.Indexer() != nil
}

func (this *ResourceViewControllerBuilder) IsSearcher() bool {
	return this.Searcher() != nil
}

func (this *ResourceViewControllerBuilder) RegisterDefaultHandlers() {
	if this.Handlers == nil {
		this.Handlers = make(Handlers)
	}
	var (
		rc  = this.ResourceController
		res = this.ResourceController.Resource
	)

	if rc.IsCreator() {
		createConfig := &RouteConfig{PermissionMode: roles.Create, Resource: res}

		if vc := this.FormCreator(); vc != nil && !this.HasHandler(VA_CREATE_FROM) {
			this.Handlers[VA_CREATE_FROM] = NewHandler(vc.New, createConfig).WithName(VA_CREATE_FROM)
		}

		if vc := this.Creator(); vc != nil && !this.HasHandler(VA_CREATE) {
			this.Handlers[VA_CREATE] = NewHandler(vc.Create, createConfig).WithName(VA_CREATE)
		}
	}

	if rc.IsReader() {
		readConfig := &RouteConfig{PermissionMode: roles.Read, Resource: res}

		if vc := this.Shower(); vc != nil && !this.HasHandler(VA_READ) {
			this.Handlers[VA_READ] = NewHandler(vc.Show, readConfig).WithName(VA_READ)
		}
	}

	if rc.IsIndexer() {
		readConfig := &RouteConfig{PermissionMode: roles.Read, Resource: res}
		if vc := this.Indexer(); vc != nil && !this.HasHandler(VA_INDEX) {
			this.Handlers[VA_INDEX] = NewHandler(vc.Index, readConfig).WithName(VA_INDEX)
		}

		if vc := this.Searcher(); vc != nil && !this.HasHandler(VA_SEARCH) {
			this.Handlers[VA_SEARCH] = NewHandler(vc.Search, &RouteConfig{
				PermissionMode: roles.Read,
				Resource:       res,
			}).WithName(VA_SEARCH)
		}
	}

	if rc.IsUpdater() {
		updateConfig := &RouteConfig{PermissionMode: roles.Update, Resource: res}

		if vc := this.FormUpdater(); vc != nil && !this.HasHandler(VA_UPDATE_FORM) {
			this.Handlers[VA_UPDATE_FORM] = NewHandler(vc.Edit, updateConfig).WithName(VA_UPDATE_FORM)
		}

		if vc := this.Updater(); vc != nil && !this.HasHandler(VA_UPDATE) {
			this.Handlers[VA_UPDATE] = NewHandler(vc.Update, updateConfig).WithName(VA_UPDATE)
		}
	}

	if rc.IsDeleter() {
		if vc := this.Deleter(); vc != nil && !this.HasHandler(VA_DELETE) {
			this.Handlers[VA_DELETE] = NewHandler(vc.Delete, &RouteConfig{
				PermissionMode: roles.Delete,
				Resource:       res,
			}).WithName(VA_DELETE)
		}
	}

	if rc.IsBulkDeleter() {
		if vc := this.BulkDeleter(); vc != nil && !this.HasHandler(VA_BULK_DELETE) {
			this.Handlers[VA_BULK_DELETE] = NewHandler(vc.BulkDelete, &RouteConfig{
				PermissionMode: roles.Delete,
				Resource:       res,
			}).WithName(VA_BULK_DELETE)
		}
	}

	if rc.IsRestorer() {
		if vc := this.Restorer(); vc != nil && !this.HasHandler(VA_RESTORE) {
			this.Handlers[VA_RESTORE] = NewHandler(vc.Restore, &RouteConfig{
				PermissionMode: roles.Update,
				Resource:       res,
			}).WithName(VA_RESTORE)
		}

		if vc := this.DeletedIndexer(); vc != nil && !this.HasHandler(VA_DELETED_INDEX) {
			this.Handlers[VA_DELETED_INDEX] = NewHandler(vc.DeletedIndex, &RouteConfig{
				PermissionMode: roles.Read,
				Resource:       res,
			}).WithName(VA_DELETED_INDEX)
		}
	}
}

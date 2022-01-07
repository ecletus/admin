package admin

import "github.com/ecletus/roles"

var defaultControllerActions = []string{
	A_CREATE,
	A_READ,
	A_UPDATE,
	A_DELETE,
	A_INDEX,
	A_SEARCH,
	A_BULK_DELETE,
	A_RESTORE,
	A_DELETED_INDEX,
}

type ResourceControllerBuilder struct {
	Resource       *Resource
	Controller     interface{}
	ViewController *ResourceViewControllerBuilder
	defaultActions map[string]bool
}

func (this *ResourceControllerBuilder) AppendDefaultActions(actions ...string) {
	if this.defaultActions == nil {
		this.defaultActions = map[string]bool{}
	}

	for _, action := range actions {
		this.defaultActions[action] = true
	}
}

func (this *ResourceControllerBuilder) DefaultActions() (actions []string) {
	if this.defaultActions == nil || len(this.defaultActions) == 0 {
		this.defaultActions = map[string]bool{}
		for _, action := range defaultControllerActions {
			this.defaultActions[action] = true
		}
	} else {
		for action := range this.defaultActions {
			actions = append(actions, action)
		}
	}
	return
}

func (this *ResourceControllerBuilder) HasDefaultAction(name string) (ok bool) {
	if this.defaultActions == nil || len(this.defaultActions) == 0 {
		this.DefaultActions()
	}
	_, ok = this.defaultActions[name]
	return
}

func (this *ResourceControllerBuilder) RegisterDefaultRouters() {
	if this.Resource.Config.Wizard != nil {
		this.RegisterWizardRouters()
	} else if this.Resource.Config.Singleton {
		this.RegisterDefaultSingletonRouters()
	} else {
		this.RegisterDefaultNormalRouters()
	}
}

func (this *ResourceControllerBuilder) IsCreator() (ok bool) {
	_, ok = this.Controller.(ControllerCreator)
	return ok
}

func (this *ResourceControllerBuilder) IsReader() (ok bool) {
	_, ok = this.Controller.(ControllerReader)
	return
}

func (this *ResourceControllerBuilder) IsUpdater() (ok bool) {
	_, ok = this.Controller.(ControllerUpdater)
	return
}

func (this *ResourceControllerBuilder) IsDeleter() (ok bool) {
	_, ok = this.Controller.(ControllerDeleter)
	return ok && !this.Resource.Config.Singleton
}

func (this *ResourceControllerBuilder) IsBulkDeleter() (ok bool) {
	_, ok = this.Controller.(ControllerBulkDeleter)
	return ok && this.IsDeleter()
}

func (this *ResourceControllerBuilder) IsRestorer() (ok bool) {
	_, ok = this.Controller.(ControllerRestorer)
	return ok && this.IsDeleter()
}

func (this *ResourceControllerBuilder) IsIndexer() (ok bool) {
	_, ok = this.Controller.(ControllerIndex)
	return ok && !this.Resource.Config.Singleton
}

func (this *ResourceControllerBuilder) IsSearcher() (ok bool) {
	_, ok = this.Controller.(ControllerSearcher)
	return
}

func (this *ResourceControllerBuilder) Creatable() (ok bool) {
	return this.IsCreator() && this.HasDefaultAction(A_CREATE)
}

func (this *ResourceControllerBuilder) Readable() (ok bool) {
	return this.IsReader() && this.HasDefaultAction(A_READ)
}

func (this *ResourceControllerBuilder) Updatable() (ok bool) {
	return this.IsUpdater() && this.HasDefaultAction(A_READ)
}

func (this *ResourceControllerBuilder) Deletable() (ok bool) {
	return this.IsDeleter() && this.HasDefaultAction(A_DELETE)
}

func (this *ResourceControllerBuilder) BulkDeletable() (ok bool) {
	return this.IsBulkDeleter() && this.HasDefaultAction(A_BULK_DELETE)
}

func (this *ResourceControllerBuilder) Restorable() (ok bool) {
	return this.IsRestorer() &&
		this.HasDefaultAction(A_RESTORE) &&
		this.HasDefaultAction(A_DELETED_INDEX)
}

func (this *ResourceControllerBuilder) Indexable() (ok bool) {
	return this.IsIndexer() && this.HasDefaultAction(A_INDEX)
}

func (this *ResourceControllerBuilder) Searchable() (ok bool) {
	return this.IsSearcher() && this.HasDefaultAction(A_SEARCH)
}

func (this *ResourceControllerBuilder) HasMode(mode roles.PermissionMode) (ok *bool) {
	var b bool
	switch mode {
	case roles.Create:
		b = len(this.Resource.createWizards) > 0 || this.IsCreator()
	case roles.Read:
		b = this.IsReader()
	case roles.Update:
		b = this.IsUpdater()
	case roles.Delete:
		b = this.IsDeleter()
	default:
		return nil
	}
	return &b
}

func (this *ResourceControllerBuilder) HasModes(mode roles.PermissionMode, modeN ...roles.PermissionMode) (ok *bool) {
	if ok = this.HasMode(mode); ok != nil {
		return
	}
	for _, mode := range modeN {
		if ok = this.HasMode(mode); ok != nil {
			return
		}
	}
	return
}

package admin

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

type ResourceController struct {
	Resource       *Resource
	Controller     interface{}
	ViewController *ResourceViewController
	defaultActions map[string]bool
}

func (c *ResourceController) AppendDefaultActions(actions ...string) {
	if c.defaultActions == nil {
		c.defaultActions = map[string]bool{}
	}

	for _, action := range actions {
		c.defaultActions[action] = true
	}
}

func (c *ResourceController) DefaultActions() (actions []string) {
	if c.defaultActions == nil || len(c.defaultActions) == 0 {
		c.defaultActions = map[string]bool{}
		for _, action := range defaultControllerActions {
			c.defaultActions[action] = true
		}
	} else {
		for action := range c.defaultActions {
			actions = append(actions, action)
		}
	}
	return
}

func (c *ResourceController) HasDefaultAction(name string) (ok bool) {
	if c.defaultActions == nil || len(c.defaultActions) == 0 {
		c.DefaultActions()
	}
	_, ok = c.defaultActions[name]
	return
}

func (rc *ResourceController) RegisterDefaultRouters() {
	rc.ViewController.InitDefaultHandlers(rc)
	if rc.Resource.Config.Singleton {
		rc.RegisterDefaultSingletonRouters()
	} else {
		rc.RegisterDefaultNormalRouters()
	}
}

func (rc *ResourceController) IsCreator() (ok bool) {
	_, ok = rc.Controller.(ControllerCreator)
	return ok
}

func (rc *ResourceController) IsReader() (ok bool) {
	_, ok = rc.Controller.(ControllerReader)
	return
}

func (rc *ResourceController) IsUpdater() (ok bool) {
	_, ok = rc.Controller.(ControllerUpdater)
	return
}

func (rc *ResourceController) IsDeleter() (ok bool) {
	_, ok = rc.Controller.(ControllerDeleter)
	return ok && !rc.Resource.Config.Singleton
}

func (rc *ResourceController) IsBulkDeleter() (ok bool) {
	_, ok = rc.Controller.(ControllerBulkDeleter)
	return ok && rc.IsDeleter()
}

func (rc *ResourceController) IsRestorer() (ok bool) {
	_, ok = rc.Controller.(ControllerRestorer)
	return ok && rc.IsDeleter()
}

func (rc *ResourceController) IsIndexer() (ok bool) {
	_, ok = rc.Controller.(ControllerIndex)
	return ok
}

func (rc *ResourceController) IsSearcher() (ok bool) {
	_, ok = rc.Controller.(ControllerSearcher)
	return
}

func (rc *ResourceController) Creatable() (ok bool) {
	return rc.IsCreator() && rc.HasDefaultAction(A_CREATE)
}

func (rc *ResourceController) Readable() (ok bool) {
	return rc.IsReader() && rc.HasDefaultAction(A_READ)
}

func (rc *ResourceController) Updatable() (ok bool) {
	return rc.IsUpdater() && rc.HasDefaultAction(A_READ)
}

func (rc *ResourceController) Deletable() (ok bool) {
	return rc.IsDeleter() && rc.HasDefaultAction(A_DELETE)
}

func (rc *ResourceController) BulkDeletable() (ok bool) {
	return rc.IsBulkDeleter() && rc.HasDefaultAction(A_BULK_DELETE)
}

func (rc *ResourceController) Restorable() (ok bool) {
	return rc.IsRestorer() &&
		rc.HasDefaultAction(A_RESTORE) &&
		rc.HasDefaultAction(A_DELETED_INDEX)
}

func (rc *ResourceController) Indexable() (ok bool) {
	return rc.IsIndexer() && rc.HasDefaultAction(A_INDEX)
}

func (rc *ResourceController) Searchable() (ok bool) {
	return rc.IsSearcher() && rc.HasDefaultAction(A_SEARCH)
}

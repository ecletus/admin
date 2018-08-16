package admin

type ResourceAccess interface {
	GetResources() []*Resource
	WalkResources(f func(res *Resource) bool) bool
	GetResourceByID(id string) *Resource
	GetParentResourceByID(id string) *Resource
	GetOrParentResourceByID(id string) *Resource
	GetResourceByParam(name string) *Resource
}

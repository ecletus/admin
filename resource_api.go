package admin

type ResourceSetuper interface {
	AdminResourceSetup(res *Resource, defaultSetup func())
}

type ResourceMetaHasInitialValuer interface {
	AdminIsDefaultValue(meta *Meta) bool
}
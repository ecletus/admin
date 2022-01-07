package admin

type (
	ResourceSetuper interface {
		AdminResourceSetup(res *Resource, defaultSetup func())
	}

	ResourceTagsGetter interface {
		AdminGetResourceTags(res *Resource) (tags *ResourceTags)
	}

	ResourceMetaHasInitialValuer interface {
		AdminIsDefaultValue(meta *Meta) bool
	}
)

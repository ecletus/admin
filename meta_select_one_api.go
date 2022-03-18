package admin

type SelectEqualer interface {
	SelectOneItemEq(a interface{}, b string) bool
}

type SelectCollectionProvider interface {
	GetCollection() [][]string
}

type SelectCollectionContextProvider interface {
	GetCollection(ctx *Context) [][]string
}

type SelectCollectionAsyncResource interface {
	GetAsyncResource(cfg *SelectOneConfig) *Resource
}

type SelectCollectionRecordContextProvider interface {
	GetCollection(record interface{}, ctx *Context) [][]string
}

type SelectCollectionRecordContextProviderFunc func(record interface{}, ctx *Context) [][]string

func (f SelectCollectionRecordContextProviderFunc) GetCollection(record interface{}, ctx *Context) [][]string {
	return f(record, ctx)
}

type SelectCollectionMethodFactorer interface {
	Factory(meta *Meta) SelectCollectionRecordContextProvider
}

type SelectCollectionMethodFactorerFunc func(meta *Meta) SelectCollectionRecordContextProvider

func (f SelectCollectionMethodFactorerFunc) Factory(meta *Meta) SelectCollectionRecordContextProvider {
	return f(meta)
}

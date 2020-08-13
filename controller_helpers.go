package admin

type CheckLoaderForUpdater interface {
	IsLoadForUpdate(ctx *Context) bool
}

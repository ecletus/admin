// +build assetfs_bindata

package admin

import (
	"strings"

	assetfs "github.com/moisespsena/go-assetfs"
)

func (context *Context) getAsset(fs assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, err error) {
	key := "assets:" + strings.Join(layouts, "!")
	if value, ok := context.Admin.Cache.Load(key); ok {
		return value.(assetfs.AssetInterface), nil
	}
	if asset, err = context.findAsset(fs, layouts...); err == nil {
		context.Admin.Cache.Store(key, asset)
	}
	return
}

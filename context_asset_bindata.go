// +build assetfs_bindata

package admin

import (
	"strings"

	assetfs "github.com/moisespsena-go/assetfs"
)

func (this *Context) getAsset(fs []assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, err error) {
	key := "assets:" + strings.Join(layouts, "!")
	if value, ok := this.Admin.Cache.Load(key); ok {
		return value.(assetfs.AssetInterface), nil
	}
	var afs assetfs.Interface
	if asset, afs, err = this.findAsset(fs, layouts...); err == nil {
		if _, ok := afs.(assetfsCache); ok {
			this.Admin.Cache.Store(key, asset)
		}
	}
	return
}

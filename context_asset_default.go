// +build !assetfs_bindata

package admin

import "github.com/moisespsena/go-assetfs"

func (context *Context) getAsset(fs assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, err error) {
	return context.findAsset(fs, layouts...)
}

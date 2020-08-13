// +build !assetfs_bindata

package admin

import "github.com/moisespsena-go/assetfs"

func (this *Context) getAsset(fs []assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, err error) {
	asset, _, err = this.findAsset(fs, layouts...)
	return
}

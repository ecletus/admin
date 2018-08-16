package admin

import (
	"os"
	"github.com/moisespsena/go-assetfs"
)

var (
	root, _         = os.Getwd()
	globalViewPaths []string
	globalAssetFSes []assetfs.Interface
)

func init() {
	if path := os.Getenv("WEB_ROOT"); path != "" {
		root = path
	}
}

// RegisterViewPath register view path for all assetfs
func RegisterViewPath(pth string) {
	globalViewPaths = append(globalViewPaths, pth)

	for _, assetFS := range globalAssetFSes {
		if assetFS.RegisterPath(pth) != nil {
			return
		}
	}
}

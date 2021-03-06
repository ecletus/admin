package admin

import (
	"errors"

	"github.com/moisespsena-go/assetfs"
	"github.com/ecletus/core/resource"
)

// CollectionEditConfig meta configuration used for collection edit
type CollectionEditConfig struct {
	Template string
	Max      uint
	metaConfig
}

// GetTemplate get template for collection edit
func (collectionEditConfig CollectionEditConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && collectionEditConfig.Template != "" {
		return context.Asset(collectionEditConfig.Template)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure collection edit meta
func (collectionEditConfig *CollectionEditConfig) ConfigureQorMeta(metaor resource.Metaor) {
}

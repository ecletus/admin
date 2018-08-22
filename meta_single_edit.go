package admin

import (
	"errors"
	"github.com/moisespsena/go-assetfs"
	"github.com/aghape/core/resource"
)

// SingleEditConfig meta configuration used for single edit
type SingleEditConfig struct {
	Template string
	metaConfig
	ExcludeEditAttrs []string
	ExcludeNewAttrs []string
	ExcludeShowAttrs []string
	AfterParseMetaValues func(record interface{}, context *Context)
}

func (s *SingleEditConfig) EditSections(res *Resource) []*Section {
	var attrs []interface{}
	for _, a := range res.EditAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range s.ExcludeEditAttrs {
		attrs = append(attrs, "-" + a)
	}
	return res.SectionsList(attrs...)
}

func (s *SingleEditConfig) NewSections(res *Resource) []*Section {
	var attrs []interface{}
	for _, a := range res.NewAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range s.ExcludeNewAttrs {
		attrs = append(attrs, "-" + a)
	}
	return res.SectionsList(attrs...)
}

func (s *SingleEditConfig) ShowSections(res *Resource) []*Section {
	var attrs []interface{}
	for _, a := range res.ShowAttrs() {
		attrs = append(attrs, a)
	}
	for _, a := range s.ExcludeShowAttrs {
		attrs = append(attrs, "-" + a)
	}
	return res.SectionsList(attrs...)
}

// GetTemplate get template for single edit
func (singleEditConfig *SingleEditConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && singleEditConfig.Template != "" {
		return context.Asset(singleEditConfig.Template)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure single edit meta
func (singleEditConfig *SingleEditConfig) ConfigureQorMeta(metaor resource.Metaor) {
}

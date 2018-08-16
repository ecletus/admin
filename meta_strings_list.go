package admin

import (
	"github.com/aghape/aghape/resource"
)

type StringsListConfig struct {
	Field string
	metaConfig
}

// ConfigureQorMeta configure rich editor meta
func (s *StringsListConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		meta.Type = "strings_list"
	}
}

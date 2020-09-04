package admin

import (
	"fmt"
)

var metaConfigorMaps = map[string]func(*Meta){
	"date": func(meta *Meta) {
		if meta.Config == nil {
			cfg := &DateConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	},

	"time": func(meta *Meta) {
		if meta.Config == nil {
			cfg := &TimeConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	},

	"select_one": func(meta *Meta) {
		if metaConfig, ok := meta.Config.(*SelectOneConfig); !ok || metaConfig == nil {
			meta.Config = &SelectOneConfig{Collection: meta.Collection}
			meta.Config.ConfigureQorMeta(meta)
		} else if meta.Collection != nil {
			metaConfig.Collection = meta.Collection
			meta.Config.ConfigureQorMeta(meta)
		}
	},

	"select_many": func(meta *Meta) {
		if metaConfig, ok := meta.Config.(*SelectManyConfig); !ok || metaConfig == nil {
			meta.Config = &SelectManyConfig{Collection: meta.Collection}
			meta.Config.ConfigureQorMeta(meta)
		} else if meta.Collection != nil {
			metaConfig.Collection = meta.Collection
			meta.Config.ConfigureQorMeta(meta)
		}
	},

	"single_edit": func(meta *Meta) {
		if _, ok := meta.Config.(*SingleEditConfig); !ok || meta.Config == nil {
			meta.Config = &SingleEditConfig{}
			meta.Config.ConfigureQorMeta(meta)
		}
	},

	"collection_edit": func(meta *Meta) {
		if _, ok := meta.Config.(*CollectionEditConfig); !ok || meta.Config == nil {
			meta.Config = &CollectionEditConfig{}
			meta.Config.ConfigureQorMeta(meta)
		}
	},
}

func RegisterMetaConfigor(name string, configor func(meta *Meta)) {
	if _, ok := metaConfigorMaps[name]; ok {
		panic(fmt.Errorf("duplicate meta configor %q", name))
	}
	metaConfigorMaps[name] = configor
}

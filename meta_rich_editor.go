package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
)

type RichEditorConfig struct {
	AssetManager *Resource
	Plugins      []RedactorPlugin
	Settings     map[string]interface{}
	metaConfig
}

type RedactorPlugin struct {
	Name   string
	Source string
}

// ConfigureQorMeta configure rich editor meta
func (this *RichEditorConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		meta.Type = "rich_editor"

		// Compatible with old rich editor setting
		if meta.Resource != nil {
			this.AssetManager = meta.Resource
			meta.Resource = nil
		}

		setter := meta.GetSetter()
		meta.SetSetter(func(resource interface{}, metaValue *resource.MetaValue, context *core.Context) error {
			metaValue.Value = utils.HTMLSanitizer.Sanitize(utils.ToString(metaValue.Value))
			return setter(resource, metaValue, context)
		})

		if this.Settings == nil {
			this.Settings = map[string]interface{}{}
		}

		plugins := []string{"source"}
		for _, plugin := range this.Plugins {
			plugins = append(plugins, plugin.Name)
		}
		this.Settings["plugins"] = plugins
	}
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &RichEditorConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("rich_editor", cfg)
}

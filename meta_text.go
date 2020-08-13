package admin

import (
	"github.com/ecletus/core/resource"
)

type TextWordBreak uint8

func (this TextWordBreak) Style() string {
	switch this {
	case WordBreakAll:
		return "break-all"
	case WordBreakHyphens:
		return "keep-all"
	default:
		return ""
	}
}

const (
	WordBreakNone TextWordBreak = iota
	WordBreakHyphens
	WordBreakAll
)

type TextConfig struct {
	WordBreak TextWordBreak
	Copy      bool
	StringConfig
}

func (this *TextConfig) ConfigureQorMeta(metaor resource.Metaor) {
	this.StringConfig.ConfigureQorMeta(metaor)
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &TextConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		} else {
			meta.Config.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("text", cfg)
}

package admin

import "github.com/ecletus/core"

const (
	FlagFormCreateContinueEditingDisabled FormFlagBits = 1 << iota
	FlagFormEditContinueEditingDisabled
)

type FormFlagBits uint8

func (b FormFlagBits) Set(flag FormFlagBits) FormFlagBits    { return b | flag }
func (b FormFlagBits) Clear(flag FormFlagBits) FormFlagBits  { return b &^ flag }
func (b FormFlagBits) Toggle(flag FormFlagBits) FormFlagBits { return b ^ flag }
func (b FormFlagBits) Has(flag FormFlagBits) bool            { return b&flag != 0 }

func OptFormFlags(flags FormFlagBits) core.Option {
	return core.OptionFunc(func(configor core.Configor) {
		configor.ConfigSet("form:flags", flags)
	})
}
func GetOptFormFlags(configor core.Configor) (bits FormFlagBits) {
	if v, ok := configor.ConfigGet("form:flags"); ok {
		return v.(FormFlagBits)
	}
	return
}

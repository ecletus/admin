package admin

const (
	_ ContextRenderFlag = 1 << iota
	CtxRenderMetaSingleMode
	CtxRenderEncode
	CtxRenderCSV
	LazyLoadDisabled
)

type ContextRenderFlag uint16

func (b ContextRenderFlag) Set(flag ContextRenderFlag) ContextRenderFlag    { return b | flag }
func (b ContextRenderFlag) Clear(flag ContextRenderFlag) ContextRenderFlag  { return b &^ flag }
func (b ContextRenderFlag) Toggle(flag ContextRenderFlag) ContextRenderFlag { return b ^ flag }
func (b ContextRenderFlag) Has(flag ...ContextRenderFlag) bool {
	for _, flag := range flag {
		if (b & flag) != 0 {
			return true
		}
	}
	return false
}

func (b ContextRenderFlag) HasName(name ...string) bool {
	for _, n := range name {
		switch n {
		case "single_mode":
			if b.Has(CtxRenderMetaSingleMode) {
				return true
			}
		}
	}
	return false
}

func (b ContextRenderFlag) Names() (names []string) {
	if b.Has(CtxRenderMetaSingleMode) {
		names = append(names, "single_mode")
	}
	return
}

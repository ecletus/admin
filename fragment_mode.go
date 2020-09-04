package admin

const (
	FragmentFormCast FragmentMode = 1 << iota
	FragmentFormInline
)

type FragmentMode uint8

func (b FragmentMode) Set(flag FragmentMode) FragmentMode    { return b | flag }
func (b FragmentMode) Clear(flag FragmentMode) FragmentMode  { return b &^ flag }
func (b FragmentMode) Toggle(flag FragmentMode) FragmentMode { return b ^ flag }
func (b FragmentMode) Has(flag FragmentMode) bool            { return b&flag != 0 }
func (b FragmentMode) Cast() bool                            { return b.Has(FragmentFormCast) }
func (b FragmentMode) Inline() bool                          { return b.Has(FragmentFormInline) }

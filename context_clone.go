package admin

func (this *Context) Clone() *Context {
	clone := *this
	clone.Context = this.Context.Clone()
	if this.Searcher != nil {
		searcher := *this.Searcher
		clone.Searcher = &searcher
		clone.Searcher.Context = &clone
	}
	return &clone
}

func (this *Context) CloneErr(f ...func(ctx *Context)) error {
	clone := this.Clone()
	for _, f := range f {
		f(clone)
		if clone.HasError() {
			break
		}
	}
	return clone.Err()
}

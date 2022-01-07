package admin

import "github.com/ecletus/core"

type SortableAttrs struct {
	Parent *SortableAttrs
	Names  []string
}

func (this *SortableAttrs) Add(name ...string) {
	this.Names = append(this.Names, name...)
}

func (this *SortableAttrs) Has(name string) bool {
	e := this
	for e != nil {
		if core.StringSlice(e.Names).Has(name) {
			return true
		}
		e = e.Parent
	}
	return false
}

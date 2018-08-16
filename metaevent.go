package admin

import (
	"github.com/moisespsena/go-edis"
	"github.com/aghape/aghape"
)

const (
	// Meta events
	E_META_VALUE           = "value"
	E_META_FORMATTED_VALUE = "formattedValue"
)

type MetaEvent struct {
	edis.EventInterface
	Meta     *Meta
	Resource *Resource
	Context  *qor.Context
}

type MetaRecordeEvent struct {
	MetaEvent
	Recorde interface{}
}

type MetaValueEvent struct {
	MetaRecordeEvent
	valuer              MetaValuer
	Value               interface{}
	originalValue       interface{}
	originalValueCalled bool
}

func (mve *MetaValueEvent) OriginalValue() interface{} {
	if !mve.originalValueCalled {
		mve.originalValue = mve.valuer(mve.Recorde, mve.Context)
		mve.originalValueCalled = true
	}
	return mve.originalValue
}

func (mve *MetaValueEvent) GetOrOriginalValue() interface{} {
	if mve.Value == nil {
		mve.Value = mve.OriginalValue()
	}
	return mve.OriginalValue()
}

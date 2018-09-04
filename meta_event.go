package admin

import (
	"github.com/aghape/core"
	"github.com/moisespsena/go-edis"
)

const (
	// Meta events
	E_META_VALUE                = "value"
	E_META_FORMATTED_VALUE      = "formattedValue"
	E_META_POST_FORMATTED_VALUE = "postFormattedValue"
	E_META_DEFAULT_VALUE        = "defaultValue"
)

type MetaEvent struct {
	edis.EventInterface
	Meta     *Meta
	Resource *Resource
	Context  *core.Context
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

func OnMetaValue(meta *Meta, eventName string, cb func(e *MetaValueEvent)) {
	meta.On(eventName, func(e edis.EventInterface) {
		cb(e.(*MetaValueEvent))
	})
}

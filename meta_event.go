package admin

import (
	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/moisespsena/go-edis"
)

const (
	// Meta events
	E_META_SET                  = "set"
	E_META_CHANGED              = "changed"
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

type MetaValueChangedEvent struct {
	MetaRecordeEvent
	Old    interface{}
	valuer MetaValuer
}

func (mve *MetaValueChangedEvent) Value() interface{} {
	return mve.valuer(mve.Recorde, mve.Context)
}

type MetaValuerEvent struct {
	MetaRecordeEvent
	valuer              MetaValuer
	Value               interface{}
	originalValue       interface{}
	originalValueCalled bool
}

func (mve *MetaValuerEvent) Set(value interface{}) {
	mve.Value = value
}

func (mve *MetaValuerEvent) OriginalValue() interface{} {
	if !mve.originalValueCalled {
		mve.originalValue = mve.valuer(mve.Recorde, mve.Context)
		mve.originalValueCalled = true
	}
	return mve.originalValue
}

func (mve *MetaValuerEvent) GetOrOriginalValue() interface{} {
	if mve.Value == nil {
		mve.Value = mve.OriginalValue()
	}
	return mve.OriginalValue()
}

func OnMetaValue(meta *Meta, eventName string, cb func(e *MetaValuerEvent)) *Meta {
	meta.On(eventName, func(e edis.EventInterface) {
		cb(e.(*MetaValuerEvent))
	})
	return meta
}

func (meta *Meta) OnValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(meta, E_META_VALUE, cb)
}

func (meta *Meta) OnFormattedValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(meta, E_META_FORMATTED_VALUE, cb)
}

func (meta *Meta) OnPostFormattedValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(meta, E_META_POST_FORMATTED_VALUE, cb)
}

type MetaSetEvent struct {
	MetaRecordeEvent
	Setter             MetaSetter
	MetaValue          *resource.MetaValue
	currentValue       interface{}
	currentValueCalled bool
}

func (mse *MetaSetEvent) CurrentValue() interface{} {
	if !mse.currentValueCalled {
		mse.currentValueCalled = true
		mse.currentValue = mse.Meta.Value(mse.Context, mse.Recorde)
	}
	return mse.currentValue
}

func (mse *MetaSetEvent) SetValue(value interface{}) {
	mse.MetaValue.Value = value
}

func (mse *MetaSetEvent) Value() interface{} {
	return mse.MetaValue.Value
}

func (meta *Meta) OnSet(cb func(e *MetaSetEvent)) *Meta {
	meta.On(E_META_SET, func(e edis.EventInterface) {
		cb(e.(*MetaSetEvent))
	})
	return meta
}

func (meta *Meta) OnChanged(cb func(e *MetaValueChangedEvent)) *Meta {
	meta.On(E_META_CHANGED, func(e edis.EventInterface) {
		cb(e.(*MetaValueChangedEvent))
	})
	return meta
}

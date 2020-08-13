package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/edis"
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

func (this *Meta) OnValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(this, E_META_VALUE, cb)
}

func (this *Meta) OnFormattedValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(this, E_META_FORMATTED_VALUE, cb)
}

func (this *Meta) OnPostFormattedValue(cb func(e *MetaValuerEvent)) *Meta {
	return OnMetaValue(this, E_META_POST_FORMATTED_VALUE, cb)
}

type MetaSetEvent struct {
	MetaRecordeEvent
	Setter             MetaSetter
	MetaValue          *resource.MetaValue
	currentValue       interface{}
	currentValueCalled bool
}

func (this *MetaSetEvent) CurrentValue() interface{} {
	if !this.currentValueCalled {
		this.currentValueCalled = true
		this.currentValue = this.Meta.Value(this.Context, this.Recorde)
	}
	return this.currentValue
}

func (this *MetaSetEvent) SetValue(value interface{}) {
	this.MetaValue.Value = value
}

func (this *MetaSetEvent) Value() interface{} {
	return this.MetaValue.Value
}

func (this *MetaSetEvent) FirstStringValue() (value string) {
	if v := this.Value(); v != nil {
		return v.([]string)[0]
	}
	return
}

func (this *Meta) OnSet(cb func(e *MetaSetEvent) error) *Meta {
	this.On(E_META_SET, func(e edis.EventInterface) error {
		return cb(e.(*MetaSetEvent))
	})
	return this
}

func (this *Meta) OnChanged(cb func(e *MetaValueChangedEvent) error) *Meta {
	this.On(E_META_CHANGED, func(e edis.EventInterface) error {
		return cb(e.(*MetaValueChangedEvent))
	})
	return this
}

package admin

import (
	"github.com/aghape/core"
	"github.com/aghape/core/utils"
	"github.com/aghape/plug"
	"github.com/moisespsena/go-edis"
)

const (
	E_RESOURCE_ADDED = "resourceAdded"
	E_DONE           = "done"
)

type ResourceEvent struct {
	edis.EventInterface
	Resource   *Resource
	Registered bool
}

type AdminEvent struct {
	edis.EventInterface
	Admin *Admin
}

func (admin *Admin) TriggerResource(e *ResourceEvent) (err error) {
	err = admin.Trigger(e)
	if err == nil {
		defer e.With(e.Name() + ":" + e.Resource.UID)()
		admin.Trigger(e)
	}
	return
}

func (admin *Admin) TriggerDone(e *AdminEvent) error {
	return admin.Trigger(e)
}

func (admin *Admin) OnResourceAdded(cb func(e *ResourceEvent)) error {
	return admin.OnE(E_RESOURCE_ADDED, func(e edis.EventInterface) {
		cb(e.(*ResourceEvent))
	})
}

func (admin *Admin) OnResourceAddedE(cb func(e *ResourceEvent) error) error {
	return admin.OnE(E_RESOURCE_ADDED, func(e edis.EventInterface) error {
		return cb(e.(*ResourceEvent))
	})
}

func (admin *Admin) OnResourceValueAdded(value interface{}, cb func(e *ResourceEvent)) error {
	uid := utils.TypeId(value)
	if res := admin.ResourcesByUID[uid]; res != nil {
		admin.triggerResourceAdded(res)
		return nil
	}

	return admin.OnE(E_RESOURCE_ADDED+":"+utils.TypeId(value), func(e edis.EventInterface) {
		cb(e.(*ResourceEvent))
	})
}

func (admin *Admin) OnResourceValueAddedE(value interface{}, cb func(e *ResourceEvent) error) error {
	uid := utils.TypeId(value)
	if res := admin.ResourcesByUID[uid]; res != nil {
		admin.triggerResourceAdded(res)
		return nil
	}

	return admin.OnE(E_RESOURCE_ADDED+":"+utils.TypeId(value), func(e edis.EventInterface) error {
		return cb(e.(*ResourceEvent))
	})
}

func (admin *Admin) OnDone(cb func(e *AdminEvent)) error {
	return admin.OnE(E_DONE, func(e edis.EventInterface) {
		cb(e.(*AdminEvent))
	})
}

func (admin *Admin) OnDoneE(cb func(e *AdminEvent) error) error {
	return admin.OnE(E_DONE, func(e edis.EventInterface) error {
		return cb(e.(*AdminEvent))
	})
}

type RecordeEvent struct {
	plug.EventInterface
	Resource *Resource
	Context  *core.Context
	Recorde  interface{}
}

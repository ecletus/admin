package admin

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ecletus/plug"
	"github.com/moisespsena-go/edis"
	"github.com/moisespsena-go/logging"
	"github.com/pkg/errors"

	"github.com/ecletus/core"
)

const (
	E_RESOURCE_ADDED  = "resourceAdded"
	E_RESOURCES_ADDED = "resourcesAdded"
	E_DONE            = "done"
)

type ResourceEvent struct {
	edis.EventInterface
	Resource   *Resource
	Resources  []*Resource
	Registered bool
}

type AdminEvent struct {
	edis.EventInterface
	Admin *Admin
}

func (this *Admin) TriggerResource(e *ResourceEvent) (err error) {
	if err = this.Trigger(e); err != nil {
		return errors.Wrapf(err, "trigger %q", e.Name())
	}
	func() {
		defer e.With(e.Name() + ":" + e.Resource.UID)()
		err = errors.Wrapf(this.Trigger(e), "trigger %q", e.Name())
	}()
	if err != nil {
		return
	}
	defer e.With(e.Name() + ":" + e.Resource.FullID())()
	return errors.Wrapf(this.Trigger(e), "trigger %q", e.Name())
}

func (this *Admin) TriggerDone(e *AdminEvent) error {
	return this.Trigger(e)
}

func (this *Admin) triggerResourceAdded(res *Resource, cb ...func(e *ResourceEvent) error) (err error) {
	if callbacks, ok := this.onResourceTypeAdded[res.ModelStruct.Type]; ok {
		for _, cb := range callbacks {
			cb(res)
		}
	}

	e := &ResourceEvent{edis.NewEvent(E_RESOURCE_ADDED), res, []*Resource{res}, true}
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "trigger %s added", res.ModelStruct.Fqn())
		}
	}()
	if len(cb) > 0 {
		for i, cb := range cb {
			if err = cb(e); err != nil {
				return errors.Wrapf(err, "callback[%d]", i)
			}
		}
		return nil
	}
	return this.TriggerResource(e)
}

var onResourcesAddID int

func (this *Admin) OnResourcesAdded(cb func(e *ResourceEvent) error, id_ string, ids ...string) (err error) {
	key := onResourcesAddID
	onResourcesAddID++

	ids = append([]string{id_}, ids...)
	var waitForID []struct {
		i  int
		id string
	}

	var done = make(chan interface{})
	var resources = make([]*Resource, len(ids), len(ids))

	for i, id_ := range ids {
		if res, ok := this.ResourcesByUID[id_]; ok {
			resources[i] = res
		} else if res := this.GetResourceByID(id_); res != nil {
			resources[i] = res
		} else {
			waitForID = append(waitForID, struct {
				i  int
				id string
			}{i, id_})
		}
	}

	log := logging.WithPrefix(log, "on resources added #"+fmt.Sprint(key, ids))

	var newCb func() error
	if len(ids) == 1 {
		newCb = func() error {
			res := resources[0]
			return cb(&ResourceEvent{
				edis.NewEvent(E_RESOURCE_ADDED + ":" + res.FullID()),
				res, resources,
				true})
		}
	} else {
		newCb = func() error {
			return cb(&ResourceEvent{edis.NewEvent(E_RESOURCES_ADDED), nil, resources, true})
		}
	}

	var waitCount = len(waitForID)
	if waitCount == 0 {
		return newCb()
	}
	go func() {
		msg := "wait " + reflect.TypeOf(cb).PkgPath()

		var (
			count  int
			waited bool
		)
		for {
			select {
			case <-done:
				log.Debug(msg + " done")
				return
			default:
				var waits []string
				for _, v := range waitForID {
					if resources[v.i] == nil {
						waits = append(waits, v.id)
					}
				}
				if len(waits) > 0 {
					if count++; count == 60 {
						log.Warning(msg, "for", waits)
					} else {
						log.Debug(msg, "for", waits)
					}
					if waited {
						return
					}
					waited = true
					time.Sleep(time.Second / 2)
				}
			}
		}
	}()

	for _, id_ := range waitForID {
		func(id_ string, i int) {
			err = this.OnE(E_RESOURCE_ADDED+":"+id_, func(e edis.EventInterface) error {
				resources[i] = e.(*ResourceEvent).Resource
				waitCount--
				if waitCount == 0 {
					close(done)
					log.Debug("callback start")
					defer log.Debug("callback done")
					return newCb()
				}
				return nil
			})
		}(id_.id, id_.i)
		if err != nil {
			return
		}
	}
	return
}

func (this *Admin) OnResourceAdded(cb func(e *ResourceEvent)) error {
	log.Warning("DEPRECATED: method Admin.OnResourceAdded, use Admin.OnResourcesAdded")
	return this.OnE(E_RESOURCE_ADDED, func(e edis.EventInterface) {
		cb(e.(*ResourceEvent))
	})
}

func (this *Admin) OnResourceAddedE(cb func(e *ResourceEvent) error) error {
	deprecated("Admin.OnResourceAdded", "use Admin.OnResourcesAdded")
	return this.OnE(E_RESOURCE_ADDED, func(e edis.EventInterface) error {
		return cb(e.(*ResourceEvent))
	})
}

func (this *Admin) OnDone(cb func(e *AdminEvent)) error {
	return this.OnE(E_DONE, func(e edis.EventInterface) {
		cb(e.(*AdminEvent))
	})
}

func (this *Admin) OnDoneE(cb func(e *AdminEvent) error) error {
	return this.OnE(E_DONE, func(e edis.EventInterface) error {
		return cb(e.(*AdminEvent))
	})
}

type RecordeEvent struct {
	plug.EventInterface
	Resource *Resource
	Context  *core.Context
	Recorde  interface{}
}

type resourceAddWaiter struct {
}

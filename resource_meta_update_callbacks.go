package admin

type MetaUpdateCallback = func(meta *Meta)

type MetaUpdateCallbacks struct {
	callbacks []MetaUpdateCallback
	m         map[string]map[string]int
}

func (this *MetaUpdateCallbacks) Set(metaName, callbackName string, cb MetaUpdateCallback) {
	if this.m == nil {
		this.m = map[string]map[string]int{}
	}
	if callbacks, ok := this.m[metaName]; ok {
		if index, ok := callbacks[callbackName]; ok {
			this.callbacks[index] = cb
		}
		return
	}
	this.callbacks = append(this.callbacks, cb)
	if callbacks, ok := this.m[metaName]; ok {
		callbacks[callbackName] = len(this.callbacks)-1
	} else {
		this.m[metaName] = map[string]int{
			callbackName: len(this.callbacks)-1,
		}
	}
}

func (this *MetaUpdateCallbacks) Get(metaName, callbackName string) MetaUpdateCallback {
	if this.m == nil {
		return nil
	}
	if callbacks, ok := this.m[metaName]; ok {
		if index, ok := callbacks[callbackName]; ok {
			return this.callbacks[index]
		}
	}
	return nil
}

func (this *MetaUpdateCallbacks) GetAll(metaName string) (result []MetaUpdateCallback) {
	if this.m != nil {
		if indexes, ok := this.m[metaName]; ok {
			result = make([]MetaUpdateCallback, len(indexes), len(indexes))
			var i int
			for _, index := range indexes {
				result[i] = this.callbacks[index]
			}
		}
	}
	return
}

func (this *Resource) MetaUpdateCallback(metaName string) MetaUpdateCallbackManager {
	return MetaUpdateCallbackManager{
		func(name string) MetaUpdateCallback {
			return this.metaUpdateCallbacks.Get(metaName, name)
		},
		func(name string, cb MetaUpdateCallback) {
			this.metaUpdateCallbacks.Set(metaName, name, cb)
			if m, ok := this.MetasByName[metaName]; ok {
				cb(m)
			}
		},
	}
}

type MetaUpdateCallbackManager struct {
	Get func(name string) MetaUpdateCallback
	Set func(name string, cb MetaUpdateCallback)
}

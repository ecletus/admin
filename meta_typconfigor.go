package admin

import "reflect"

var metaTypeConfigorMaps = map[reflect.Type]func(meta *Meta){}

func RegisterMetaTypeConfigor(typ reflect.Type, configor func(meta *Meta)) {
	if old, ok := metaTypeConfigorMaps[typ]; ok {
		oldT := reflect.TypeOf(old)
		newT := reflect.TypeOf(configor)
		log.Warningf("override meta type cofigor for %s.%s from %s.%s to %s.%s",
			typ.PkgPath(), typ.Name(), oldT.PkgPath(), oldT.Name(), newT.PkgPath(), newT.Name())
	}
	metaTypeConfigorMaps[typ] = configor
}

func GetMetaTypeConfigor(typ reflect.Type) func(meta *Meta) {
	if old, ok := metaTypeConfigorMaps[typ]; ok {
		return old
	}
	return nil
}

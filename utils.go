package admin

import (
	"fmt"
	"net/url"
	"reflect"
)

func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func equalAsString(a interface{}, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func HasDeletedUrlQuery(values url.Values) (ok bool) {
	_, ok = values[":deleted"]
	return
}

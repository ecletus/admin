package admin

import (
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
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

func deprecated(funcName, msg string) {
	_, file, line, _ := runtime.Caller(2)
	log.Warningf("DEPRECATED: `" + funcName + "`: " + msg + " -> " + file + ":" + strconv.Itoa(line))
}

func deprecatedf(funcName, msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	log.Warningf("DEPRECATED: `" + funcName + "`: " + fmt.Sprintf(msg, args...) + " -> " + file + ":" + strconv.Itoa(line))
}

func StringChunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}
	var (
		l      int
		chunks []string
	)
	chunk := make([]rune, chunkSize)
	for _, r := range s {
		chunk[l] = r
		l++
		if l == chunkSize {
			chunks = append(chunks, string(chunk))
			l = 0
		}
	}
	if l > 0 {
		chunks = append(chunks, string(chunk[:l]))
	}
	return chunks
}

func InStrings(value string, lis ...string) bool {
	for _, v := range lis {
		if v == value {
			return true
		}
	}
	return false
}
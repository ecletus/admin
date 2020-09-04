package admin_helpers

import "strconv"

func CollectionIntRange(from, to, step int) (items [][]string) {
	for ; from <= to; from += step {
		v := strconv.Itoa(from)
		items = append(items, []string{v, v})
	}
	return
}

package admin

type resourceTree struct {
	resources map[string]*resourceTree
	value     *Resource
}

func (rt *resourceTree) Get(path ...string) (v *Resource) {
	t := rt
	var ok bool
	for _, p := range path {
		if t.resources == nil {
			return
		}
		if t, ok = t.resources[p]; !ok {
			return
		}
	}
	return t.value
}

func (rt *resourceTree) Set(v *Resource, path ...string) (t *resourceTree) {
	t = rt
	var (
		node *resourceTree
		ok   bool
	)
	for _, p := range path {
		if t.resources == nil {
			t.resources = map[string]*resourceTree{}
		}

		if node, ok = t.resources[p]; !ok {
			node = &resourceTree{}
			t.resources[p] = node
		}
		t = node
	}
	t.value = v
	return
}

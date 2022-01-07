package admin

type MetaStackItem struct {
	Meta *Meta
	Key  string
}

type MetaStack []*MetaStackItem

func (this MetaStack) Empty() bool {
	return len(this) == 0
}

func (this MetaStack) Index(pos int) *MetaStackItem {
	if pos < 0 {
		pos = len(this) + pos
	}
	if pos < 0 || pos > len(this) {
		return nil
	}
	return this[pos]
}

func (this *MetaStack) Push(m *Meta, key ...string) (pop func()) {
	var k string
	for _, k = range key {
	}
	*this = append(*this, &MetaStackItem{m, k})
	return this.Pop
}

func (this *MetaStack) Pop() {
	*this = (*this)[:len(*this)-1]
}

func (this MetaStack) PeekPair() (meta *Meta, key string) {
	l := len(this)
	if l == 0 {
		return
	}
	item := this[l-1]
	meta, key = item.Meta, item.Key
	return
}

func (this MetaStack) Peek() (meta *Meta) {
	l := len(this)
	if l == 0 {
		return
	}
	item := this[l-1]
	return item.Meta
}

func (this MetaStack) Path() (pth []string) {
	pth = make([]string, len(this))
	for i, item := range this {
		if item.Key == "" {
			pth[i] = item.Meta.Name
		} else {
			pth[i] = item.Key
		}
	}
	return
}

func (this *MetaStack) PushNames(names ...string) (pop func()) {
	l := len(*this)
	for _, name := range names {
		this.Push(&Meta{}, name)
	}
	return func() {
		*this = (*this)[:l]
	}
}

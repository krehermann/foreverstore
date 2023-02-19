package util

type Iterable[V any] struct {
	pos  int
	data []V
}

func NewIterable[V any](v []V) *Iterable[V] {
	return &Iterable[V]{
		data: v,
	}
}

func (it *Iterable[V]) hasNext() bool {
	return it.pos < len(it.data)
}
func (it *Iterable[V]) Next() (V, bool) {
	var dflt V
	if it.pos >= len(it.data) {
		return dflt, false //fmt.Errorf("out of range at %d [%d]", it.pos, len(it.data))
	}

	val := it.data[it.pos]
	it.pos += 1
	return val, true
}

func (it *Iterable[V]) Reset() {
	it.pos = 0
}

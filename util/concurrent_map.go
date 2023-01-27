package util

import "sync"

type ConcurrentMap[T comparable, V any] struct {
	mu   sync.RWMutex
	data map[T]V
}

func NewConcurrentMap[T comparable, V any]() *ConcurrentMap[T, V] {
	return &ConcurrentMap[T, V]{
		mu:   sync.RWMutex{},
		data: make(map[T]V),
	}
}

func (m *ConcurrentMap[T, V]) Put(key T, val V) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
	return nil
}
func (m *ConcurrentMap[T, V]) Get(key T) (V, bool) {
	m.mu.RLock()
	val, ok := m.data[key]
	m.mu.RUnlock()
	return val, ok
}

func (m *ConcurrentMap[T, V]) Delete(key T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

package store

import (
	"fmt"
	"sort"
	"sync"

	"github.com/krehermann/foreverstore/util"
)

type Metastore interface {
	Register(*ObjectRef) error
	Get(key string, version int) (*VersionedObjectRef, error)
	GetLatest(key string) (*VersionedObjectRef, error)
}

type MemMeta struct {
	mu sync.RWMutex
	m  util.ConcurrentMap[string, []*VersionedObjectRef]
}

func (m *MemMeta) Register(r *ObjectRef) error {
	// there is a race here -- multiple clients could be writing
	// seems to defeat the purpose of concurrent map...
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, exists := m.m.Get(r.Key)
	if !exists {
		cur = make([]*VersionedObjectRef, 0)
	}
	v := &VersionedObjectRef{
		ObjectRef: r,
		Version:   len(cur),
	}
	cur = append(cur, v)
	return m.m.Put(r.Key, cur)
}

func (m *MemMeta) Get(key string, version int) (*VersionedObjectRef, error) {
	objs, ok := m.m.Get(key)
	if !ok {
		return nil, fmt.Errorf("key does not exist")
	}
	for _, obj := range objs {
		if obj.Version == version {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("version does not exist")
}

func (m *MemMeta) GetLatest(key string) (*VersionedObjectRef, error) {
	objs, ok := m.m.Get(key)
	if !ok {
		return nil, fmt.Errorf("key does not exist")
	}
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].Version > objs[j].Version
	})
	return objs[0], nil
}

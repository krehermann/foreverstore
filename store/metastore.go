package store

import (
	"fmt"
	"io"
	"io/fs"
	"sort"
	"sync"

	"github.com/krehermann/foreverstore/util"
)

type GetConfig struct {
	version int
}

type GetOpt func(*GetConfig)

func GetVersion(version int) GetOpt {
	return func(c *GetConfig) {
		c.version = version
	}
}

type Metastore interface {
	//	Register(*ObjectRef) error
	Get(key string, opts ...GetOpt) (*VersionedObjectRef, error)
	Put(key string, r io.Reader) error
	//	GetLatest(key string) (*VersionedObjectRef, error)
	//	Create(key string) (*VersionedObjectRef, error)
	ReadWriteFS
}

type MemMeta struct {
	mu sync.RWMutex
	m  *util.ConcurrentMap[string, []*VersionedObjectRef]
	fs ReadWriteFS
}

func NewMemMeta(fs ReadWriteFS) *MemMeta {
	return &MemMeta{
		fs: fs,
		m:  util.NewConcurrentMap[string, []*VersionedObjectRef](),
	}
}

func (m *MemMeta) Open(key string) (fs.File, error) {
	v, err := m.GetLatest(key)
	if err != nil {
		return nil, err
	}
	return m.fs.Open(v.Path)
}

func (m *MemMeta) Create(key string) (*Blob, error) {
	return NewWritableBlob(key)
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
			f, err := m.fs.Open(obj.Path)
			if err != nil {
				return nil, err
			}
			obj.handle = f
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

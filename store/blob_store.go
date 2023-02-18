package store

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

type BlobStoreConfig struct {
	PathFunc
	// optional. consider moving to opts func instead of config
	Root   string
	Logger *zap.Logger
}

type BlobStore struct {
	config BlobStoreConfig

	registerCh chan<- *ObjectRef
	// blobMap tracks key-> path relationship
	// TODO persistency & loading
	blobMap *util.ConcurrentMap[string, string]
}

var _ ReadWriteStatFS = (*BlobStore)(nil)

func NewBlobStore(config BlobStoreConfig) (*BlobStore, error) {
	if config.PathFunc == nil {
		config.PathFunc = ContentPath
	}
	if config.Logger == nil {
		var err error
		config.Logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}
	config.Logger = config.Logger.Named("BlobStore")
	if config.Root == "" {
		d, err := os.MkdirTemp("", "fs-root")
		if err != nil {
			return nil, err
		}
		config.Root = d
	}

	_, err := os.Stat(config.Root)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(config.Root, 0755)
		if err != nil {
			return nil, err
		}
	}

	return &BlobStore{
		config:     config,
		registerCh: make(chan<- *ObjectRef),
		blobMap:    util.NewConcurrentMap[string, string](),
	}, nil
}

func (s *BlobStore) Remove(key string) error {
	s.config.Logger.Sugar().Infof("removing key %s", key)
	pth, ok := s.blobMap.Get(key)
	if !ok {
		return fmt.Errorf("%w: %s", os.ErrNotExist, key)
	}
	fp := s.fullPath(pth)
	s.config.Logger.Sugar().Debugf("removing key '%s' at path %s", key, fp)
	// delete the file

	err := os.Remove(fp)
	if err != nil {
		return err
	}

	// remove from map
	s.blobMap.Delete(key)

	filepath.Walk(s.config.Root, func(path string, info fs.FileInfo, err error) error {
		s.config.Logger.Sugar().Debugf("walking root %s: %s %v %v", s.config.Root,
			path, info, err)
		return nil
	})

	// walk fs from top level dir of the given file
	// delete dirs that are empty
	relPath := s.relPath(pth)
	s.config.Logger.Sugar().Debugf("walking rel path %s to cleanup", relPath)

	splits := strings.Split(relPath, string(os.PathSeparator))

	s.config.Logger.Sugar().Debugf("splits %+v to cleanup '%s'", splits, splits[0])

	if len(splits) == 0 {
		return nil
	}
	relRoot := s.fullPath(splits[0])
	s.config.Logger.Sugar().Debugf("walking rel root %s to cleanup", relRoot)
	stack := []string{}
	walkDirFn := func(path string, d fs.DirEntry, err error) error {
		s.config.Logger.Sugar().Debugf("walkDirfn path %s, d %+v, err %v", path, d, err)
		if err != nil {
			s.config.Logger.Sugar().Errorf("walkDirFn %v", err)
			return nil
		}
		if d.IsDir() {
			stack = append(stack, path)
		}
		return nil
	}

	err = fs.WalkDir(s, relRoot, walkDirFn)
	if err != nil {
		return nil
	}
	for i := len(stack) - 1; i >= 0; i-- {
		pth := stack[i]
		dirInfo, err := fs.ReadDir(s, pth)
		if err != nil {
			return err
		}
		if len(dirInfo) == 0 {
			err = os.Remove(s.fullPath(pth))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *BlobStore) onClose(b *Blob) error {
	pth := filepath.Join(s.config.Root, s.config.PathFunc(b.Hash))
	err := os.MkdirAll(filepath.Dir(pth), 0755)
	if err != nil {
		return err
	}
	err = os.Rename(b.f.Name(), pth)
	if err != nil {
		return err
	}
	// register in the blob key->path map
	key := b.Name()
	b.rename(s.relPath(pth))
	s.blobMap.Put(key, b.Name())
	return nil
}

func (s *BlobStore) Create(name string) (WriteFile, error) {
	// the blob is not tracked in the map until it's closed
	return NewWritableBlob(name, WithCloseFn(s.onClose))
}

func (s *BlobStore) ReadFile(key string) ([]byte, error) {
	pth, ok := s.blobMap.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", os.ErrNotExist, key)
	}
	return ioutil.ReadFile(s.fullPath(pth))
}

func (s *BlobStore) Open(key string) (fs.File, error) {
	pth, ok := s.blobMap.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", os.ErrNotExist, key)
	}

	fp := s.fullPath(pth)
	_, err := s.Stat(fp)
	if err != nil {
		panic(fmt.Sprintf("key %s in map, but underlying file %s stat fails: %v", key, fp, err))
	}
	return NewReadonlyBlob(s.fullPath(pth))
}

func (s *BlobStore) fullPath(p string) string {
	if !strings.HasPrefix(p, s.config.Root) {
		p = filepath.Join(s.config.Root, p)
	}
	return p
}

func (s *BlobStore) relPath(p string) string {
	if filepath.IsAbs(p) {
		p = strings.TrimPrefix(p, s.config.Root)
	}
	return strings.TrimLeft(p, string(os.PathSeparator))

}

// path is the resolved path in the blob store, not the key
func (s *BlobStore) Stat(path string) (fs.FileInfo, error) {
	fp := s.fullPath(path)
	s.config.Logger.Sugar().Infof("stat %s", fp)
	return os.Stat(fp)
}

// path is the resolved path in the blob store, not the key
func (s *BlobStore) ReadDir(path string) ([]fs.DirEntry, error) {
	fp := s.fullPath(path)
	s.config.Logger.Sugar().Infof("stat %s", fp)
	return os.ReadDir(fp)
}

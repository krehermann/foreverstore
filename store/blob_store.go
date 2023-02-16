package store

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

type BlobStoreConfig struct {
	PathFunc
	Root   string
	Logger *zap.Logger
	Metastore
}

type BlobStore struct {
	config BlobStoreConfig

	registerCh chan<- *ObjectRef
}

var _ ReadWriteFS = (*BlobStore)(nil)

func NewBlobStore(config BlobStoreConfig) (*BlobStore, error) {
	if config.PathFunc == nil {
		config.PathFunc = awsContentPath
	}
	if config.Logger == nil {
		var err error
		config.Logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}
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
	}, nil
}

func (s *BlobStore) Remove(p string) error {
	// delete the file
	err := os.Remove(s.fullPath(p))
	if err != nil {
		return err
	}

	// walk fs from top level dir of the given file
	// delete dirs that are empty
	relPath := s.relPath(p)
	splits := filepath.SplitList(relPath)
	if len(splits) == 0 {
		return nil
	}
	relRoot := s.fullPath(splits[0])
	stack := []string{}
	walkDirFn := func(path string, d fs.DirEntry, err error) error {
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
	b.rename(s.relPath(pth))
	return nil
}

func (s *BlobStore) Create(name string) (WriteFile, error) {
	return NewWritableBlob(name, WithCloseFn(s.onClose))
}

func (s *BlobStore) ReadFile(pth string) ([]byte, error) {

	return ioutil.ReadFile(s.fullPath(pth))
}

func (s *BlobStore) Open(name string) (fs.File, error) {
	return NewReadonlyBlob(s.fullPath(name))
}

func (s *BlobStore) fullPath(p string) string {
	return filepath.Join(s.config.Root, p)
}

func (s *BlobStore) relPath(p string) string {
	if filepath.IsAbs(p) {
		return strings.TrimPrefix(p, s.config.Root)
	} else {
		return p
	}
}

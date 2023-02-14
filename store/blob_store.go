package store

import (
	"errors"
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

/*
func (s *FileStore) Create(key string, r io.Reader) (*ObjectRef, error) {

	// write to a temporary location
	// mv to content address

	hashWriter := sha256.New()
	tempFile, err := os.CreateTemp("", key)
	if err != nil {
		return nil, err
	}
	mw := io.MultiWriter(tempFile, hashWriter)
	n, err := io.Copy(mw, r)
	if err != nil {
		return nil, err
	}

	err = tempFile.Close()
	if err != nil {
		return nil, err
	}

	pth := filepath.Join(s.config.Root, s.config.PathFunc(hashWriter, key))
	err = os.MkdirAll(filepath.Dir(pth), 0755)
	if err != nil {
		return nil, err
	}
	err = os.Rename(tempFile.Name(), pth)
	if err != nil {
		return nil, err
	}

	return &ObjectRef{
		Key:  key,
		Path: pth,
		Size: n,
	}, nil

}
*/

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

/*
func (s *BlobStore) Create(name string) (*Blob, error) {
	return NewWritableBlob(name, WithCloseFn(s.onClose))
}
*/

func (s *BlobStore) Create(name string) (WriteFile, error) {
	return NewWritableBlob(name, WithCloseFn(s.onClose))
}

func (s *BlobStore) ReadFile(pth string) ([]byte, error) {

	return ioutil.ReadFile(s.fullPath(pth))
}

func (s *BlobStore) Open(name string) (File, error) {
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

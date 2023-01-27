package store

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type FileStoreConfig struct {
	PathFunc
	Root   string
	Logger *zap.Logger
	Metastore
}

type FileStore struct {
	config FileStoreConfig

	registerCh chan<- *ObjectRef
}

func NewFileStore(config FileStoreConfig) (*FileStore, error) {
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
		d, err := ioutil.TempDir("", "fs-root")
		if err != nil {
			return nil, err
		}
		config.Root = d
	}
	return &FileStore{
		config:     config,
		registerCh: make(chan<- *ObjectRef),
	}, nil
}

func (s *FileStore) Create(key string, r io.Reader) (*ObjectRef, error) {

	// write to a temporary location
	// mv to content address

	hashWriter := sha256.New()
	tempFile, err := ioutil.TempFile("", key)
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

package store

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

type blobMode int

const (
	ReadOnly blobMode = iota
	ReadWrite
)

type closeFn func(b *Blob) error

var _ WriteFile = (*Blob)(nil)
var _ fs.File = (*Blob)(nil)

type Blob struct {
	mode blobMode

	mu   sync.RWMutex
	name string

	closeFn
	//handle File
	hash.Hash
	f           *os.File
	multiWriter io.Writer
}

type BlobOpt func(*Blob)

func WithCloseFn(fn closeFn) BlobOpt {
	return func(b *Blob) {
		b.closeFn = fn
	}
}

// read vs write blob?
func NewWritableBlob(name string, opts ...BlobOpt) (*Blob, error) {

	t, err := os.CreateTemp("", fmt.Sprintf("blob-%s", name))
	if err != nil {
		return nil, err
	}
	hashWriter := sha256.New()
	b := &Blob{
		name:        name,
		mode:        ReadWrite,
		Hash:        hashWriter,
		f:           t,
		multiWriter: io.MultiWriter(t, hashWriter),
	}

	for _, opt := range opts {
		opt(b)
	}
	return b, nil
}

func NewReadonlyBlob(name string) (*Blob, error) {
	return &Blob{
		mode: ReadOnly,
		name: name,
	}, nil
}

func (b *Blob) Write(buf []byte) (int, error) {
	if b.mode == ReadOnly {
		return 0, fmt.Errorf("can't write a read only blob")
	}
	r := bytes.NewReader(buf)
	n, err := io.Copy(b.multiWriter, r)
	return int(n), err

}

func (b *Blob) Close() error {

	err := b.f.Close()
	if err != nil {
		return err
	}
	if b.closeFn != nil {
		if err := b.closeFn(b); err != nil {
			return nil
		}
	}
	b.f = nil
	return nil
}

func (b *Blob) Read(buf []byte) (int, error) {
	return b.f.Read(buf)
}

func (b *Blob) Stat() (os.FileInfo, error) {
	return &BlobInfo{
		name: b.name,
	}, nil
}

func (b *Blob) Name() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.name
}

func (b *Blob) rename(n string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.name = n
	return nil
}

type BlobInfo struct {
	name string
}

func (i *BlobInfo) Name() string {
	return i.name
}

func (i *BlobInfo) Size() int64 {
	return 0
}

func (i *BlobInfo) Mode() os.FileMode {
	return os.FileMode(0)
}
func (i *BlobInfo) ModTime() time.Time {
	// modification time
	return time.Unix(0, 0)
}
func (i *BlobInfo) IsDir() bool {
	return false
} // abbreviation for Mode().IsDir()

func (i *BlobInfo) Sys() any {
	return nil
}

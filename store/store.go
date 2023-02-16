package store

import (
	"io"
	"io/fs"
	"os"
)

type Storer interface {
	io.WriteCloser
}

// TODO thinking ahead -- an object can be reliplicated
// who and where is that tracked?

type ObjectRef struct {
	Key  string
	Path string
	//	host string
	Size int64

	handle fs.File //*os.File
}

func (r *ObjectRef) Read(b []byte) (int, error) {
	return r.handle.Read(b)
}

func (r *ObjectRef) Close() error {
	return r.handle.Close()
}

func (r *ObjectRef) Stat() (os.FileInfo, error) {
	return r.handle.Stat()
}

type VersionedObjectRef struct {
	*ObjectRef
	Version int
}

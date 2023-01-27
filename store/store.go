package store

import "io"

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
}

type VersionedObjectRef struct {
	*ObjectRef
	Version int
}

package store

import (
	"io/fs"
)

type WriteFile interface {
	fs.File
	Write([]byte) (int, error)
}

type WriteFS interface {
	fs.FS
	Create(name string) (WriteFile, error)
	Remove(name string) error
}

type ReadWriteStatFS interface {
	fs.ReadFileFS
	WriteFS
	fs.StatFS
	fs.ReadDirFS
}

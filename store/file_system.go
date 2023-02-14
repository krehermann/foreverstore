package store

import (
	"os"
)

type File interface {
	Read([]byte) (int, error)
	Close() error
	Stat() (os.FileInfo, error)
}

type WriteFile interface {
	File
	Write([]byte) (int, error)
}

type FS interface {
	Open(name string) (File, error)
}

type ReadFileFS interface {
	FS
	ReadFile(name string) ([]byte, error)
}

type WriteFS interface {
	FS
	Create(string) WriteFile
}

type ReadWriteFS interface {
	ReadFileFS
	WriteFS
}

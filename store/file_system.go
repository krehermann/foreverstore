package store

import "io/fs"

/*
	type File interface {
		Read([]byte) (int, error)
		Close() error
		Stat() (os.FileInfo, error)
	}
*/
type WriteFile interface {
	fs.File
	Write([]byte) (int, error)
}

type FS interface {
	Open(name string) (fs.File, error)
}

type ReadFileFS interface {
	FS
	ReadFile(name string) ([]byte, error)
}

type WriteFS interface {
	FS
	Create(name string) (WriteFile, error)
	Remove(name string) error
}

type ReadWriteFS interface {
	ReadFileFS
	WriteFS
}

package memos

import "os"

type File interface {
	Write(b []byte) (n int, err error)
	Close() error
}

var _ File = &os.File{}

type OS interface {
	Create(name string) (File, error)
	Getwd() (dir string, err error)
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)

	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

package memos

import "os"

type RealOS struct{}

func (RealOS) Create(name string) (File, error)             { return os.Create(name) }
func (RealOS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (RealOS) Getwd() (dir string, err error)               { return os.Getwd() }

func (RealOS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func (RealOS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (RealOS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (RealOS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

var _ OS = &RealOS{}

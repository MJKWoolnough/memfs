package memfs

import "io/fs"

type FS map[string]fs.DirEntry

func (f *FS) Open(path string) (fs.File, error) {
	return nil, nil
}

func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	return nil, nil
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	return nil, nil
}

func (f *FS) Sub(dir string) (fs.FS, error) {
	return nil, nil
}

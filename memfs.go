package memfs

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

const separator = string(filepath.Separator)

type FS dnode

func New() *FS {
	return &FS{
		modtime: time.Now(),
		mode:    0o777,
	}
}

func (f *FS) Open(path string) (fs.File, error) {
	dirName, fileName := filepath.Split(path)
	for _, p := range strings.Split(dirName, separator) {
		if p == "" {
			continue
		}

		for _, de := range f.entries {
			if de.name == p {
				if d, ok := de.directoryEntry.(*directory); ok {
					f = (*FS)(d.dnode)
				} else {
					return nil, fs.ErrNotExist
				}
			}
		}
	}

	for _, de := range f.entries {
		if de.name == fileName {
			return de.open(fileName, opRead|opSeek), nil
		}
	}

	return nil, fs.ErrNotExist
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

func (f *FS) Mkdir(name string, perm fs.FileMode) error {
	return nil
}

func (f *FS) MkdirAll(path string, perm fs.FileMode) error {
	return nil
}

type File interface {
	fs.File
	Write([]byte) (int, error)
}

func (f *FS) Create(name string) (File, error) {
	return nil, nil
}

func (f *FS) Link(oldname, newname string) error {
	return nil
}

func (f *FS) Symlink(oldname, newname string) error {
	return nil
}

func (f *FS) Rename(oldpath, newpath string) error {
	return nil
}

func (f *FS) Remove(name string) error {
	return nil
}

func (f *FS) RemoveAll(path string) error {
	return nil
}

func (f *FS) LStat(name string) (fs.FileInfo, error) {
	return nil, nil
}

func (f *FS) Readlink(name string) (string, error) {
	return "", nil
}

func (f *FS) Chown(name string, uid, gid int) error {
	return nil
}

func (f *FS) Chmod(name string, mode fs.FileMode) error {
	return nil
}

func (f *FS) Lchown(name string, uid, gid int) error {
	return nil
}

func (f *FS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

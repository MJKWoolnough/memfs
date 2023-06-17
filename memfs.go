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
	de := f.getEntry(path)
	if de == nil {
		return nil, fs.ErrNotExist
	}

	_, fileName := filepath.Split(path)

	return de.open(fileName, opRead|opSeek), nil
}

func (f *FS) getDirEnt(path string) *dnode {
	d := (*dnode)(f)
	for _, p := range strings.Split(path, separator) {
		if p == "" {
			continue
		}

		de := d.get(p)

		if de, ok := de.directoryEntry.(*directory); ok {
			d = de.dnode
		} else {
			return nil
		}
	}

	return d
}

func (f *FS) getEntry(path string) *dirEnt {
	dirName, fileName := filepath.Split(path)

	d := f.getDirEnt(dirName)
	if d == nil {
		return nil
	}

	return d.get(fileName)
}

func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	d := f.getDirEnt(name)
	if d == nil {
		return nil, fs.ErrNotExist
	}

	dirs := make([]fs.DirEntry, len(d.entries))

	for i := range d.entries {
		dirs[i] = d.entries[i]
	}

	return dirs, nil
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	de := f.getEntry(name)
	inode, ok := de.directoryEntry.(*inode)
	if !ok {
		return nil, fs.ErrInvalid
	}

	data := make([]byte, len(inode.data))

	copy(data, inode.data)

	return data, nil
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	de := f.getEntry(name)
	if de == nil {
		return nil, fs.ErrNotExist
	}

	return de.Info()
}

func (f *FS) Sub(dir string) (fs.FS, error) {
	dn := f.getDirEnt(dir)
	if dn == nil {
		return nil, fs.ErrNotExist
	}

	return (*FS)(dn), nil
}

func (f *FS) Mkdir(name string, perm fs.FileMode) error {
	parent, child := filepath.Split(name)
	d := f.getDirEnt(parent)
	if d == nil {
		return fs.ErrNotExist
	}

	if d.get(child) != nil {
		return fs.ErrExist
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: &dnode{
			name:    child,
			modtime: time.Now(),
			mode:    perm,
		},
		name: child,
	})
	return nil
}

func (f *FS) MkdirAll(path string, perm fs.FileMode) error {
	d := (*dnode)(f)

	var ok bool

	for _, p := range strings.Split(path, string(filepath.Separator)) {
		e := d.get(p)
		if e == nil {
			d := &dnode{
				name:    p,
				modtime: time.Now(),
				mode:    perm,
			}
			e = &dirEnt{
				directoryEntry: d,
				name:           p,
			}

			d.entries = append(d.entries, e)
		} else if d, ok = e.directoryEntry.(*dnode); !ok {
			return fs.ErrInvalid
		}
	}

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

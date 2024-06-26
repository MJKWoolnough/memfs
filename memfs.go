// Package memfs contains both ReadOnly and ReadWrite implementations of an in
// memory FileSystem, supporting all of the FS interfaces and more.
package memfs // import "vimagination.zapto.org/memfs"

import (
	"errors"
	"io/fs"
	"path"
)

type fsRO struct {
	de directoryEntry
}

func (f *fsRO) joinRoot(p string) string {
	return path.Join(slash, p)
}

func (f *fsRO) Open(p string) (fs.File, error) {
	de, err := f.getEntry(p)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: p, Err: err}
	}

	_, fileName := path.Split(p)

	of, err := de.open(fileName, opRead|opSeek)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: p, Err: err}
	}

	return of, nil
}

func (f *fsRO) getDirEnt(path string) (dNode, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, err
	} else if d, ok := de.(dNode); !ok {
		return nil, fs.ErrInvalid
	} else {
		return d, nil
	}
}

const (
	maxRedirects uint8 = 255
	slash              = "/"
)

func (f *fsRO) getEntry(path string) (directoryEntry, error) {
	if !fs.ValidPath(path) {
		return nil, fs.ErrInvalid
	}

	return f.getEntryWithoutCheck(path)
}

func (f *fsRO) getLEntry(p string) (*dirEnt, error) {
	if !fs.ValidPath(p) {
		return nil, fs.ErrInvalid
	}

	jpath := f.joinRoot(p)
	dirName, fileName := path.Split(jpath)

	de, err := f.getEntryWithoutCheck(dirName)
	if err != nil {
		return nil, err
	}

	if jpath == slash {
		return &dirEnt{
			directoryEntry: f.de,
			name:           slash,
		}, nil
	}

	return de.getEntry(fileName)
}

type exists byte

const (
	mustNotExist exists = iota
	mustExist
	doesntMatter
)

func (f *fsRO) getEntryWithParent(path string, exists exists) (dNode, *dirEnt, error) {
	parent, child := splitPath(path)
	if child == "" {
		return nil, nil, fs.ErrInvalid
	}

	d, err := f.getDirEnt(parent)
	if err != nil {
		return nil, nil, err
	}

	c, err := d.getEntry(child)
	if !errors.Is(err, fs.ErrNotExist) || exists == mustExist {
		if err != nil {
			return nil, nil, err
		} else if exists == mustNotExist {
			return nil, nil, fs.ErrExist
		}
	}

	return d, c, nil
}

func (f *fsRO) ReadDir(path string) ([]fs.DirEntry, error) {
	d, err := f.getDirEnt(path)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: path, Err: err}
	}

	es, err := d.getEntries()
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: path, Err: err}
	}

	return es, nil
}

func (f *fsRO) ReadFile(path string) ([]byte, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: path, Err: err}
	}

	data, err := de.bytes()
	if err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: path, Err: err}
	}

	return data, nil
}

func (f *fsRO) Stat(p string) (fs.FileInfo, error) {
	de, err := f.getEntry(p)
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: p, Err: err}
	}

	base := path.Base(p)

	if base == "." {
		base = slash
	}

	return &dirEnt{
		name:           base,
		directoryEntry: de,
	}, nil
}

func (f *fsRO) LStat(path string) (fs.FileInfo, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: path, Err: err}
	}

	return de, nil
}

func (f *fsRO) Readlink(path string) (string, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return "", &fs.PathError{Op: "readlink", Path: path, Err: err}
	}

	if de.Mode()&fs.ModeSymlink == 0 {
		return "", &fs.PathError{Op: "readlink", Path: path, Err: fs.ErrInvalid}
	}

	b, err := de.string()
	if err != nil {
		return "", &fs.PathError{Op: "readlink", Path: path, Err: err}
	}

	return b, nil
}

func (f *fsRO) sub(path string) (directoryEntry, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{Op: "sub", Path: path, Err: err}
	} else if !de.IsDir() {
		return nil, &fs.PathError{Op: "sub", Path: path, Err: fs.ErrInvalid}
	}

	return de, nil
}

func (f *fsRO) Sub(path string) (fs.FS, error) {
	de, err := f.sub(path)
	if err != nil {
		return nil, err
	}

	return &fsRO{
		de: de,
	}, nil
}

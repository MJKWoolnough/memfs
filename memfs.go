package memfs

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type FS struct {
	de directoryEntry
}

func (f *FS) joinRoot(path string) string {
	return filepath.Join("/", path)
}

func (f *FS) Open(path string) (fs.File, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  err,
		}
	}

	_, fileName := filepath.Split(path)

	of, err := de.open(fileName, opRead|opSeek)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  err,
		}
	}

	return of, nil
}

func (f *FS) getDirEnt(path string) (dNode, error) {
	redirectsRemaining := maxRedirects

	de, err := f.getResolvedDirEnt(f.joinRoot(path), &redirectsRemaining)
	if err != nil {
		return nil, err
	} else if d, ok := de.(dNode); !ok {
		return nil, fs.ErrInvalid
	} else {
		return d, nil
	}
}

var maxRedirects uint8 = 255

func (f *FS) getResolvedDirEnt(path string, remainingRedirects *uint8) (directoryEntry, error) {
	var (
		de  directoryEntry
		d   *dirEnt
		dn  dNode
		err error
		ok  bool
	)

	dir, base := filepath.Split(path)
	if dir == "" || dir == "/" {
		if f.de.Mode()&0o444 == 0 {
			return nil, fs.ErrPermission
		}

		de = f.de
	} else {
		if de, err = f.getResolvedDirEnt(filepath.Clean(dir), remainingRedirects); err != nil {
			return nil, err
		}
	}

	if base == "" {
		return de, nil
	} else if dn, ok = de.(dNode); !ok {
		return nil, fs.ErrInvalid
	} else if d, err = dn.getEntry(base); err != nil {
		return nil, err
	} else if d.Mode()&fs.ModeSymlink == 0 {
		return d.directoryEntry, nil
	} else if *remainingRedirects == 0 {
		return nil, fs.ErrInvalid
	}

	*remainingRedirects--

	b, err := d.bytes()
	if err != nil {
		return nil, err
	}

	link := string(b)

	if !strings.HasPrefix(link, "/") {
		link = filepath.Join(dir, link)
	}

	return f.getResolvedDirEnt(link, remainingRedirects)
}

func (f *FS) getEntry(path string) (directoryEntry, error) {
	redirectsRemaining := maxRedirects

	return f.getResolvedDirEnt(f.joinRoot(path), &redirectsRemaining)
}

func (f *FS) getLEntry(path string) (*dirEnt, error) {
	jpath := f.joinRoot(path)
	dirName, fileName := filepath.Split(jpath)
	redirectsRemaining := maxRedirects

	de, err := f.getResolvedDirEnt(dirName, &redirectsRemaining)
	if err != nil {
		return nil, err
	}

	d, ok := de.(dNode)
	if !ok {
		fmt.Printf("%T %v", de, de)
		return nil, fs.ErrInvalid
	} else if jpath == "/" {
		return &dirEnt{
			directoryEntry: f.de,
			name:           "/",
		}, nil
	}

	return d.getEntry(fileName)
}

type exists byte

const (
	mustNotExist exists = iota
	mustExist
	doesntMatter
)

func (f *FS) getEntryWithParent(path string, exists exists) (dNode, *dirEnt, error) {
	parent, child := filepath.Split(path)
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

func (f *FS) ReadDir(path string) ([]fs.DirEntry, error) {
	d, err := f.getDirEnt(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  err,
		}
	}

	es, err := d.getEntries()
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  err,
		}
	}

	return es, nil
}

func (f *FS) ReadFile(path string) ([]byte, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: path,
			Err:  err,
		}
	}

	b, err := de.bytes()
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: path,
			Err:  err,
		}
	}

	data := make([]byte, len(b))

	copy(data, b)

	return data, nil
}

func (f *FS) Stat(path string) (fs.FileInfo, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: path,
			Err:  err,
		}
	}

	base := filepath.Base(path)

	if base == "." {
		base = "/"
	}

	return &dirEnt{
		name:           base,
		directoryEntry: de,
	}, nil
}

func (f *FS) LStat(path string) (fs.FileInfo, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "lstat",
			Path: path,
			Err:  err,
		}
	}

	return de, nil
}

func (f *FS) Readlink(path string) (string, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  err,
		}
	}

	if de.Mode()&fs.ModeSymlink == 0 {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  fs.ErrInvalid,
		}
	}

	b, err := de.bytes()
	if err != nil {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  err,
		}
	}

	return string(b), nil
}

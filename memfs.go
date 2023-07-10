package memfs

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type fsRO struct {
	de directoryEntry
}

func (f *fsRO) joinRoot(path string) string {
	return filepath.Join(slash, path)
}

func (f *fsRO) Open(path string) (fs.File, error) {
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

var (
	maxRedirects uint8 = 255
	slash              = string(filepath.Separator)
)

func (f *fsRO) getEntry(path string) (directoryEntry, error) {
	if f.de.Mode()&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	path = f.joinRoot(path)
	remainingRedirects := maxRedirects

	curr := f.de
	currPath := slash
	path = path[1:]

	for path != "" {
		slashPos := strings.Index(path, slash)

		var name string

		if slashPos == -1 {
			name = path
			path = ""
		} else {
			name = path[:slashPos]
			path = path[slashPos+1:]
		}

		if name == "" {
			continue
		}

		if next, err := curr.getEntry(name); err != nil {
			return nil, err
		} else if next.Mode()&fs.ModeSymlink == 0 {
			curr = next.directoryEntry
			currPath = filepath.Join(currPath, name)
		} else if remainingRedirects == 0 {
			return nil, fs.ErrInvalid
		} else {

			remainingRedirects--

			b, err := next.bytes()
			if err != nil {
				return nil, err
			}

			link := filepath.Clean(string(b))

			if !strings.HasPrefix(link, slash) {
				link = filepath.Join(currPath, link)
			}

			currPath = slash
			path = filepath.Join(link, path)
			curr = f.de
		}
	}

	return curr, nil
}

func (f *fsRO) getLEntry(path string) (*dirEnt, error) {
	jpath := f.joinRoot(path)
	dirName, fileName := filepath.Split(jpath)

	de, err := f.getEntry(dirName)
	if err != nil {
		return nil, err
	}

	d, ok := de.(dNode)
	if !ok {
		fmt.Printf("%T %v", de, de)
		return nil, fs.ErrInvalid
	} else if jpath == slash {
		return &dirEnt{
			directoryEntry: f.de,
			name:           slash,
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

func (f *fsRO) getEntryWithParent(path string, exists exists) (dNode, *dirEnt, error) {
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

func (f *fsRO) ReadDir(path string) ([]fs.DirEntry, error) {
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

func (f *fsRO) ReadFile(path string) ([]byte, error) {
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

func (f *fsRO) Stat(path string) (fs.FileInfo, error) {
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
		return nil, &fs.PathError{
			Op:   "lstat",
			Path: path,
			Err:  err,
		}
	}

	return de, nil
}

func (f *fsRO) Readlink(path string) (string, error) {
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

func (f *fsRO) sub(path string) (directoryEntry, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "sub",
			Path: path,
			Err:  err,
		}
	} else if !de.IsDir() {
		return nil, &fs.PathError{
			Op:   "sub",
			Path: path,
			Err:  fs.ErrInvalid,
		}
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

package memfs

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FSRW struct {
	mu sync.RWMutex
	FS
}

func (f *FSRW) ReadDir(path string) ([]fs.DirEntry, error) {
	d, err := f.getDirEnt(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  err,
		}
	}

	if d.mode&0o444 == 0 {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  fs.ErrPermission,
		}
	}

	dirs := make([]fs.DirEntry, len(d.entries))

	for i := range d.entries {
		dirs[i] = d.entries[i]
	}

	return dirs, nil
}

func (f *FSRW) ReadFile(path string) ([]byte, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: path,
			Err:  err,
		}
	}

	inode, ok := de.directoryEntry.(*inode)
	if !ok {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: path,
			Err:  fs.ErrInvalid,
		}
	}

	if inode.mode&0o444 == 0 {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: path,
			Err:  fs.ErrPermission,
		}
	}

	data := make([]byte, len(inode.data))

	copy(data, inode.data)

	return data, nil
}

func (f *FSRW) Stat(path string) (fs.FileInfo, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: path,
			Err:  err,
		}
	}

	fi, err := de.Info()
	if err != nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: path,
			Err:  err,
		}
	}

	return fi, nil
}

func (f *FSRW) Mkdir(path string, perm fs.FileMode) error {
	return f.mkdir("mkdir", path, path, perm)
}

func (f *FSRW) mkdir(op, opath, path string, perm fs.FileMode) error {
	d, _, err := f.getEntryWithParent(path, mustNotExist)
	if err != nil {
		return &fs.PathError{
			Op:   op,
			Path: opath,
			Err:  err,
		}
	}

	if d.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   op,
			Path: opath,
			Err:  fs.ErrPermission,
		}
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: &dnodeRW{
			dnode: dnode{
				modtime: time.Now(),
				mode:    fs.ModeDir | perm,
			},
		},
		name: filepath.Base(path),
	})
	d.modtime = time.Now()

	return nil
}

func (f *FSRW) MkdirAll(path string, perm fs.FileMode) error {
	cpath := filepath.Join("/", path)
	last := 0

	for {
		pos := strings.IndexRune(cpath[last:], filepath.Separator)
		if pos < 0 {
			break
		} else if pos == 0 {
			last++

			continue
		}

		last += pos

		if err := f.mkdir("mkdirall", path, cpath[:last], perm); err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}
	}

	return f.mkdir("mkdirall", path, cpath, perm)
}

type File interface {
	fs.File
	Write([]byte) (int, error)
}

func (f *FSRW) Create(path string) (File, error) {
	d, existingFile, err := f.getEntryWithParent(path, doesntMatter)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "create",
			Path: path,
			Err:  err,
		}
	}

	fileName := filepath.Base(path)

	if existingFile == nil {
		if d.mode&0o222 == 0 {
			return nil, &fs.PathError{
				Op:   "create",
				Path: path,
				Err:  fs.ErrPermission,
			}
		}

		i := &inode{
			modtime: time.Now(),
			mode:    fs.ModePerm,
		}
		d.entries = append(d.entries, &dirEnt{
			directoryEntry: i,
			name:           fileName,
		})
		d.modtime = i.modtime

		return &file{
			name:   fileName,
			inode:  i,
			opMode: opRead | opWrite | opSeek,
		}, nil
	}

	of, err := existingFile.open(fileName, opRead|opWrite|opSeek)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "create",
			Path: path,
			Err:  err,
		}
	}

	ef, ok := of.(*file)
	if !ok {
		return nil, &fs.PathError{
			Op:   "create",
			Path: path,
			Err:  fs.ErrInvalid,
		}
	}

	ef.modtime = time.Now()
	ef.data = ef.data[:0]

	return ef, nil
}

func (f *FSRW) Link(oldPath, newPath string) error {
	oe, err := f.getLEntry(oldPath)
	if err != nil {
		return &fs.PathError{
			Op:   "link",
			Path: oldPath,
			Err:  err,
		}
	} else if oe.IsDir() {
		return &fs.PathError{
			Op:   "link",
			Path: oldPath,
			Err:  fs.ErrInvalid,
		}
	}

	d, _, err := f.getEntryWithParent(newPath, mustNotExist)
	if err != nil {
		return &fs.PathError{
			Op:   "link",
			Path: newPath,
			Err:  err,
		}
	} else if d.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   "link",
			Path: newPath,
			Err:  fs.ErrPermission,
		}
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: oe.directoryEntry,
		name:           filepath.Base(newPath),
	})
	d.modtime = time.Now()

	return nil
}

func (f *FSRW) Symlink(oldPath, newPath string) error {
	d, _, err := f.getEntryWithParent(newPath, mustNotExist)
	if err != nil {
		return &fs.PathError{
			Op:   "symlink",
			Path: newPath,
			Err:  err,
		}
	} else if d.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   "symlink",
			Path: newPath,
			Err:  fs.ErrPermission,
		}
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: &inode{
			data:    []byte(filepath.Clean(oldPath)),
			modtime: time.Now(),
			mode:    fs.ModeSymlink | fs.ModePerm,
		},
		name: filepath.Base(newPath),
	})
	d.modtime = time.Now()

	return nil
}

func (f *FSRW) Rename(oldPath, newPath string) error {
	od, oldFile, err := f.getEntryWithParent(oldPath, mustExist)
	if err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: oldPath,
			Err:  err,
		}
	}

	nd, _, err := f.getEntryWithParent(newPath, mustNotExist)
	if err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  err,
		}
	} else if nd.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  fs.ErrPermission,
		}
	}

	if err := od.remove(oldFile.name); err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  err,
		}
	}

	nd.entries = append(nd.entries, &dirEnt{
		directoryEntry: oldFile.directoryEntry,
		name:           filepath.Base(newPath),
	})
	nd.modtime = time.Now()

	return nil
}

func (f *FSRW) Remove(path string) error {
	d, de, err := f.getEntryWithParent(path, mustExist)
	if err != nil {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  err,
		}
	} else if d.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  fs.ErrPermission,
		}
	}

	if de.IsDir() {
		dir, _ := de.directoryEntry.(*dnode)

		if len(dir.entries) > 0 {
			return &fs.PathError{
				Op:   "remove",
				Path: path,
				Err:  fs.ErrInvalid,
			}
		}
	}

	if err := d.remove(de.name); err != nil {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) RemoveAll(path string) error {
	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return &fs.PathError{
			Op:   "removeall",
			Path: path,
			Err:  err,
		}
	}

	if err := d.remove(fileName); err != nil {
		return &fs.PathError{
			Op:   "removeall",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) LStat(path string) (fs.FileInfo, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "lstat",
			Path: path,
			Err:  err,
		}
	}

	fi, err := de.Info()
	if err != nil {
		return nil, &fs.PathError{
			Op:   "lstat",
			Path: path,
			Err:  err,
		}
	}

	return fi, nil
}

func (f *FSRW) Readlink(path string) (string, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  err,
		}
	}

	s, ok := de.directoryEntry.(*inode)
	if !ok || s.mode&fs.ModeSymlink == 0 {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  fs.ErrInvalid,
		}
	} else if s.mode&0o444 == 0 {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: path,
			Err:  fs.ErrPermission,
		}
	}

	return string(s.data), nil
}

func (f *FSRW) Chown(path string, uid, gid int) error {
	if _, err := f.getEntry(path); err != nil {
		return &fs.PathError{
			Op:   "chown",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) Chmod(path string, mode fs.FileMode) error {
	de, err := f.getEntry(path)
	if err != nil {
		return &fs.PathError{
			Op:   "chmod",
			Path: path,
			Err:  err,
		}
	}

	de.setMode(mode & fs.ModePerm)

	return nil
}

func (f *FSRW) Lchown(path string, uid, gid int) error {
	if _, err := f.getLEntry(path); err != nil {
		return &fs.PathError{
			Op:   "lchown",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) Chtimes(path string, atime time.Time, mtime time.Time) error {
	de, err := f.getEntry(path)
	if err != nil {
		return &fs.PathError{
			Op:   "chtimes",
			Path: path,
			Err:  err,
		}
	}

	de.setTimes(atime, mtime)

	return nil
}

func (f *FSRW) Lchtimes(path string, atime time.Time, mtime time.Time) error {
	de, err := f.getLEntry(path)
	if err != nil {
		return &fs.PathError{
			Op:   "lchtimes",
			Path: path,
			Err:  err,
		}
	}

	de.setTimes(atime, mtime)

	return nil
}

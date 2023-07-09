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
	fsRO
}

func New() *FSRW {
	return &FSRW{
		fsRO: fsRO{
			de: &dnodeRW{
				dnode: dnode{
					mode:    fs.ModeDir | fs.ModePerm,
					modtime: time.Now(),
				},
			},
		},
	}
}

func (f *FSRW) Seal() *fsRO {
	f.mu.Lock()
	defer f.mu.Unlock()

	return &fsRO{
		de: f.de.seal(),
	}
}

func (f *FSRW) ReadDir(path string) ([]fs.DirEntry, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	d, err := f.getDirEnt(path)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  err,
		}
	}

	des, err := d.getEntries()
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: path,
			Err:  err,
		}
	}

	return des, nil
}

func (f *FSRW) ReadFile(path string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fsRO.ReadFile(path)
}

func (f *FSRW) Stat(path string) (fs.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fsRO.Stat(path)
}

func (f *FSRW) Mkdir(path string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

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

	if err := d.setEntry(&dirEnt{
		directoryEntry: &dnodeRW{
			dnode: dnode{
				modtime: time.Now(),
				mode:    fs.ModeDir | perm,
			},
		},
		name: filepath.Base(path),
	}); err != nil {
		return &fs.PathError{
			Op:   op,
			Path: opath,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) MkdirAll(path string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

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
	f.mu.Lock()
	defer f.mu.Unlock()

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
		i := &inodeRW{
			inode: inode{
				modtime: time.Now(),
				mode:    fs.ModePerm,
			},
		}

		if err := d.setEntry(&dirEnt{
			directoryEntry: i,
			name:           fileName,
		}); err != nil {
			return nil, &fs.PathError{
				Op:   "create",
				Path: path,
				Err:  err,
			}
		}

		return &fileRW{
			mu: &i.mu,
			file: file{
				name:   fileName,
				inode:  &i.inode,
				opMode: opRead | opWrite | opSeek,
			},
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

	ef, ok := of.(*fileRW)
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
	f.mu.Lock()
	defer f.mu.Unlock()

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
	}

	if err := d.setEntry(&dirEnt{
		directoryEntry: oe.directoryEntry,
		name:           filepath.Base(newPath),
	}); err != nil {
		return &fs.PathError{
			Op:   "link",
			Path: newPath,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) Symlink(oldPath, newPath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	d, _, err := f.getEntryWithParent(newPath, mustNotExist)
	if err != nil {
		return &fs.PathError{
			Op:   "symlink",
			Path: newPath,
			Err:  err,
		}
	}

	if err = d.setEntry(&dirEnt{
		directoryEntry: &inodeRW{
			inode: inode{
				data:    []byte(filepath.Clean(oldPath)),
				modtime: time.Now(),
				mode:    fs.ModeSymlink | fs.ModePerm,
			},
		},
		name: filepath.Base(newPath),
	}); err != nil {
		return &fs.PathError{
			Op:   "symlink",
			Path: newPath,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) Rename(oldPath, newPath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

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
	} else if nd.Mode()&0o222 == 0 {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  fs.ErrPermission,
		}
	}

	if err := od.removeEntry(oldFile.name); err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  err,
		}
	}

	if err := nd.setEntry(&dirEnt{
		directoryEntry: oldFile.directoryEntry,
		name:           filepath.Base(newPath),
	}); err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: newPath,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) Remove(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	d, de, err := f.getEntryWithParent(path, mustExist)
	if err != nil {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  err,
		}
	}

	if de.IsDir() {
		dir, _ := de.directoryEntry.(dNode)

		if dir.hasEntries() {
			return &fs.PathError{
				Op:   "remove",
				Path: path,
				Err:  fs.ErrInvalid,
			}
		}
	}

	if err := d.removeEntry(de.name); err != nil {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) RemoveAll(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return &fs.PathError{
			Op:   "removeall",
			Path: path,
			Err:  err,
		}
	}

	if err := d.removeEntry(fileName); err != nil {
		return &fs.PathError{
			Op:   "removeall",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FSRW) LStat(path string) (fs.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fsRO.LStat(path)
}

func (f *FSRW) Readlink(path string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fsRO.Readlink(path)
}

func (f *FSRW) Chown(path string, uid, gid int) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

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
	f.mu.RLock()
	defer f.mu.RUnlock()

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
	f.mu.RLock()
	defer f.mu.RUnlock()

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
	f.mu.RLock()
	defer f.mu.RUnlock()

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
	f.mu.RLock()
	defer f.mu.RUnlock()

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

func (f *FSRW) Sub(path string) (fs.FS, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	de, err := f.fsRO.sub(path)
	if err != nil {
		return nil, err
	}

	return &FSRW{
		fsRO: fsRO{
			de: de,
		},
	}, nil
}

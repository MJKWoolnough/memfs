package memfs

import (
	"errors"
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
		mode:    fs.ModeDir | fs.ModePerm,
	}
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

func (f *FS) getDirEnt(path string) (*dnode, error) {
	redirectsRemaining := maxRedirects

	de, err := f.getResolvedDirEnt(f.joinRoot(path), &redirectsRemaining)
	if err != nil {
		return nil, err
	} else if d, ok := de.directoryEntry.(*dnode); !ok {
		return nil, fs.ErrInvalid
	} else {
		return d, nil
	}
}

var maxRedirects uint8 = 255

func (f *FS) getResolvedDirEnt(path string, remainingRedirects *uint8) (*dirEnt, error) {
	var de *dirEnt

	dir, base := filepath.Split(path)
	if dir == "" || dir == "/" {
		de = &dirEnt{
			directoryEntry: (*dnode)(f),
			name:           "/",
		}
	} else {
		var err error

		if de, err = f.getResolvedDirEnt(filepath.Clean(dir), remainingRedirects); err != nil {
			return nil, err
		}
	}

	if mode := de.Mode(); mode&0o444 == 0 {
		return nil, fs.ErrPermission
	} else if !mode.IsDir() {
		return nil, fs.ErrInvalid
	}

	d, _ := de.directoryEntry.(*dnode)

	if base == "" {
		return de, nil
	} else if de = d.get(base); de == nil {
		return nil, fs.ErrNotExist
	} else if de.Mode()&fs.ModeSymlink == 0 {
		return de, nil
	} else if *remainingRedirects == 0 {
		return nil, fs.ErrInvalid
	}

	*remainingRedirects--

	se, _ := de.directoryEntry.(*inode)

	link := string(se.data)

	if !strings.HasPrefix(link, "/") {
		link = filepath.Join(dir, link)
	}

	return f.getResolvedDirEnt(link, remainingRedirects)
}

func (f *FS) getEntry(path string) (*dirEnt, error) {
	redirectsRemaining := maxRedirects

	return f.getResolvedDirEnt(f.joinRoot(path), &redirectsRemaining)
}

func (f *FS) getLEntry(path string) (*dirEnt, error) {
	dirName, fileName := filepath.Split(path)

	redirectsRemaining := maxRedirects

	de, err := f.getResolvedDirEnt(f.joinRoot(dirName), &redirectsRemaining)
	if err != nil {
		return nil, err
	} else if d, ok := de.directoryEntry.(*dnode); !ok {
		return nil, fs.ErrInvalid
	} else if e := d.get(fileName); e == nil {
		return e, fs.ErrNotExist
	} else {
		return e, nil
	}
}

type exists byte

const (
	mustNotExist exists = iota
	mustExist
	doesntMatter
)

func (f *FS) getEntryWithParent(path string, exists exists) (*dnode, *dirEnt, error) {
	parent, child := filepath.Split(path)
	if child == "" {
		return nil, nil, fs.ErrInvalid
	}

	d, err := f.getDirEnt(parent)
	if err != nil {
		return nil, nil, err
	}

	if d.mode&0o444 == 0 {
		return nil, nil, fs.ErrPermission
	}

	c := d.get(child)
	if c == nil && exists == mustExist {
		return nil, nil, fs.ErrNotExist
	} else if c != nil && exists == mustNotExist {
		return nil, nil, fs.ErrExist
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

func (f *FS) ReadFile(path string) ([]byte, error) {
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

func (f *FS) Stat(path string) (fs.FileInfo, error) {
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

func (f *FS) Mkdir(path string, perm fs.FileMode) error {
	return f.mkdir("mkdir", path, path, perm)
}

func (f *FS) mkdir(op, opath, path string, perm fs.FileMode) error {
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
		directoryEntry: &dnode{
			modtime: time.Now(),
			mode:    fs.ModeDir | perm,
		},
		name: filepath.Base(path),
	})
	d.modtime = time.Now()

	return nil
}

func (f *FS) MkdirAll(path string, perm fs.FileMode) error {
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

func (f *FS) Create(path string) (File, error) {
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

func (f *FS) Link(oldPath, newPath string) error {
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

func (f *FS) Symlink(oldPath, newPath string) error {
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

func (f *FS) Rename(oldPath, newPath string) error {
	od, oldFile, err := f.getEntryWithParent(oldPath, mustExist)
	if err != nil {
		return &fs.PathError{
			Op:   "rename",
			Path: oldPath,
			Err:  err,
		}
	} else if od.mode&0o222 == 0 {
		return &fs.PathError{
			Op:   "rename",
			Path: oldPath,
			Err:  fs.ErrPermission,
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

	od.remove(oldFile.name)
	nd.entries = append(nd.entries, &dirEnt{
		directoryEntry: oldFile.directoryEntry,
		name:           filepath.Base(newPath),
	})
	nd.modtime = time.Now()

	return nil
}

func (f *FS) Remove(path string) error {
	d, de, err := f.getEntryWithParent(path, mustExist)
	if err != nil {
		return &fs.PathError{
			Op:   "remove",
			Path: path,
			Err:  err,
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

func (f *FS) RemoveAll(path string) error {
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

func (f *FS) LStat(path string) (fs.FileInfo, error) {
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

	s, _ := de.directoryEntry.(*inode)

	return string(s.data), nil
}

func (f *FS) Chown(path string, uid, gid int) error {
	if _, err := f.getEntry(path); err != nil {
		return &fs.PathError{
			Op:   "chown",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FS) Chmod(path string, mode fs.FileMode) error {
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

func (f *FS) Lchown(path string, uid, gid int) error {
	if _, err := f.getLEntry(path); err != nil {
		return &fs.PathError{
			Op:   "lchown",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

func (f *FS) Chtimes(path string, atime time.Time, mtime time.Time) error {
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

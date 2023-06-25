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
		return nil, err
	}

	_, fileName := filepath.Split(path)

	return de.open(fileName, opRead|opSeek)
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
	} else if mode&fs.ModeSymlink == 0 {
		d, _ := de.directoryEntry.(*dnode)

		if base == "" {
			return de, nil
		}

		de := d.get(base)
		if de == nil {
			return nil, fs.ErrNotExist
		}

		return de, nil
	} else if *remainingRedirects == 0 {
		return nil, fs.ErrInvalid
	}

	*remainingRedirects--

	se, _ := de.directoryEntry.(*inode)

	link := string(se.data)

	if !strings.HasPrefix("/", link) {
		link = filepath.Join(dir, link)
	}

	return f.getResolvedDirEnt(path, remainingRedirects)
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
	} else {
		return d.get(fileName), nil
	}
}

func (f *FS) ReadDir(path string) ([]fs.DirEntry, error) {
	d, err := f.getDirEnt(path)
	if err != nil {
		return nil, err
	}

	if d.mode&0o444 == 0 {
		return nil, fs.ErrPermission
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
		return nil, err
	}

	inode, ok := de.directoryEntry.(*inode)
	if !ok {
		return nil, fs.ErrInvalid
	}

	if inode.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	data := make([]byte, len(inode.data))

	copy(data, inode.data)

	return data, nil
}

func (f *FS) Stat(path string) (fs.FileInfo, error) {
	de, err := f.getEntry(path)
	if err != nil {
		return nil, err
	}

	return de.Info()
}

func (f *FS) Mkdir(path string, perm fs.FileMode) error {
	parent, child := filepath.Split(path)
	if child == "" {
		return fs.ErrInvalid
	}

	d, err := f.getDirEnt(parent)
	if err != nil {
		return err
	}

	if d.mode&0o222 == 0 {
		return fs.ErrPermission
	}

	if d.get(child) != nil {
		return fs.ErrExist
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: &dnode{
			modtime: time.Now(),
			mode:    fs.ModeDir | perm,
		},
		name: child,
	})
	d.modtime = time.Now()

	return nil
}

func (f *FS) MkdirAll(path string, perm fs.FileMode) error {
	path = filepath.Join("/", path)
	last := 0

	for {
		pos := strings.IndexRune(path[last:], filepath.Separator)
		if pos < 0 {
			break
		} else if pos == 0 {
			last++

			continue
		}

		last += pos

		if err := f.Mkdir(path[:last], perm); err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}
	}

	return f.Mkdir(path, perm)
}

type File interface {
	fs.File
	Write([]byte) (int, error)
}

func (f *FS) Create(path string) (File, error) {
	dirName, fileName := filepath.Split(path)
	if fileName == "" {
		return nil, fs.ErrInvalid
	}

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return nil, err
	}

	existingFile := d.get(fileName)
	if existingFile == nil {
		if d.mode&0o222 == 0 {
			return nil, fs.ErrPermission
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
		return nil, err
	}

	ef, ok := of.(*file)
	if !ok {
		return nil, fs.ErrInvalid
	}

	ef.modtime = time.Now()
	ef.data = ef.data[:0]

	return ef, nil
}

func (f *FS) Link(oldPath, newPath string) error {
	oe, err := f.getLEntry(oldPath)
	if err != nil {
		return err
	} else if oe.IsDir() {
		return fs.ErrInvalid
	}

	dirName, fileName := filepath.Split(newPath)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return err
	}

	existingFile := d.get(fileName)
	if existingFile != nil {
		return fs.ErrExist
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: oe,
		name:           fileName,
	})
	d.modtime = time.Now()

	return nil
}

func (f *FS) Symlink(oldPath, newPath string) error {
	dirName, fileName := filepath.Split(newPath)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return err
	}

	existingFile := d.get(fileName)
	if existingFile != nil {
		return fs.ErrExist
	}

	d.entries = append(d.entries, &dirEnt{
		directoryEntry: &inode{
			data:    []byte(filepath.Clean(oldPath)),
			modtime: time.Now(),
			mode:    fs.ModeSymlink | fs.ModePerm,
		},
		name: fileName,
	})
	d.modtime = time.Now()

	return nil
}

func (f *FS) Rename(oldPath, newPath string) error {
	oldDirName, oldFileName := filepath.Split(oldPath)

	od, err := f.getDirEnt(oldDirName)
	if err != nil {
		return err
	}

	oldFile := od.get(oldFileName)
	if oldFile == nil {
		return fs.ErrNotExist
	}

	newDirName, newFileName := filepath.Split(newPath)

	nd, err := f.getDirEnt(newDirName)
	if err != nil {
		return err
	}

	if nd.get(oldFileName) != nil {
		return fs.ErrExist
	}

	od.remove(oldFileName)
	nd.entries = append(nd.entries, &dirEnt{
		directoryEntry: oldFile.directoryEntry,
		name:           newFileName,
	})
	nd.modtime = time.Now()

	return nil
}

func (f *FS) Remove(path string) error {
	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return err
	}

	de := d.get(fileName)
	if de.IsDir() {
		dir, _ := de.directoryEntry.(*dnode)

		if len(dir.entries) > 0 {
			return fs.ErrInvalid
		}
	}

	return d.remove(fileName)
}

func (f *FS) RemoveAll(path string) error {
	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return err
	}

	return d.remove(fileName)
}

func (f *FS) LStat(path string) (fs.FileInfo, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return nil, err
	}

	return de.Info()
}

func (f *FS) Readlink(path string) (string, error) {
	de, err := f.getLEntry(path)
	if err != nil {
		return "", err
	}

	if de.Mode()&fs.ModeSymlink == 0 {
		return "", fs.ErrInvalid
	}

	s, _ := de.directoryEntry.(*inode)

	return string(s.data), nil
}

func (f *FS) Chown(path string, uid, gid int) error {
	_, err := f.getEntry(path)

	return err
}

func (f *FS) Chmod(path string, mode fs.FileMode) error {
	de, err := f.getEntry(path)
	if err != nil {
		return err
	}

	de.setMode(mode & fs.ModePerm)

	return nil
}

func (f *FS) Lchown(path string, uid, gid int) error {
	_, err := f.getLEntry(path)

	return err
}

func (f *FS) Chtimes(path string, atime time.Time, mtime time.Time) error {
	de, err := f.getEntry(path)
	if err != nil {
		return err
	}

	de.setTimes(atime, mtime)

	return nil
}

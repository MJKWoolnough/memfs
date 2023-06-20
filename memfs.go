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
	de, err := f.getEntry(path)
	if err != nil {
		return nil, err
	}

	_, fileName := filepath.Split(path)

	return de.open(fileName, opRead|opSeek)
}

func (f *FS) getDirEnt(path string) (*dnode, error) {
	d := (*dnode)(f)
	for _, p := range strings.Split(path, separator) {
		if p == "" {
			continue
		}

		if d.mode&0o440 == 0 {
			return nil, fs.ErrPermission
		}

		de := d.get(p)

		if de, ok := de.directoryEntry.(*directory); ok {
			d = de.dnode
		} else {
			return nil, fs.ErrNotExist
		}
	}

	return d, nil
}

var maxRedirects uint8 = 100

func (f *FS) getResolvedDirEnt(path string, remainingRedirects *uint8) (*dirEnt, error) {
	var de *dirEnt

	dir, base := filepath.Split(path)
	if dir == "" {
		if de = (*dnode)(f).get(base); de == nil {
			return nil, fs.ErrNotExist
		}
	} else {
		var err error

		if de, err = f.getResolvedDirEnt(dir, remainingRedirects); err != nil {
			return nil, err
		}
	}

	if mode := de.Mode(); mode&0o444 == 0 {
		return nil, fs.ErrPermission
	} else if !mode.IsDir() {
		return nil, fs.ErrInvalid
	} else if mode&fs.ModeSymlink == 0 {
		d, _ := de.directoryEntry.(*dnode)

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
	for i := 0; i < 100; i++ {
		de, err := f.getLEntry(path)
		if err != nil {
			return nil, err
		}

		if de.Mode()&fs.ModeSymlink == 0 {
			return de, nil
		}

		f, _ := de.directoryEntry.(*inode)

		path = string(f.data)
	}

	return nil, fs.ErrInvalid
}

func (f *FS) getLEntry(path string) (*dirEnt, error) {
	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return nil, err
	}

	if d.mode&0o440 == 0 {
		return nil, fs.ErrPermission
	}

	return d.get(fileName), nil
}

func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := f.getDirEnt(name)
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

func (f *FS) ReadFile(name string) ([]byte, error) {
	de, err := f.getEntry(name)
	if err != nil {
		return nil, err
	}

	inode, ok := de.directoryEntry.(*inode)
	if !ok {
		return nil, fs.ErrInvalid
	}

	data := make([]byte, len(inode.data))

	copy(data, inode.data)

	return data, nil
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	de, err := f.getEntry(name)
	if err != nil {
		return nil, err
	}

	return de.Info()
}

func (f *FS) Sub(dir string) (fs.FS, error) {
	dn, err := f.getDirEnt(dir)
	if err != nil {
		return nil, err
	}

	if dn.mode&0o110 == 0 {
		return nil, fs.ErrPermission
	}

	return (*FS)(dn), nil
}

func (f *FS) Mkdir(name string, perm fs.FileMode) error {
	parent, child := filepath.Split(name)
	d, err := f.getDirEnt(parent)
	if err != nil {
		return err
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

func (f *FS) Create(path string) (File, error) {
	dirName, fileName := filepath.Split(path)

	d, err := f.getDirEnt(dirName)
	if err != nil {
		return nil, err
	}

	existingFile := d.get(fileName)
	if existingFile == nil {
		i := &inode{
			modtime: time.Now(),
			mode:    0o777,
		}
		d.entries = append(d.entries, &dirEnt{
			directoryEntry: i,
			name:           fileName,
		})

		return &file{
			name:   fileName,
			inode:  i,
			opMode: opWrite | opSeek,
		}, nil
	}

	of, err := existingFile.open(fileName, opWrite|opSeek)
	if err != nil {
		return nil, err
	}

	ef, ok := of.(*file)
	if !ok {
		return nil, fs.ErrInvalid
	}

	ef.data = ef.data[:0]

	return ef, nil
}

func (f *FS) Link(oldname, newname string) error {
	oe, err := f.getLEntry(oldname)
	if err != nil {
		return err
	} else if oe.IsDir() {
		return fs.ErrInvalid
	}

	dirName, fileName := filepath.Split(newname)

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

	return nil
}

func (f *FS) Symlink(oldname, newname string) error {
	dirName, fileName := filepath.Split(newname)

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
			data:    []byte(oldname),
			modtime: time.Now(),
			mode:    0o777,
		},
		name: fileName,
	})

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

func (f *FS) LStat(name string) (fs.FileInfo, error) {
	de, err := f.getLEntry(name)
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

func (f *FS) Chown(name string, uid, gid int) error {
	return nil
}

func (f *FS) Chmod(name string, mode fs.FileMode) error {
	de, err := f.getEntry(name)
	if err != nil {
		return err
	}

	de.setMode(mode)

	return nil
}

func (f *FS) Lchown(name string, uid, gid int) error {
	return nil
}

func (f *FS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	de, err := f.getEntry(name)
	if err != nil {
		return err
	}

	de.setTimes(atime, mtime)

	return nil
}

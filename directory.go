package memfs

import (
	"io"
	"io/fs"
	"time"
)

type directoryEntry interface {
	IsDir() bool
	ModTime() time.Time
	Type() fs.FileMode
	Mode() fs.FileMode
	Size() int64
	open(name string, mode opMode) (fs.File, error)
	setMode(fs.FileMode)
	setTimes(time.Time, time.Time)
}

type dNode interface {
	getEntry(string) (*dirEnt, error)
	setEntry(*dirEnt) error
	hasEntries() bool
	getEntries() ([]fs.DirEntry, error)
	removeEntry(string) error
	fs.FileInfo
}

type dirEnt struct {
	directoryEntry
	name string
}

func (d *dirEnt) Info() (fs.FileInfo, error) {
	return d, nil
}

func (d *dirEnt) Name() string {
	return d.name
}

func (d *dirEnt) Sys() any {
	return d.directoryEntry
}

type dnode struct {
	entries []*dirEnt
	modtime time.Time
	mode    fs.FileMode
}

func (d *dnode) open(name string, _ opMode) (fs.File, error) {
	if d.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	return &directory{
		dnode: d,
		name:  name,
	}, nil
}

func (d *dnode) getEntry(name string) (*dirEnt, error) {
	if d.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	for _, de := range d.entries {
		if de.name == name {
			return de, nil
		}
	}

	return nil, fs.ErrNotExist
}

func (d *dnode) setEntry(de *dirEnt) error {
	if d.mode&0o222 == 0 {
		return fs.ErrPermission
	}

	d.entries = append(d.entries, de)
	d.modtime = time.Now()

	return nil
}

func (d *dnode) hasEntries() bool {
	return len(d.entries) > 0
}

func (d *dnode) getEntries() ([]fs.DirEntry, error) {
	if d.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	dirs := make([]fs.DirEntry, len(d.entries))

	for i := range d.entries {
		dirs[i] = d.entries[i]
	}

	return dirs, nil
}

func (d *dnode) removeEntry(name string) error {
	if d.mode&0o222 == 0 {
		return fs.ErrPermission
	}

	for n, de := range d.entries {
		if de.name == name {
			d.entries = append(d.entries[:n], d.entries[n+1:]...)
			d.modtime = time.Now()

			return nil
		}
	}

	return fs.ErrNotExist
}

func (d *dnode) setMode(mode fs.FileMode) {
	d.mode = fs.ModeDir | mode
}

func (d *dnode) setTimes(_, mtime time.Time) {
	d.modtime = mtime
}

type directory struct {
	*dnode
	name string
	pos  int
}

func (d *directory) Info() (fs.FileInfo, error) {
	return d, nil
}

func (d *directory) Stat() (fs.FileInfo, error) {
	return d, nil
}

func (d *directory) Read(_ []byte) (int, error) {
	return 0, &fs.PathError{
		Op:   "read",
		Path: d.name,
		Err:  fs.ErrInvalid,
	}
}

func (d *directory) Close() error {
	return nil
}

func (d *directory) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		return d.getEntries()
	} else if d.mode&0o444 == 0 {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: d.name,
			Err:  fs.ErrPermission,
		}
	}

	left := len(d.entries) - d.pos

	if left < n {
		n = left
	}

	if n == 0 {
		return nil, io.EOF
	}

	dirs := make([]fs.DirEntry, n)

	for i := range dirs {
		dirs[i] = d.entries[d.pos]
		d.pos++
	}

	return dirs, nil
}

func (d *directory) Name() string {
	return d.name
}

func (d *dnode) Size() int64 {
	return 0
}

func (d *dnode) Type() fs.FileMode {
	return d.mode
}

func (d *dnode) Mode() fs.FileMode {
	return d.mode
}

func (d *dnode) ModTime() time.Time {
	return d.modtime
}

func (d *dnode) IsDir() bool {
	return true
}

func (d *directory) Sys() any {
	return d
}

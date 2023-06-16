package memfs

import (
	"io/fs"
	"time"
)

type directoryEntry interface {
	IsDir() bool
	ModTime() time.Time
	Type() fs.FileMode
	Mode() fs.FileMode
	Size() int64
	open(name string, mode opMode) fs.File
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
	name    string
	entries []*dirEnt
	modtime time.Time
	mode    fs.FileMode
}

func (d *dnode) open(_ string, _ opMode) fs.File {
	return &directory{
		dnode: d,
	}
}

func (d *dnode) get(name string) *dirEnt {
	for _, de := range d.entries {
		if de.name == name {
			return de
		}
	}

	return nil
}

type directory struct {
	*dnode
	pos int
}

func (d *directory) Info() (fs.FileInfo, error) {
	return d, nil
}

func (d *directory) Stat() (fs.FileInfo, error) {
	return d, nil
}

func (d *directory) Read(_ []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (d *directory) Close() error {
	return nil
}

func (d *directory) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		dirs := make([]fs.DirEntry, len(d.entries))

		for i := range d.entries {
			dirs[i] = d.entries[i]
		}

		return dirs, nil
	}

	left := len(d.entries) - d.pos

	if left < n {
		n = left
	}

	dirs := make([]fs.DirEntry, 0, n)

	for i := range d.entries {
		dirs[i] = d.entries[i]
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

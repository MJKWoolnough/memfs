package memfs

import (
	"io/fs"
	"time"
)

type direntry struct {
	name    string
	entries []fs.DirEntry
	modtime time.Time
	mode    fs.FileMode
}

type directory struct {
	*direntry
	pos int
}

func (d *directory) Stat() (fs.FileInfo, error) {
	return d, nil
}

func (d *directory) Read(_ []byte) (int, error) {
	return 0, nil
}

func (d *directory) Close() error {
	return nil
}

func (d *directory) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		dirs := make([]fs.DirEntry, len(d.entries))

		copy(dirs, d.entries)

		return dirs, nil
	}

	left := len(d.entries) - d.pos

	if left < n {
		n = left
	}

	dirs := make([]fs.DirEntry, n)

	d.pos += copy(dirs, d.entries[d.pos:])

	return dirs, nil
}

func (d *directory) Name() string {
	return d.name
}

func (d *directory) Size() int64 {
	return 0
}

func (d *directory) Type() fs.FileMode {
	return d.mode
}

func (d *directory) Mode() fs.FileMode {
	return d.mode
}

func (d *directory) ModTime() time.Time {
	return d.modtime
}

func (d *directory) IsDir() bool {
	return true
}

func (d *directory) Sys() any {
	return d
}

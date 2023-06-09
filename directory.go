package memfs

import (
	"io/fs"
	"time"
)

type directory struct {
	name    string
	entries []fs.DirEntry
	modtime time.Time
	mode    fs.FileMode
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
	return nil, nil
}

func (d *directory) Name() string {
	return d.name
}

func (d *directory) Size() int64 {
	return 0
}

func (d *directory) Mode() fs.FileMode {
	return 0
}

func (d *directory) ModTime() time.Time {
	return time.Time{}
}

func (d *directory) IsDir() bool {
	return true
}

func (d *directory) Sys() any {
	return d
}

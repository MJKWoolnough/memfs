package memfs

import (
	"io/fs"
	"sync"
	"time"
)

type dnodeRW struct {
	dnode
	mu sync.RWMutex
}

func (d *dnodeRW) open(name string, _ opMode) (fs.File, error) {
	if d.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	return &directoryRW{
		mu: &d.mu,
		directory: directory{
			dnode: &d.dnode,
			name:  name,
		},
	}, nil
}

func (d *dnodeRW) get(name string) *dirEnt {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.dnode.get(name)
}

func (d *dnodeRW) remove(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

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

func (d *dnodeRW) setMode(mode fs.FileMode) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = fs.ModeDir | mode
}

func (d *dnodeRW) setTimes(_, mtime time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.modtime = mtime
}

func (d *dnodeRW) Type() fs.FileMode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.mode
}

func (d *dnodeRW) Mode() fs.FileMode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.mode
}

func (d *dnodeRW) ModTime() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.modtime
}

type directoryRW struct {
	directory
	mu *sync.RWMutex
}

func (d *directoryRW) ReadDir(n int) ([]fs.DirEntry, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.directory.ReadDir(n)
}

func (d *directoryRW) Sys() any {
	return d
}

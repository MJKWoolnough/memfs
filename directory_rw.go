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
	if d.mode&modeRead == 0 {
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

func (d *dnodeRW) getEntry(name string) (*dirEnt, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.dnode.getEntry(name)
}

func (d *dnodeRW) setEntry(de *dirEnt) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dnode.setEntry(de)
}

func (d *dnodeRW) hasEntries() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.dnode.hasEntries()
}

func (d *dnodeRW) getEntries() ([]fs.DirEntry, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.dnode.getEntries()
}

func (d *dnodeRW) removeEntry(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dnode.removeEntry(name)
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

func (d *dnodeRW) seal() directoryEntry {
	d.mu.Lock()
	defer d.mu.Unlock()

	de := d.dnode

	for n, e := range de.entries {
		de.entries[n].directoryEntry = e.seal()
	}

	d.dnode = dnode{}

	return &de
}

func (d *dnodeRW) Type() fs.FileMode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.mode.Type()
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
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.directory.ReadDir(n)
}

func (d *directoryRW) Sys() any {
	return d
}

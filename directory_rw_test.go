package memfs

import (
	"errors"
	"io"
	"io/fs"
	"reflect"
	"sync"
	"testing"
)

var _ dNode = &directoryRW{}

func testReadAllRW(d *directoryRW) ([]fs.DirEntry, error) {
	return d.ReadDir(-1)
}

func testReadOnesRW(d *directoryRW) ([]fs.DirEntry, error) {
	var read []fs.DirEntry

	for {
		des, err := d.ReadDir(1)
		if errors.Is(err, io.EOF) {
			return read, nil
		} else if err != nil {
			return nil, err
		}

		read = append(read, des...)
	}
}

func testReadTwosRW(d *directoryRW) ([]fs.DirEntry, error) {
	var read []fs.DirEntry

	for {
		des, err := d.ReadDir(2)
		if errors.Is(err, io.EOF) {
			return read, nil
		} else if err != nil {
			return nil, err
		}

		read = append(read, des...)
	}
}

func makeDirectoryRW(dirs []*dirEnt) *directoryRW {
	return &directoryRW{
		mu: &sync.RWMutex{},
		directory: directory{
			dnode: &dnode{
				entries: dirs,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
	}
}

func TestReadDirRW(t *testing.T) {
	for n, test := range [...][]*dirEnt{
		{
			{
				name: "1",
			},
		},
		{
			{
				name: "1",
			},
			{
				name: "2",
			},
		},
		{
			{
				name: "1",
			},
			{
				name: "3",
			},
			{
				name: "3",
			},
		},
		{
			{
				name: "1",
			},
			{
				name: "2",
			},
			{
				name: "3",
			},
			{
				name: "4",
			},
		},
	} {
		if all, err := testReadAllRW(makeDirectoryRW(test)); err != nil {
			t.Errorf("test %d.1: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), all) {
			t.Errorf("test %d.1: 'all' does not equal", n+1)
		}

		if ones, err := testReadOnesRW(makeDirectoryRW(test)); err != nil {
			t.Errorf("test %d.2: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), ones) {
			t.Errorf("test %d.2: 'ones' does not equal", n+1)
		}

		if twos, err := testReadTwosRW(makeDirectoryRW(test)); err != nil {
			t.Errorf("test %d.3: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), twos) {
			t.Errorf("test %d.3: 'twos' does not equal", n+1)
		}
	}
}

func TestDnodeRemoveRW(t *testing.T) {
	d := dnodeRW{
		dnode: dnode{
			entries: []*dirEnt{
				{
					name: "1",
				},
				{
					name: "2",
				},
				{
					name: "3",
				},
				{
					name: "4",
				},
			},
			mode: fs.ModeDir | fs.ModePerm,
		},
	}

	if err := d.removeEntry("2"); err != nil {
		t.Errorf("test 1: unexpected error: %s", err)

		return
	}

	if err := d.removeEntry("2"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("test 2: unexpected Not Exist, got %s", err)

		return
	}

	if len(d.entries) != 3 {
		t.Errorf("test 3: expecting 3 entries, got %d", len(d.entries))

		return
	}

	expecting := []*dirEnt{
		{
			name: "1",
		},
		{
			name: "3",
		},
		{
			name: "4",
		},
	}

	if !reflect.DeepEqual(expecting, d.entries) {
		t.Errorf("test 4: expecting %v, got %v", expecting, d.entries)
	}
}

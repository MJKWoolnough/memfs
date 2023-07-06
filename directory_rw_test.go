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

func TestDnodeGetRW(t *testing.T) {
	d := dnodeRW{
		dnode: dnode{
			entries: []*dirEnt{
				{
					name: "1",
				},
				{
					name: "2",
					directoryEntry: &dnode{
						mode: fs.ModePerm,
					},
				},
				{
					name: "3",
					directoryEntry: &dnode{
						mode: 0o222,
					},
				},
				{
					name: "4",
					directoryEntry: &dnode{
						mode: 0o444,
					},
				},
			},
		},
	}

	if got, err := d.getEntry("1"); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("test 1: expecting to err %v, got %v", fs.ErrPermission, err)
	} else if got != nil {
		t.Errorf("test 1: expecting to get nil, got %v", got)
	}

	if got, err := d.getEntry("2"); err != nil {
		t.Errorf("test 2: expecting to nil err, got %v", err)
	} else if got == nil || got.name != "2" {
		t.Errorf("test 2: expecting to get '2', got %v", got)
	}

	if got, err := d.getEntry("3"); err != nil {
		t.Errorf("test 3: expecting to nil err, got %v", err)
	} else if got == nil || got.name != "3" {
		t.Errorf("test 3: expecting to get '3', got %v", got)
	}

	if got, err := d.getEntry("4"); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("test 4: expecting to err %v, got %v", fs.ErrPermission, err)
	} else if got != nil {
		t.Errorf("test 4: expecting to get nil, got %v", got)
	}

	if got, err := d.getEntry("5"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("test 5: expecting to err %v, got %v", fs.ErrNotExist, err)
	} else if got != nil {
		t.Errorf("test 5: expecting to get nil, got %v", got)
	}
}

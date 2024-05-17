package memfs

import (
	"errors"
	"io"
	"io/fs"
	"reflect"
	"testing"
)

var _ dNode = &directory{}

func testReadAll(d *directory) ([]fs.DirEntry, error) {
	return d.ReadDir(-1)
}

func testReadOnes(d *directory) ([]fs.DirEntry, error) {
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

func testReadTwos(d *directory) ([]fs.DirEntry, error) {
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

func makeDirectory(dirs []*dirEnt) *directory {
	return &directory{
		dnode: &dnode{
			entries: dirs,
			mode:    fs.ModeDir | fs.ModePerm,
		},
	}
}

func asDirEntries(dirs []*dirEnt) []fs.DirEntry {
	de := make([]fs.DirEntry, len(dirs))

	for n, d := range dirs {
		de[n] = d
	}

	return de
}

func TestReadDir(t *testing.T) {
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
		if all, err := testReadAll(makeDirectory(test)); err != nil {
			t.Errorf("test %d.1: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), all) {
			t.Errorf("test %d.1: 'all' does not equal", n+1)
		}

		if ones, err := testReadOnes(makeDirectory(test)); err != nil {
			t.Errorf("test %d.2: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), ones) {
			t.Errorf("test %d.2: 'ones' does not equal", n+1)
		}

		if twos, err := testReadTwos(makeDirectory(test)); err != nil {
			t.Errorf("test %d.3: received unexpected error: %s", n+1, err)
		} else if !reflect.DeepEqual(asDirEntries(test), twos) {
			t.Errorf("test %d.3: 'twos' does not equal", n+1)
		}
	}
}

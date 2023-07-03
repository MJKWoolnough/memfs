package memfs

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"time"
)

func TestOpen(t *testing.T) {
	for n, test := range [...]struct {
		FS   FS
		Path string
		File fs.File
		Err  error
	}{
		{ // 1
			FS:   FS{},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "open",
				Path: "/file",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "open",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 3
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "file",
					},
				},
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "open",
				Path: "/file",
				Err:  fs.ErrPermission,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			File: &file{
				name: "file",
				inode: &inode{
					mode: fs.ModeDir | fs.ModePerm,
				},
				opMode: opRead | opSeek,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "otherFile",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "open",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										mode: fs.ModeDir | fs.ModePerm,
									},
									name: "deepFile",
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir/deepFile",
			File: &file{
				name: "deepFile",
				inode: &inode{
					mode: fs.ModeDir | fs.ModePerm,
				},
				opMode: opRead | opSeek,
			},
		},
	} {
		f, err := test.FS.Open(test.Path)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !reflect.DeepEqual(f, test.File) {
			t.Errorf("test %d: expected file %v, got %v", n+1, test.File, f)
		}
	}
}

func TestFSReadDir(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output []fs.DirEntry
		Err    error
	}{
		{ // 1
			FS: FS{},
			Err: &fs.PathError{
				Op:   "readdir",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: []fs.DirEntry{},
		},
		{ // 3
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 0),
							mode:    2,
						},
						name: "test",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: []fs.DirEntry{
				&dirEnt{
					directoryEntry: &inode{
						modtime: time.Unix(1, 0),
						mode:    2,
					},
					name: "test",
				},
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 0),
							mode:    2,
						},
						name: "test",
					},
					{
						directoryEntry: &inode{
							modtime: time.Unix(3, 0),
							mode:    4,
						},
						name: "test2",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: []fs.DirEntry{
				&dirEnt{
					directoryEntry: &inode{
						modtime: time.Unix(1, 0),
						mode:    2,
					},
					name: "test",
				},
				&dirEnt{
					directoryEntry: &inode{
						modtime: time.Unix(3, 0),
						mode:    4,
					},
					name: "test2",
				},
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 0),
							mode:    2,
						},
						name: "test",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(3, 0),
										mode:    4,
									},
									name: "test3",
								},
							},
							modtime: time.Unix(5, 0),
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "test2",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/",
			Output: []fs.DirEntry{
				&dirEnt{
					directoryEntry: &inode{
						modtime: time.Unix(1, 0),
						mode:    2,
					},
					name: "test",
				},
				&dirEnt{
					directoryEntry: &dnode{
						entries: []*dirEnt{
							{
								directoryEntry: &inode{
									modtime: time.Unix(3, 0),
									mode:    4,
								},
								name: "test3",
							},
						},
						modtime: time.Unix(5, 0),
						mode:    fs.ModeDir | fs.ModePerm,
					},
					name: "test2",
				},
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 0),
							mode:    2,
						},
						name: "test",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(3, 0),
										mode:    4,
									},
									name: "test3",
								},
							},
							modtime: time.Unix(5, 0),
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "test2",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/test2",
			Output: []fs.DirEntry{
				&dirEnt{
					directoryEntry: &inode{
						modtime: time.Unix(3, 0),
						mode:    4,
					},
					name: "test3",
				},
			},
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 0),
							mode:    2,
						},
						name: "test",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(3, 0),
										mode:    4,
									},
									name: "test3",
								},
							},
							modtime: time.Unix(5, 0),
							mode:    fs.ModeDir,
						},
						name: "test2",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/test2",
			Err: &fs.PathError{
				Op:   "readdir",
				Path: "/test2",
				Err:  fs.ErrPermission,
			},
		},
	} {
		de, err := test.FS.ReadDir(test.Path)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !reflect.DeepEqual(test.Output, de) {
			t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, de)
		}
	}
}

func TestReadFile(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output []byte
		Err    error
	}{
		{ // 1
			FS: FS{},
			Err: &fs.PathError{
				Op:   "readfile",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "readfile",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "readfile",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "notFile",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "readfile",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "readfile",
				Path: "/file",
				Err:  fs.ErrPermission,
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							data: []byte("DATA"),
							mode: fs.ModePerm,
						},
						name: "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/file",
			Output: []byte("DATA"),
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							data: []byte("DATA"),
							mode: fs.ModePerm,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										data: []byte("MORE DATA"),
										mode: fs.ModePerm,
									},
									name: "file2",
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "DIR",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/DIR/file2",
			Output: []byte("MORE DATA"),
		},
	} {
		data, err := test.FS.ReadFile(test.Path)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !bytes.Equal(test.Output, data) {
			t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, data)
		}
	}
}

func TestStat(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output fs.FileInfo
		Err    error
	}{
		{ // 1
			FS: FS{},
			Err: &fs.PathError{
				Op:   "stat",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				modtime: time.Unix(1, 2),
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: &dirEnt{
				directoryEntry: &dnode{
					modtime: time.Unix(1, 2),
					mode:    fs.ModeDir | fs.ModePerm,
				},
				name: "/",
			},
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "stat",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "notFile",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "stat",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: &dirEnt{
				directoryEntry: &inode{
					modtime: time.Unix(1, 2),
					mode:    3,
				},
				name: "file",
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							modtime: time.Unix(4, 5),
							mode:    6,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: &dirEnt{
				directoryEntry: &dnode{
					modtime: time.Unix(4, 5),
					mode:    6,
				},
				name: "dir",
			},
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(4, 5),
										mode:    6,
									},
									name: "anotherFile",
								},
							},
							modtime: time.Unix(7, 8),
							mode:    9,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir/anotherFile",
			Err: &fs.PathError{
				Op:   "stat",
				Path: "/dir/anotherFile",
				Err:  fs.ErrPermission,
			},
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(4, 5),
										mode:    6,
									},
									name: "anotherFile",
								},
							},
							modtime: time.Unix(7, 8),
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir/anotherFile",
			Output: &dirEnt{
				directoryEntry: &inode{
					modtime: time.Unix(4, 5),
					mode:    6,
				},
				name: "anotherFile",
			},
		},
	} {
		stat, err := test.FS.Stat(test.Path)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !reflect.DeepEqual(test.Output, stat) {
			t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, stat)
		}
	}
}

func TestMkdir(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS        FS
		Path      string
		PathPerms fs.FileMode
		Output    FS
		Err       error
	}{
		{ // 1
			FS:     FS{},
			Output: FS{},
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 2
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Path: "/",
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "/",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 4
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a",
		},
		{ // 5
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "/a/b",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 6
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "/a/b",
				Err:  fs.ErrPermission,
			},
		},
		{ // 7
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdir",
				Path: "/a/b",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 8
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										modtime: now,
										mode:    fs.ModeDir | 0o123,
									},
									name: "b",
								},
							},
							modtime: now,
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:      "/a/b",
			PathPerms: 0o123,
		},
	} {
		if err := test.FS.Mkdir(test.Path, test.PathPerms); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func withinRange(dt time.Duration) bool {
	return dt > -10*time.Second && dt < 10*time.Second
}

func fixTimes(d *dnode, now time.Time) {
	if withinRange(d.modtime.Sub(now)) {
		d.modtime = now
	}

	for _, e := range d.entries {
		if de, ok := e.directoryEntry.(*dnode); ok {
			fixTimes(de, now)
		} else if f, ok := e.directoryEntry.(*inode); ok && withinRange(f.modtime.Sub(now)) {
			f.modtime = now
		}
	}
}

func TestMkdirAll(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS        FS
		Path      string
		PathPerms fs.FileMode
		Output    FS
		Err       error
	}{
		{ // 1
			FS:     FS{},
			Output: FS{},
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 2
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Path: "/",
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "/",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 4
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a",
		},
		{ // 5
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "/a/b",
				Err:  fs.ErrPermission,
			},
		},
		{ // 6
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "/a/b",
				Err:  fs.ErrPermission,
			},
		},
		{ // 7
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "mkdirall",
				Path: "/a/b",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 8
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now,
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										modtime: now,
										mode:    fs.ModeDir | 0o123,
									},
									name: "b",
								},
							},
							modtime: now,
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:      "/a/b",
			PathPerms: 0o123,
		},
		{ // 9
			FS: FS{
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										modtime: now,
										mode:    fs.ModeDir | 0o765,
									},
									name: "b",
								},
							},
							modtime: now,
							mode:    fs.ModeDir | 0o765,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:      "/a/b",
			PathPerms: 0o765,
		},
	} {
		if err := test.FS.MkdirAll(test.Path, test.PathPerms); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func TestCreate(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS         FS
		Path       string
		PathPerms  fs.FileMode
		OutputFS   FS
		OutputFile fs.File
		Err        error
	}{
		{ // 1
			FS:       FS{},
			OutputFS: FS{},
			Err: &fs.PathError{
				Op:   "create",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "create",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS:       FS{},
			OutputFS: FS{},
			Path:     "/a",
			Err: &fs.PathError{
				Op:   "create",
				Path: "/a",
				Err:  fs.ErrPermission,
			},
		},
		{ // 4
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/",
			Err: &fs.PathError{
				Op:   "create",
				Path: "/",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 5
			FS: FS{
				modtime: now.Add(-20 * time.Second),
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFile: &file{
				inode: &inode{
					modtime: now,
					mode:    fs.ModePerm,
				},
				name:   "a",
				opMode: opRead | opWrite | opSeek,
			},
			Path: "/a",
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							data:    []byte("Hello"),
							modtime: now.Add(-20 * time.Second),
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				modtime: now.Add(-20 * time.Second),
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							data:    ([]byte("Hello"))[:0],
							modtime: now,
							mode:    fs.ModePerm,
						},
						name: "a",
					},
				},
				modtime: now.Add(-20 * time.Second),
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFile: &file{
				inode: &inode{
					data:    ([]byte("Hello"))[:0],
					modtime: now,
					mode:    fs.ModePerm,
				},
				name:   "a",
				opMode: opRead | opWrite | opSeek,
			},
			Path: "/a",
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now.Add(-20 * time.Second),
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: now,
										mode:    fs.ModePerm,
									},
									name: "b",
								},
							},
							modtime: now,
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "a",
					},
				},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFile: &file{
				inode: &inode{
					modtime: now,
					mode:    fs.ModePerm,
				},
				name:   "b",
				opMode: opRead | opWrite | opSeek,
			},
			Path: "/a/b",
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now.Add(-20 * time.Second),
							mode:    fs.ModeDir | 0o444,
						},
						name: "a",
					},
				},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			OutputFS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							modtime: now.Add(-20 * time.Second),
							mode:    fs.ModeDir | 0o444,
						},
						name: "a",
					},
				},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err: &fs.PathError{
				Op:   "create",
				Path: "/a/b",
				Err:  fs.ErrPermission,
			},
		},
	} {
		if f, err := test.FS.Create(test.Path); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.OutputFile, f) {
				t.Errorf("test %d: expecting to get file %v, got %v", n+1, test.OutputFile, f)
			} else if !reflect.DeepEqual(test.OutputFS, test.FS) {
				t.Errorf("test %d: expecting to get FS %v, got %v", n+1, test.OutputFS, test.FS)
			}
		}
	}
}

func TestLink(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS       FS
		From, To string
		Output   FS
		Err      error
	}{
		{ // 1
			FS:     FS{},
			Output: FS{},
			Err: &fs.PathError{
				Op:   "link",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "link",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			Err: &fs.PathError{
				Op:   "link",
				Path: "/a",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			Err: &fs.PathError{
				Op:   "link",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			To:   "/a",
			Err: &fs.PathError{
				Op:   "link",
				Path: "/a",
				Err:  fs.ErrExist,
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir,
			},
			From: "/a",
			To:   "/b",
			Err: &fs.PathError{
				Op:   "link",
				Path: "/a",
				Err:  fs.ErrPermission,
			},
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode:    fs.ModeDir | fs.ModePerm,
				modtime: now,
			},
			From: "/a",
			To:   "/b",
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Hello"),
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Hello"),
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode:    fs.ModeDir | fs.ModePerm,
				modtime: now,
			},
			From: "/a/b",
			To:   "/c",
		},
		{ // 9
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "c",
									directoryEntry: &inode{
										data: []byte("Hello"),
									},
								},
							},
							mode:    fs.ModeDir | fs.ModePerm,
							modtime: now,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			To:   "/b/c",
		},
	} {
		if err := test.FS.Link(test.From, test.To); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get FS %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func TestSymlink(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS       FS
		From, To string
		Output   FS
		Err      error
	}{
		{ // 1
			FS:     FS{},
			Output: FS{},
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name:           "a",
						directoryEntry: &inode{},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			To:   "/a",
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "/a",
				Err:  fs.ErrExist,
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir,
			},
			From: "/a",
			To:   "/b",
			Err: &fs.PathError{
				Op:   "symlink",
				Path: "/b",
				Err:  fs.ErrPermission,
			},
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &inode{
							data:    []byte("/a"),
							modtime: now,
							mode:    fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode:    fs.ModeDir | fs.ModePerm,
				modtime: now,
			},
			From: "/a",
			To:   "/b",
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Hello"),
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Hello"),
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data:    []byte("/a/b"),
							mode:    fs.ModeSymlink | fs.ModePerm,
							modtime: now,
						},
					},
				},
				mode:    fs.ModeDir | fs.ModePerm,
				modtime: now,
			},
			From: "/a/b",
			To:   "/c",
		},
		{ // 9
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Output: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "c",
									directoryEntry: &inode{
										data:    []byte("/a"),
										modtime: now,
										mode:    fs.ModeSymlink | fs.ModePerm,
									},
								},
							},
							mode:    fs.ModeDir | fs.ModePerm,
							modtime: now,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			From: "/a",
			To:   "/b/c",
		},
	} {
		if err := test.FS.Symlink(test.From, test.To); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get FS %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func TestSymlinkResolveFile(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output []byte
		Err    error
	}{
		{ // 1
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
							mode: fs.ModePerm,
						},
					},
					{
						name: "b",
						directoryEntry: &inode{
							data: []byte("/a"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/b",
			Output: []byte("Hello"),
		},
		{ // 2
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
							mode: fs.ModePerm,
						},
					},
					{
						name: "b",
						directoryEntry: &inode{
							data: []byte("/c"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/b",
			Err:  fs.ErrNotExist,
		},
		{ // 3
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &inode{
							data: []byte("Hello"),
						},
					},
					{
						name: "b",
						directoryEntry: &inode{
							data: []byte("/a"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/b",
			Err:  fs.ErrPermission,
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("World"),
										mode: fs.ModePerm,
									},
								},
							},
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("/a/b"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/c",
			Output: []byte("World"),
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("World"),
										mode: fs.ModePerm,
									},
								},
							},
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("a/b"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/c",
			Output: []byte("World"),
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("/c"),
										mode: fs.ModeSymlink | fs.ModePerm,
									},
								},
							},
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("FooBar"),
							mode: fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/a/b",
			Output: []byte("FooBar"),
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("../c"),
										mode: fs.ModeSymlink | fs.ModePerm,
									},
								},
							},
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("FooBar"),
							mode: fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/a/b",
			Output: []byte("FooBar"),
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("../c"),
										mode: fs.ModeSymlink | fs.ModePerm,
									},
								},
							},
						},
					},
					{
						name: "c",
						directoryEntry: &inode{
							data: []byte("d"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
					{
						name: "d",
						directoryEntry: &inode{
							data: []byte("Baz"),
							mode: fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path:   "/a/b",
			Output: []byte("Baz"),
		},
	} {
		if output, err := test.FS.ReadFile(test.Path); !errors.Is(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !bytes.Equal(test.Output, output) {
			t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, output)
		}
	}
}

func TestSymlinkResolveDir(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output []fs.DirEntry
		Err    error
	}{
		{ // 1
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Foo"),
										mode: fs.ModePerm,
									},
								},
								{
									name: "c",
									directoryEntry: &inode{
										data: []byte("Bar"),
										mode: fs.ModePerm,
									},
								},
								{
									name: "d",
									directoryEntry: &inode{
										data: []byte("Baz"),
										mode: fs.ModePerm,
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
					{
						name: "e",
						directoryEntry: &inode{
							data: []byte("/f"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/e",
			Err:  fs.ErrNotExist,
		},
		{ // 1
			FS: FS{
				entries: []*dirEnt{
					{
						name: "a",
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									name: "b",
									directoryEntry: &inode{
										data: []byte("Foo"),
										mode: fs.ModePerm,
									},
								},
								{
									name: "c",
									directoryEntry: &inode{
										data: []byte("Bar"),
										mode: fs.ModePerm,
									},
								},
								{
									name: "d",
									directoryEntry: &inode{
										data: []byte("Baz"),
										mode: fs.ModePerm,
									},
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
					},
					{
						name: "e",
						directoryEntry: &inode{
							data: []byte("/a"),
							mode: fs.ModeSymlink | fs.ModePerm,
						},
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/e",
			Output: []fs.DirEntry{
				&dirEnt{
					name: "b",
					directoryEntry: &inode{
						data: []byte("Foo"),
						mode: fs.ModePerm,
					},
				},
				&dirEnt{
					name: "c",
					directoryEntry: &inode{
						data: []byte("Bar"),
						mode: fs.ModePerm,
					},
				},
				&dirEnt{
					name: "d",
					directoryEntry: &inode{
						data: []byte("Baz"),
						mode: fs.ModePerm,
					},
				},
			},
		},
	} {
		de, err := test.FS.ReadDir(test.Path)
		if !errors.Is(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !reflect.DeepEqual(test.Output, de) {
			t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, de)
		}
	}
}

func TestRemove(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output FS
		Err    error
	}{
		{ // 1
			Err: &fs.PathError{
				Op:   "remove",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "remove",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 3
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: FS{
				entries: []*dirEnt{},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: FS{
				entries: []*dirEnt{},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{},
									name:           "file",
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{},
									name:           "file",
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "remove",
				Path: "/dir",
				Err:  fs.ErrInvalid,
			},
		},
	} {
		if err := test.FS.Remove(test.Path); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get FS %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func TestRemoveAll(t *testing.T) {
	now := time.Now()
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output FS
		Err    error
	}{
		{ // 1
			Err: &fs.PathError{
				Op:   "removeall",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: &fs.PathError{
				Op:   "removeall",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 3
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: FS{
				entries: []*dirEnt{},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: FS{
				entries: []*dirEnt{},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{},
									name:           "file",
								},
							},
							mode: fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: FS{
				entries: []*dirEnt{},
				modtime: now,
				mode:    fs.ModeDir | fs.ModePerm,
			},
		},
	} {
		if err := test.FS.RemoveAll(test.Path); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get FS %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func TestLStat(t *testing.T) {
	for n, test := range [...]struct {
		FS     FS
		Path   string
		Output fs.FileInfo
		Err    error
	}{
		{ // 1
			FS: FS{},
			Err: &fs.PathError{
				Op:   "lstat",
				Path: "",
				Err:  fs.ErrPermission,
			},
		},
		{ // 2
			FS: FS{
				modtime: time.Unix(1, 2),
				mode:    fs.ModeDir | fs.ModePerm,
			},
			Output: &dirEnt{
				directoryEntry: &dnode{
					modtime: time.Unix(1, 2),
					mode:    fs.ModeDir | fs.ModePerm,
				},
				name: "/",
			},
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "lstat",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 4
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{},
						name:           "notFile",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err: &fs.PathError{
				Op:   "lstat",
				Path: "/file",
				Err:  fs.ErrNotExist,
			},
		},
		{ // 5
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    fs.ModeSymlink | 3,
						},
						name: "file",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Output: &dirEnt{
				directoryEntry: &inode{
					modtime: time.Unix(1, 2),
					mode:    fs.ModeSymlink | 3,
				},
				name: "file",
			},
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							modtime: time.Unix(4, 5),
							mode:    6,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir",
			Output: &dirEnt{
				directoryEntry: &dnode{
					modtime: time.Unix(4, 5),
					mode:    6,
				},
				name: "dir",
			},
		},
		{ // 7
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(4, 5),
										mode:    6,
									},
									name: "anotherFile",
								},
							},
							modtime: time.Unix(7, 8),
							mode:    9,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir/anotherFile",
			Err: &fs.PathError{
				Op:   "lstat",
				Path: "/dir/anotherFile",
				Err:  fs.ErrPermission,
			},
		},
		{ // 8
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &inode{
							modtime: time.Unix(1, 2),
							mode:    3,
						},
						name: "file",
					},
					{
						directoryEntry: &dnode{
							entries: []*dirEnt{
								{
									directoryEntry: &inode{
										modtime: time.Unix(4, 5),
										mode:    fs.ModeSymlink | 6,
									},
									name: "anotherFile",
								},
							},
							modtime: time.Unix(7, 8),
							mode:    fs.ModeDir | fs.ModePerm,
						},
						name: "dir",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/dir/anotherFile",
			Output: &dirEnt{
				directoryEntry: &inode{
					modtime: time.Unix(4, 5),
					mode:    fs.ModeSymlink | 6,
				},
				name: "anotherFile",
			},
		},
	} {
		if f, err := test.FS.LStat(test.Path); !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !reflect.DeepEqual(f, test.Output) {
			t.Errorf("test %d: expected FileInfo %v, got %v", n+1, test.Output, f)
		}
	}
}

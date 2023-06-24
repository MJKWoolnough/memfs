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
			Err:  fs.ErrPermission,
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err:  fs.ErrNotExist,
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
			Err:  fs.ErrPermission,
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
			Err:  fs.ErrNotExist,
		},
		{ // 6
			FS: FS{
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							name: "dir",
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
		if !errors.Is(err, test.Err) {
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
			FS:  FS{},
			Err: fs.ErrPermission,
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
							name: "test2",
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
						name: "test2",
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
							name: "test2",
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
							name: "test2",
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
			Err:  fs.ErrPermission,
		},
	} {
		de, err := test.FS.ReadDir(test.Path)
		if !errors.Is(test.Err, err) {
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
			FS:  FS{},
			Err: fs.ErrPermission,
		},
		{ // 2
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Err: fs.ErrInvalid,
		},
		{ // 3
			FS: FS{
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/file",
			Err:  fs.ErrNotExist,
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
			Err:  fs.ErrNotExist,
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
			Err:  fs.ErrPermission,
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
		if !errors.Is(test.Err, err) {
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
			FS:  FS{},
			Err: fs.ErrPermission,
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
			Err:  fs.ErrNotExist,
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
			Err:  fs.ErrNotExist,
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
							name:    "dir",
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
					name:    "dir",
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
							name: "dir",
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
			Err:  fs.ErrPermission,
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
							name: "dir",
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
		if !errors.Is(test.Err, err) {
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
			FS: FS{},
			Output: FS{
				modtime: now,
			},
			Err: fs.ErrInvalid,
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
			Err: fs.ErrInvalid,
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
			Err:  fs.ErrInvalid,
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
							name:    "a",
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
			Err:  fs.ErrNotExist,
		},
		{ // 6
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							name:    "a",
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
							name:    "a",
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err:  fs.ErrPermission,
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
			Err:  fs.ErrInvalid,
		},
		{ // 8
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							name:    "a",
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
							name: "a",
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										name:    "b",
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
		if err := test.FS.Mkdir(test.Path, test.PathPerms); !errors.Is(test.Err, err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

func fixTimes(d *dnode, now time.Time) {
	d.modtime = now
	for _, e := range d.entries {
		if de, ok := e.directoryEntry.(*dnode); ok {
			fixTimes(de, now)
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
			FS: FS{},
			Output: FS{
				modtime: now,
			},
			Err: fs.ErrInvalid,
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
			Err: fs.ErrInvalid,
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
			Err:  fs.ErrInvalid,
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
							name:    "a",
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
							name:    "a",
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err:  fs.ErrPermission,
		},
		{ // 6
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							name:    "a",
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
							name:    "a",
							modtime: now,
							mode:    fs.ModeDir,
						},
						name: "a",
					},
				},
				mode: fs.ModeDir | fs.ModePerm,
			},
			Path: "/a/b",
			Err:  fs.ErrPermission,
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
			Err:  fs.ErrInvalid,
		},
		{ // 8
			FS: FS{
				modtime: now,
				entries: []*dirEnt{
					{
						directoryEntry: &dnode{
							name:    "a",
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
							name: "a",
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										name:    "b",
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
							name: "a",
							entries: []*dirEnt{
								{
									directoryEntry: &dnode{
										name:    "b",
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
		if err := test.FS.MkdirAll(test.Path, test.PathPerms); !errors.Is(test.Err, err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else {
			fixTimes((*dnode)(&test.FS), now)
			if !reflect.DeepEqual(test.Output, test.FS) {
				t.Errorf("test %d: expecting to get %v, got %v", n+1, test.Output, test.FS)
			}
		}
	}
}

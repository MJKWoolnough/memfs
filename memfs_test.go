package memfs

import (
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
		{
			FS: FS{
				dnode: &dnode{},
				root:  "/",
			},
			Path: "/file",
			Err:  fs.ErrPermission,
		},
		{
			FS: FS{
				dnode: &dnode{
					mode: fs.ModeDir | fs.ModePerm,
				},
				root: "/",
			},
			Path: "/file",
			Err:  fs.ErrNotExist,
		},
		{
			FS: FS{
				dnode: &dnode{
					entries: []*dirEnt{
						{
							directoryEntry: &inode{},
							name:           "file",
						},
					},
				},
				root: "/",
			},
			Path: "/file",
			Err:  fs.ErrPermission,
		},
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
			},
			Path: "/file",
			Err:  fs.ErrNotExist,
		},
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{},
				root:  "/",
			},
			Err: fs.ErrPermission,
		},
		{
			FS: FS{
				dnode: &dnode{
					mode: fs.ModeDir | fs.ModePerm,
				},
				root: "/",
			},
			Output: []fs.DirEntry{},
		},
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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
		{
			FS: FS{
				dnode: &dnode{
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
				root: "/",
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

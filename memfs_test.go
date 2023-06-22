package memfs

import (
	"errors"
	"io/fs"
	"reflect"
	"testing"
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

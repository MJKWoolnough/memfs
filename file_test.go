package memfs

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
)

var _ interface {
	io.ReadSeekCloser
	io.ReaderAt
	io.Writer
	io.WriterTo
	io.WriterAt
	io.RuneScanner
	io.ByteScanner
	io.ByteWriter
} = &file{}

func TestRead(t *testing.T) {
	for n, test := range [...]struct {
		Data []byte
		Mode opMode
		Err  error
	}{
		{
			Err: fs.ErrClosed,
		},
		{
			Mode: opRead,
		},
		{
			Data: []byte("Hello, World"),
			Mode: opRead,
		},
		{
			Data: []byte(strings.Repeat("Hello, World!", 1000)),
			Mode: opRead,
		},
		{
			Mode: opSeek,
			Err:  fs.ErrInvalid,
		},
	} {
		f := file{
			inode: &inode{
				data: test.Data,
			},
			opMode: test.Mode,
		}

		data, err := io.ReadAll(&f)
		if !errors.Is(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if bytes.Compare(data, test.Data) != 0 {
			t.Errorf("test %d: expecting bytes %v, got %v", n+1, test.Data, data)
		}
	}
}

func TestReadAt(t *testing.T) {
	for n, test := range [...]struct {
		Data   []byte
		Mode   opMode
		Read   [][2]int64
		Output [][]byte
		Err    error
	}{
		{
			Err: fs.ErrClosed,
		},
		{
			Data: []byte("Hello, World"),
			Mode: opRead,
			Read: [][2]int64{
				{1, 0},
			},
			Output: [][]byte{
				[]byte("H"),
			},
			Err: fs.ErrInvalid,
		},
		{
			Mode: opSeek,
			Data: []byte("Hello, World"),
			Read: [][2]int64{
				{1, 0},
			},
			Output: [][]byte{
				[]byte("H"),
			},
			Err: fs.ErrInvalid,
		},
		{
			Mode: opRead | opSeek,
			Data: []byte("Hello, World"),
			Read: [][2]int64{
				{1, 0},
			},
			Output: [][]byte{
				[]byte("H"),
			},
		},
		{
			Mode: opRead | opSeek,
			Data: []byte("Hello, World"),
			Read: [][2]int64{
				{1, 0},
				{1, 0},
				{2, 0},
				{2, 4},
				{4, 2},
			},
			Output: [][]byte{
				[]byte("H"),
				[]byte("H"),
				[]byte("He"),
				[]byte("o,"),
				[]byte("llo,"),
			},
		},
	} {
		f := file{
			inode: &inode{
				data: test.Data,
			},
			opMode: test.Mode,
		}

		readAtTests := func(o int) bool {
			for m, toRead := range test.Read {
				buf := make([]byte, toRead[0])
				l, err := f.ReadAt(buf, toRead[1])
				if !errors.Is(err, test.Err) {
					t.Errorf("test %d.%d.%d: expecting error %s, got %s", n+1, o, m+1, test.Err, err)

					return false
				} else if test.Err != nil {
					return false
				} else if len(buf) != l {
					t.Errorf("test %d.%d.%d: expecting to read %d bytes, read %d", n+1, o, m+1, len(buf), l)

					return false
				} else if string(buf) != string(test.Output[m]) {
					t.Errorf("test %d.%d.%d: expecting to read %s, read %s", n+1, o, m+1, test.Output[m], buf)

					return false
				}
			}

			return true
		}

		if !readAtTests(1) {
			continue
		}

		if !readAtTests(2) {
			continue
		}

		if test.Err != nil {
			continue
		}

		_, err := f.Read([]byte{1})
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)

			continue
		}

		if !readAtTests(3) {
			continue
		}

		_, err = f.Read([]byte{2})
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)

			continue
		}

		if !readAtTests(4) {
			continue
		}

		_, err = io.ReadAll(&f)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)

			continue
		}

		if !readAtTests(5) {
			continue
		}
	}
}

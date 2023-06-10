package memfs

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
)

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

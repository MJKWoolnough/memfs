package memfs

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"sync"
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
} = &File{}

func TestReadRW(t *testing.T) {
	for n, test := range [...]struct {
		Data []byte
		Mode opMode
		Err  error
	}{
		{
			Err: &fs.PathError{
				Op:  "read",
				Err: fs.ErrClosed,
			},
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
			Err: &fs.PathError{
				Op:  "read",
				Err: fs.ErrInvalid,
			},
		},
	} {
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: test.Data,
				},
				opMode: test.Mode,
			},
		}

		data, err := io.ReadAll(&f)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if !bytes.Equal(data, test.Data) {
			t.Errorf("test %d: expecting bytes %v, got %v", n+1, test.Data, data)
		}
	}
}

func TestReadAtRW(t *testing.T) {
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
			Err: &fs.PathError{
				Op:   "readat",
				Path: "",
				Err:  fs.ErrInvalid,
			},
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
			Err: &fs.PathError{
				Op:   "readat",
				Path: "",
				Err:  fs.ErrInvalid,
			},
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
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: test.Data,
				},
				opMode: test.Mode,
			},
		}

		readAtTests := func(o int) bool {
			for m, toRead := range test.Read {
				buf := make([]byte, toRead[0])
				l, err := f.ReadAt(buf, toRead[1])
				if !reflect.DeepEqual(err, test.Err) {
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

func TestReadByteRW(t *testing.T) {
Tests:
	for n, test := range [...]struct {
		Data []byte
		Mode opMode
		Err  error
	}{
		{
			Err: &fs.PathError{
				Op:   "readbyte",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{
			Data: []byte{'a'},
			Err: &fs.PathError{
				Op:   "readbyte",
				Path: "",
				Err:  fs.ErrClosed,
			},
		},
		{
			Data: []byte("abc"),
			Mode: opRead,
		},
	} {
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: test.Data,
				},
				opMode: test.Mode,
			},
		}

		for i := range test.Data {
			b, err := f.ReadByte()
			if !reflect.DeepEqual(err, test.Err) {
				t.Errorf("test %d.%d: expecting error %s, got %s", n+1, i+1, test.Err, err)
			} else if test.Err != nil {
				continue Tests
			} else if test.Data[i] != b {
				t.Errorf("test %d.%d: expecting to read byte %d, got %d", n+1, i+1, test.Data[0], b)
			}
		}

		if test.Err != nil {
			continue
		}

		b, err := f.ReadByte()
		if !errors.Is(err, io.EOF) {
			t.Errorf("test %d.%d: expecting error %s, got %s", n+1, len(test.Data)+1, io.EOF, err)
		} else if b != 0 {
			t.Errorf("test %d.%d: expecting to read byte %d, got %d", n+1, len(test.Data)+1, 0, b)
		}
	}
}

func TestUnreadByteRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: []byte("12345"),
			},
			opMode: opRead,
		},
	}

	f.ReadByte()

	err := f.UnreadByte()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadbyte",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)

		return
	}

	f.opMode |= opSeek

	c, _ := f.ReadByte()
	if c != '2' {
		t.Errorf("test 2: expecting to read '2', read %q", c)

		return
	}

	err = f.UnreadByte()
	if !errors.Is(err, nil) {
		t.Errorf("test 3: expecting nil error, got %s", err)

		return
	}

	c, _ = f.ReadByte()
	if c != '2' {
		t.Errorf("test 4: expecting to read '2', read %q", c)

		return
	}

	err = f.UnreadByte()
	if !errors.Is(err, nil) {
		t.Errorf("test 5: expecting nil error, got %s", err)

		return
	}

	err = f.UnreadByte()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadbyte",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 6: expecting ErrInvalid error, got %s", err)

		return
	}

	c, _ = f.ReadByte()
	if c != '2' {
		t.Errorf("test 7: expecting to read '2', read %q", c)

		return
	}

	f.ReadByte()

	err = f.UnreadByte()
	if !errors.Is(err, nil) {
		t.Errorf("test 8: expecting nil error, got %s", err)

		return
	}

	c, _ = f.ReadByte()
	if c != '3' {
		t.Errorf("test 9: expecting to read '3', read %q", c)

		return
	}

	f.Read([]byte{0})

	err = f.UnreadByte()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadbyte",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 10: expecting nil error, got %s", err)

		return
	}
}

func TestReadRuneRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: []byte("1ħᕗ🐶"),
			},
		},
	}

	_, _, err := f.ReadRune()
	if !errors.Is(err, fs.ErrClosed) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)

		return
	}

	f.opMode = opRead

	r, n, err := f.ReadRune()

	if !errors.Is(err, nil) {
		t.Errorf("test 2: expecting nil error, got %s", err)

		return
	} else if n != 1 {
		t.Errorf("test 2: expecting to read 1 byte, read %d", n)

		return
	} else if r != '1' {
		t.Errorf("test 2: expecting to read '1', read %q", r)

		return
	}

	r, n, err = f.ReadRune()

	if !errors.Is(err, nil) {
		t.Errorf("test 3: expecting nil error, got %s", err)

		return
	} else if n != 2 {
		t.Errorf("test 3: expecting to read 2 bytes, read %d", n)

		return
	} else if r != 'ħ' {
		t.Errorf("test 3: expecting to read 'ħ', read %q", r)

		return
	}

	r, n, err = f.ReadRune()

	if !errors.Is(err, nil) {
		t.Errorf("test 4: expecting nil error, got %s", err)

		return
	} else if n != 3 {
		t.Errorf("test 4: expecting to read 3 bytes, read %d", n)

		return
	} else if r != 'ᕗ' {
		t.Errorf("test 4: expecting to read 'ᕗ', read %q", r)

		return
	}

	r, n, err = f.ReadRune()

	if !errors.Is(err, nil) {
		t.Errorf("test 5: expecting nil error, got %s", err)

		return
	} else if n != 4 {
		t.Errorf("test 5: expecting to read 4 bytes, read %d", n)

		return
	} else if r != '🐶' {
		t.Errorf("test 5: expecting to read '🐶', read %q", r)

		return
	}

	r, n, err = f.ReadRune()
	if !errors.Is(err, io.EOF) {
		t.Errorf("test 6: expecting error EOF, got %s", err)

		return
	} else if n != 0 {
		t.Errorf("test 6: expecting to read 0 bytes, read %d", n)

		return
	} else if r != 0 {
		t.Errorf("test 6: expecting to read 0, read %q", r)

		return
	}
}

func TestUnreadRuneRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: []byte("1ħᕗ🐶"),
			},
			opMode: opRead,
		},
	}

	f.ReadRune()

	err := f.UnreadRune()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadrune",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)

		return
	}

	f.opMode |= opSeek
	f.pos = 0

	c, _, _ := f.ReadRune()
	if c != '1' {
		t.Errorf("test 2: expecting to read '1', read %q", c)

		return
	}

	err = f.UnreadRune()
	if !errors.Is(err, nil) {
		t.Errorf("test 3: expecting nil error, got %s", err)

		return
	}

	c, _, _ = f.ReadRune()
	if c != '1' {
		t.Errorf("test 4: expecting to read '1', read %q", c)

		return
	}

	err = f.UnreadRune()
	if !errors.Is(err, nil) {
		t.Errorf("test 5: expecting nil error, got %s", err)

		return
	}

	err = f.UnreadRune()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadrune",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 6: expecting ErrInvalid error, got %s", err)

		return
	}

	c, _, _ = f.ReadRune()
	if c != '1' {
		t.Errorf("test 7: expecting to read '1', read %q", c)

		return
	}

	err = f.UnreadRune()
	if !errors.Is(err, nil) {
		t.Errorf("test 8: expecting nil error, got %s", err)

		return
	}

	c, _, _ = f.ReadRune()
	if c != '1' {
		t.Errorf("test 9: expecting to read '1', read %q", c)

		return
	}

	f.Read([]byte{0})

	err = f.UnreadRune()
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "unreadrune",
		Path: "",
		Err:  fs.ErrInvalid,
	}) {
		t.Errorf("test 10: expecting nil error, got %s", err)

		return
	}

	f.pos = 1

	f.ReadRune()

	f.UnreadRune()

	c, _, _ = f.ReadRune()
	if c != 'ħ' {
		t.Errorf("test 11: expecting to read 'ħ', read %q", c)

		return
	}

	f.ReadRune()
	f.ReadRune()
	f.UnreadRune()

	c, _, _ = f.ReadRune()
	if c != '🐶' {
		t.Errorf("test 12: expecting to read '🐶', read %q", c)

		return
	}

	f.UnreadRune()

	c, _, _ = f.ReadRune()
	if c != '🐶' {
		t.Errorf("test 12: expecting to read '🐶', read %q", c)

		return
	}
}

func TestWriteToRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: []byte("12345"),
			},
		},
	}

	var sb strings.Builder

	_, err := f.WriteTo(&sb)
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "writeto",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 1: expecting to get error ErrClosed, got %s", err)
	}

	f.opMode = opRead

	n, err := f.WriteTo(&sb)
	if err != nil {
		t.Errorf("test 2: expecting to get no error, got %s", err)
	} else if n != 5 {
		t.Errorf("test 2: expecting to read 5 bytes, read %d", n)
	} else if str := sb.String(); str != "12345" {
		t.Errorf("test 2: expecting to write %q, wrote %q", "12345", str)
	}

	_, err = f.WriteTo(&sb)
	if !errors.Is(err, io.EOF) {
		t.Errorf("test 3: expecting to get error EOF, got %s", err)
	}

	f.pos = 1
	sb.Reset()

	n, err = f.WriteTo(&sb)
	if err != nil {
		t.Errorf("test 4: expecting to get no error, got %s", err)
	} else if n != 4 {
		t.Errorf("test 4: expecting to read 5 bytes, read %d", n)
	} else if str := sb.String(); str != "2345" {
		t.Errorf("test 4: expecting to write %q, wrote %q", "2345", str)
	}
}

func TestSeekRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: make([]byte, 100),
			},
			opMode: opSeek,
		},
	}
	for n, test := range [...]struct {
		Offset int64
		Whence int
		Pos    int64
		Err    error
	}{
		{
			Offset: 0,
			Whence: io.SeekStart,
			Pos:    0,
		},
		{
			Offset: -1,
			Whence: io.SeekStart,
			Pos:    0,
			Err: &fs.PathError{
				Op:   "seek",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{
			Offset: 10,
			Whence: io.SeekStart,
			Pos:    10,
		},
		{
			Offset: 10,
			Whence: io.SeekStart,
			Pos:    10,
		},
		{
			Offset: -1,
			Whence: io.SeekStart,
			Pos:    0,
			Err: &fs.PathError{
				Op:   "seek",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{
			Offset: 10,
			Whence: io.SeekStart,
			Pos:    10,
		},
		{
			Offset: 10,
			Whence: io.SeekCurrent,
			Pos:    20,
		},
		{
			Offset: 10,
			Whence: io.SeekCurrent,
			Pos:    30,
		},
		{
			Offset: -10,
			Whence: io.SeekCurrent,
			Pos:    20,
		},
		{
			Offset: 0,
			Whence: io.SeekCurrent,
			Pos:    20,
		},
		{
			Offset: 0,
			Whence: io.SeekEnd,
			Pos:    100,
		},
		{
			Offset: 10,
			Whence: io.SeekEnd,
			Pos:    110,
		},
		{
			Offset: -10,
			Whence: io.SeekEnd,
			Pos:    90,
		},
	} {
		pos, err := f.Seek(test.Offset, test.Whence)
		if !reflect.DeepEqual(err, test.Err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if pos != test.Pos {
			t.Errorf("test %d: expecting pos %d, got %d", n+1, test.Pos, pos)
		}
	}
}

func TestWrite(t *testing.T) {
	var toWrite [256]byte

	for n := range toWrite {
		toWrite[n] = byte(n)
	}

	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: make([]byte, 100),
			},
		},
	}

	n, err := f.Write(toWrite[:10])
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "write",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)
	} else if n != 0 {
		t.Errorf("test 1: expecting to write 0 bytes, wrote %d", n)
	}

	for n := range toWrite {
		if n == 0 {
			continue
		}
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: make([]byte, 100),
				},
				opMode: opWrite,
			},
		}
		for i := 0; i < 100; i++ {
			m, err := f.Write(toWrite[:n])
			if !errors.Is(err, nil) {
				t.Errorf("test %d: expecting no error, got %s", n+1, err)
			} else if m != n {
				t.Errorf("test %d: expecting to write %d bytes, wrote %d", n+1, n, m)
			}
		}
		expected := bytes.Repeat(toWrite[:n], 100)
		if !bytes.Equal(f.data, expected) {
			t.Errorf("test %d: expecting to write %v, wrote %v", n+1, expected, f.data)
		}
	}
}

func TestWriteAt(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: make([]byte, 10),
			},
		},
	}

	n, err := f.WriteAt([]byte{0}, 0)
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "writeat",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)
	} else if n != 0 {
		t.Errorf("test 1: expecting to write 0 bytes, wrote %d", n)
	}

	f.opMode = opWrite | opSeek

	for n, test := range [...]struct {
		ToWrite []byte
		Pos     int64
		N       int
		Err     error
		Buffer  []byte
	}{
		{
			ToWrite: []byte("Beep"),
			Pos:     2,
			N:       4,
			Err:     nil,
			Buffer:  []byte("\000\000Beep\000\000\000\000"),
		},
		{
			ToWrite: []byte("Boop"),
			Pos:     2,
			N:       4,
			Err:     nil,
			Buffer:  []byte("\000\000Boop\000\000\000\000"),
		},
		{
			ToWrite: []byte("FooBar"),
			Pos:     12,
			N:       6,
			Err:     nil,
			Buffer:  []byte("\000\000Boop\000\000\000\000\000\000FooBar"),
		},
		{
			ToWrite: []byte("Hello"),
			Pos:     0,
			N:       5,
			Err:     nil,
			Buffer:  []byte("Hellop\000\000\000\000\000\000FooBar"),
		},
	} {
		m, err := f.WriteAt(test.ToWrite, test.Pos)
		if !reflect.DeepEqual(test.Err, err) {
			t.Errorf("test %d: expecting error %s, got %s", n+1, test.Err, err)
		} else if m != test.N {
			t.Errorf("test %d: expecting to write %d bytes, wrote %d", n+1, test.N, m)
		} else if !bytes.Equal(f.data, test.Buffer) {
			t.Errorf("test %d: expecting buffer to be %v bytes, got %v", n+1, test.Buffer, f.data)
		}
	}
}

func TestWriteString(t *testing.T) {
	var toWrite [256]byte

	for n := range toWrite {
		toWrite[n] = byte(n)
	}

	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode: &inode{
				data: make([]byte, 100),
			},
		},
	}

	n, err := f.WriteString(string(toWrite[:10]))
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "writestring",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)
	} else if n != 0 {
		t.Errorf("test 1: expecting to write 0 bytes, wrote %d", n)
	}

	for n := range toWrite {
		if n == 0 {
			continue
		}
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: make([]byte, 100),
				},
				opMode: opWrite,
			},
		}
		for i := 0; i < 100; i++ {
			m, err := f.WriteString(string(toWrite[:n]))
			if !errors.Is(err, nil) {
				t.Errorf("test %d: expecting no error, got %s", n+1, err)
			} else if m != n {
				t.Errorf("test %d: expecting to write %d bytes, wrote %d", n+1, n, m)
			}
		}
		expected := bytes.Repeat(toWrite[:n], 100)
		if !bytes.Equal(f.data, expected) {
			t.Errorf("test %d: expecting to write %v, wrote %v", n+1, expected, f.data)
		}
	}
}

func TestWriteByte(t *testing.T) {
Tests:
	for n, test := range [...]struct {
		Data []byte
		Mode opMode
		Err  error
	}{
		{
			Err: &fs.PathError{
				Op:   "writebyte",
				Path: "",
				Err:  fs.ErrInvalid,
			},
		},
		{
			Data: []byte{'a'},
			Err: &fs.PathError{
				Op:   "writebyte",
				Path: "",
				Err:  fs.ErrClosed,
			},
		},
		{
			Data: []byte("abc"),
			Mode: opWrite,
		},
	} {
		f := File{
			mu: &sync.RWMutex{},
			file: file{
				inode: &inode{
					data: test.Data,
				},
				opMode: test.Mode,
			},
		}

		for i := range test.Data {
			err := f.WriteByte(test.Data[i])
			if !reflect.DeepEqual(err, test.Err) {
				t.Errorf("test %d.%d: expecting error %s, got %s", n+1, i+1, test.Err, err)
			} else if test.Err != nil {
				continue Tests
			}
		}

		if test.Err != nil {
			continue
		}

		if !bytes.Equal(test.Data, f.data) {
			t.Errorf("test %d: expecting to write %s, wrote %s", n+1, test.Data, f.data)
		}
	}
}

func TestCloseRW(t *testing.T) {
	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode:  &inode{},
			opMode: opWrite,
		},
	}

	_, err := f.WriteString("123")
	if err != nil {
		t.Errorf("test 1: expecting nil error, got %s", err)
	}

	err = f.Close()
	if err != nil {
		t.Errorf("test 2: expecting nil error, got %s", err)
	}

	_, err = f.WriteString("1")
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "writestring",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 3: expecting ErrClosed error, got %s", err)
	}

	f.opMode = opRead

	var buf [1]byte

	_, err = f.Read(buf[:])
	if err != nil {
		t.Errorf("test 4: expecting nil error, got %s", err)
	} else if buf[0] != '1' {
		t.Errorf("test 4: expecting to read '1', read %s", buf[:1])
	}

	err = f.Close()
	if err != nil {
		t.Errorf("test 5: expecting nil error, got %s", err)
	}

	_, err = f.Read(buf[:])
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "read",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 6: expecting ErrClosed error, got %s", err)
	}

	f.opMode = opSeek

	pos, err := f.Seek(1, io.SeekStart)
	if err != nil {
		t.Errorf("test 7: expecting nil error, got %s", err)
	} else if pos != 1 {
		t.Errorf("test 7: expecting to be at position 1, at %d", pos)
	}

	err = f.Close()
	if err != nil {
		t.Errorf("test 8: expecting nil error, got %s", err)
	}

	pos, err = f.Seek(1, io.SeekStart)
	if !reflect.DeepEqual(err, &fs.PathError{
		Op:   "seek",
		Path: "",
		Err:  fs.ErrClosed,
	}) {
		t.Errorf("test 9: expecting ErrClosed error, got %s", err)
	}
}

func TestReadFrom(t *testing.T) {
	buf := make([]byte, 0, 255*1024*1024)

	for i := 0; i < 1024*1024; i++ {
		for j := 0; j < 255; j++ {
			buf = append(buf, byte(j))
		}
	}

	f := File{
		mu: &sync.RWMutex{},
		file: file{
			inode:  &inode{},
			opMode: opWrite,
		},
	}

	n, err := f.ReadFrom(bytes.NewBuffer(buf))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if n != 255*1024*1024 {
		t.Errorf("expected to read %d bytes, read %d", 255*1024*1024, n)
	} else if !bytes.Equal(buf, f.data) {
		t.Errorf("File data does not match buf data")
	}
}

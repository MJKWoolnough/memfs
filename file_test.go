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

func TestReadByte(t *testing.T) {
Tests:
	for n, test := range [...]struct {
		Data []byte
		Mode opMode
		Err  error
	}{
		{
			Err: fs.ErrInvalid,
		},
		{
			Data: []byte{'a'},
			Err:  fs.ErrClosed,
		},
		{
			Data: []byte("abc"),
			Mode: opRead,
		},
	} {
		f := file{
			inode: &inode{
				data: test.Data,
			},
			opMode: test.Mode,
		}

		for i := range test.Data {
			b, err := f.ReadByte()
			if !errors.Is(test.Err, err) {
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
		if !errors.Is(io.EOF, err) {
			t.Errorf("test %d.%d: expecting error %s, got %s", n+1, len(test.Data)+1, io.EOF, err)
		} else if b != 0 {
			t.Errorf("test %d.%d: expecting to read byte %d, got %d", n+1, len(test.Data)+1, 0, b)
		}
	}
}

func TestUnreadByte(t *testing.T) {
	f := file{
		inode: &inode{
			data: []byte("12345"),
		},
		opMode: opRead,
	}

	f.ReadByte()

	err := f.UnreadByte()
	if !errors.Is(err, fs.ErrInvalid) {
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
	if !errors.Is(err, fs.ErrInvalid) {
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
	if !errors.Is(err, fs.ErrInvalid) {
		t.Errorf("test 10: expecting nil error, got %s", err)

		return
	}
}

func TestReadRune(t *testing.T) {
	f := file{
		inode: &inode{
			data: []byte("1ħᕗ🐶"),
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

func TestUnreadRune(t *testing.T) {
	f := file{
		inode: &inode{
			data: []byte("1ħᕗ🐶"),
		},
		opMode: opRead,
	}

	f.ReadRune()

	err := f.UnreadRune()
	if !errors.Is(err, fs.ErrInvalid) {
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
	if !errors.Is(err, fs.ErrInvalid) {
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
	if !errors.Is(err, fs.ErrInvalid) {
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

func TestSeek(t *testing.T) {
	f := file{
		inode: &inode{
			data: make([]byte, 100),
		},
		opMode: opSeek,
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
			Err:    fs.ErrInvalid,
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
			Err:    fs.ErrInvalid,
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
		if !errors.Is(err, test.Err) {
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

	f := file{
		inode: &inode{
			data: make([]byte, 100),
		},
	}

	n, err := f.Write(toWrite[:10])
	if !errors.Is(err, fs.ErrClosed) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)
	} else if n != 0 {
		t.Errorf("test 1: expecting to write 0 bytes, wrote %d", n)
	}

	for n := range toWrite {
		if n == 0 {
			continue
		}
		f := file{
			inode: &inode{
				data: make([]byte, 100),
			},
			opMode: opWrite,
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
	f := file{
		inode: &inode{
			data: make([]byte, 10),
		},
	}

	n, err := f.WriteAt([]byte{0}, 0)
	if !errors.Is(err, fs.ErrClosed) {
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
		if !errors.Is(test.Err, err) {
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

	f := file{
		inode: &inode{
			data: make([]byte, 100),
		},
	}

	n, err := f.WriteString(string(toWrite[:10]))
	if !errors.Is(err, fs.ErrClosed) {
		t.Errorf("test 1: expecting ErrClosed, got %s", err)
	} else if n != 0 {
		t.Errorf("test 1: expecting to write 0 bytes, wrote %d", n)
	}

	for n := range toWrite {
		if n == 0 {
			continue
		}
		f := file{
			inode: &inode{
				data: make([]byte, 100),
			},
			opMode: opWrite,
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

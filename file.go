package memfs

import (
	"io"
	"io/fs"
	"time"
	"unicode/utf8"
)

type opMode uint8

const (
	opRead opMode = 1 << iota
	opWrite
	opSeek
)

type inode struct {
	modtime time.Time
	data    []byte
	mode    fs.FileMode
}

type file struct {
	name string
	*inode
	opMode   opMode
	lastRead uint8
	pos      int64
}

func (f *file) validTo(m opMode) error {
	if f.opMode == 0 {
		return fs.ErrClosed
	}

	if f.opMode&m != m {
		return fs.ErrInvalid
	}

	return nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Read(p []byte) (int, error) {
	if err := f.validTo(opRead); err != nil {
		return 0, err
	}

	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n := copy(p, f.data[f.pos:])

	f.pos += int64(n)
	f.lastRead = 0

	return n, nil
}

func (f *file) ReadAt(p []byte, off int64) (int, error) {
	if err := f.validTo(opRead); err != nil {
		return 0, err
	}

	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n := copy(p, f.data[off:])

	if n < len(p) {
		return 0, io.EOF
	}

	return n, nil
}

func (f *file) ReadByte() (byte, error) {
	if err := f.validTo(opRead); err != nil {
		return 0, err
	}

	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}

	b := f.data[f.pos]

	f.pos++
	f.lastRead = 1

	return b, nil
}

func (f *file) UnreadByte() error {
	if err := f.validTo(opRead | opSeek); err != nil {
		return err
	}

	if f.lastRead != 1 {
		return fs.ErrInvalid
	}

	f.lastRead = 0

	f.pos--

	return nil
}

func (f *file) ReadRune() (rune, int, error) {
	if err := f.validTo(opRead); err != nil {
		return 0, 0, err
	}

	if f.pos >= int64(len(f.data)) {
		return 0, 0, io.EOF
	}

	r, s := utf8.DecodeRune(f.data[f.pos:])

	f.lastRead = uint8(s)

	return r, s, nil
}

func (f *file) UnreadRune() error {
	if err := f.validTo(opRead | opSeek); err != nil {
		return err
	}

	if f.lastRead == 0 {
		return fs.ErrInvalid
	}

	f.lastRead = 0

	f.pos -= int64(f.lastRead)

	return nil
}

func (f *file) WriteTo(w io.Writer) (int64, error) {
	if err := f.validTo(opRead); err != nil {
		return 0, err
	}

	n, err := w.Write(f.data[f.pos:])
	f.pos += int64(n)
	f.lastRead = 0

	return int64(n), err
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if err := f.validTo(opSeek); err != nil {
		return 0, err
	}

	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos += offset
	case io.SeekEnd:
		f.pos = int64(len(f.data)) + offset
	default:
		return 0, fs.ErrInvalid
	}

	f.lastRead = 0

	if f.pos < 0 {
		f.pos = 0

		return f.pos, fs.ErrInvalid
	}

	return f.pos, nil
}

func (f *file) grow(size int) {
	if size > len(f.data) {
		if size < cap(f.data) {
			f.data = (f.data)[:size]
		} else {
			var newData []byte
			if len(f.data) < 512 {
				newData = make([]byte, size, size<<1)
			} else {
				newData = make([]byte, size, size+(size>>2))
			}
			copy(newData, f.data)
			f.data = newData
		}
	}
}

func (f *file) Write(p []byte) (int, error) {
	return 0, nil
}

func (f *file) WriterAt(p []byte) (int, error) {
	return 0, nil
}

func (f *file) WriteString(str string) (int, error) {
	return 0, nil
}

func (f *file) WriteByte(c byte) error {
	return nil
}

func (f *file) Close() error {
	return nil
}

func (f *file) Name() string {
	return ""
}

func (f *file) Size() int64 {
	return 0
}

func (f *file) Mode() fs.FileMode {
	return 0
}

func (f *file) ModTime() time.Time {
	return time.Time{}
}

func (f *file) IsDir() bool {
	return false
}

func (f *file) Sys() any {
	return f
}

package memfs

import (
	"io"
	"io/fs"
	"time"
	"unicode/utf8"
)

type opMode uint8

const (
	opClose opMode = 0
	opRead  opMode = 1 << iota
	opWrite
	opSeek
)

type inode struct {
	modtime time.Time
	data    []byte
	mode    fs.FileMode
}

func (i *inode) open(name string, mode opMode) (fs.File, error) {
	if mode&opRead > 0 && i.mode&0o444 == 0 || mode&opWrite > 0 && i.mode&0o222 == 0 {
		return nil, fs.ErrPermission
	}

	return &file{
		name:   name,
		inode:  i,
		opMode: mode,
	}, nil
}

func (i *inode) bytes() ([]byte, error) {
	if i.mode&0o444 == 0 {
		return nil, fs.ErrPermission
	}

	return i.data, nil
}

func (i *inode) setMode(mode fs.FileMode) {
	i.mode = i.mode&fs.ModeSymlink | mode
}

func (i *inode) setTimes(_, mtime time.Time) {
	i.modtime = mtime
}

func (i *inode) seal() directoryEntry {
	return i
}

func (i *inode) getEntry(_ string) (*dirEnt, error) {
	return nil, fs.ErrInvalid
}

type file struct {
	name string
	*inode
	opMode   opMode
	lastRead uint8
	pos      int64
}

func (f *file) validTo(op string, m opMode, needValidPos bool) error {
	if f.opMode == opClose {
		return &fs.PathError{
			Op:   op,
			Path: f.name,
			Err:  fs.ErrClosed,
		}
	}

	if f.opMode&m != m {
		return &fs.PathError{
			Op:   op,
			Path: f.name,
			Err:  fs.ErrInvalid,
		}
	}

	if needValidPos && f.pos >= int64(len(f.data)) {
		return io.EOF
	}

	return nil
}

func (f *file) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Read(p []byte) (int, error) {
	if err := f.validTo("read", opRead, true); err != nil {
		return 0, err
	}

	n := copy(p, f.data[f.pos:])

	f.pos += int64(n)
	f.lastRead = 0

	return n, nil
}

func (f *file) ReadAt(p []byte, off int64) (int, error) {
	if err := f.validTo("readat", opRead|opSeek, false); err != nil {
		return 0, err
	}

	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n := copy(p, f.data[off:])

	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

func (f *file) ReadByte() (byte, error) {
	if err := f.validTo("readbyte", opRead, true); err != nil {
		return 0, err
	}

	b := f.data[f.pos]

	f.pos++
	f.lastRead = 1

	return b, nil
}

func (f *file) UnreadByte() error {
	if err := f.validTo("unreadbyte", opRead|opSeek, false); err != nil {
		return err
	}

	if f.lastRead != 1 {
		return &fs.PathError{
			Op:   "unreadbyte",
			Path: f.name,
			Err:  fs.ErrInvalid,
		}
	}

	f.lastRead = 0
	f.pos--

	return nil
}

func (f *file) ReadRune() (rune, int, error) {
	if err := f.validTo("readrune", opRead, true); err != nil {
		return 0, 0, err
	}

	r, s := utf8.DecodeRune(f.data[f.pos:])

	f.lastRead = uint8(s)
	f.pos += int64(s)

	return r, s, nil
}

func (f *file) UnreadRune() error {
	if err := f.validTo("unreadrune", opRead|opSeek, false); err != nil {
		return err
	}

	if f.lastRead == 0 {
		return &fs.PathError{
			Op:   "unreadrune",
			Path: f.name,
			Err:  fs.ErrInvalid,
		}
	}

	f.pos -= int64(f.lastRead)
	f.lastRead = 0

	return nil
}

func (f *file) WriteTo(w io.Writer) (int64, error) {
	if err := f.validTo("writeto", opRead, true); err != nil {
		return 0, err
	}

	n, err := w.Write(f.data[f.pos:])
	f.pos += int64(n)
	f.lastRead = 0

	return int64(n), err
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if err := f.validTo("seek", opSeek, false); err != nil {
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
		return 0, &fs.PathError{
			Op:   "seek",
			Path: f.name,
			Err:  fs.ErrInvalid,
		}
	}

	f.lastRead = 0

	if f.pos < 0 {
		f.pos = 0

		return f.pos, &fs.PathError{
			Op:   "seek",
			Path: f.name,
			Err:  fs.ErrInvalid,
		}
	}

	return f.pos, nil
}

func (f *file) Close() error {
	err := f.validTo("close", opClose, false)

	f.opMode = opClose
	f.pos = 0
	f.lastRead = 0

	return err
}

func (f *file) Name() string {
	return f.name
}

func (i *inode) Size() int64 {
	return int64(len(i.data))
}

func (i *inode) Type() fs.FileMode {
	return i.mode.Type()
}

func (i *inode) Mode() fs.FileMode {
	return i.mode
}

func (i *inode) ModTime() time.Time {
	return i.modtime
}

func (i *inode) IsDir() bool {
	return false
}

func (f *file) Sys() any {
	return f
}

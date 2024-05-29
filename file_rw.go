package memfs

import (
	"errors"
	"io"
	"io/fs"
	"sync"
	"time"
	"unicode/utf8"
)

type inodeRW struct {
	inode
	mu sync.RWMutex
}

func (i *inodeRW) open(name string, mode opMode) (fs.File, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if mode&opRead > 0 && i.mode&modeRead == 0 || mode&opWrite > 0 && i.mode&modeWrite == 0 {
		return nil, fs.ErrPermission
	}

	return &File{
		mu: &i.mu,
		file: file{
			name:   name,
			inode:  &i.inode,
			opMode: mode,
		},
	}, nil
}

func (i *inodeRW) bytes() ([]byte, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.inode.bytes()
}

func (i *inodeRW) string() (string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.inode.string()
}

func (i *inodeRW) setMode(mode fs.FileMode) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.mode = i.mode&fs.ModeSymlink | mode
}

func (i *inodeRW) setTimes(_, mtime time.Time) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.modtime = mtime
}

func (i *inodeRW) seal() directoryEntry {
	i.mu.Lock()
	defer i.mu.Unlock()

	de := i.inode
	i.inode = inode{}

	return &de
}

func (i *inodeRW) Size() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return int64(len(i.data))
}

func (i *inodeRW) Type() fs.FileMode {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.mode
}

func (i *inodeRW) Mode() fs.FileMode {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.mode
}

func (i *inodeRW) ModTime() time.Time {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.modtime
}

// File represents an open file, that can be used for reading and writing
// (depending on how it was opened).
//
// The file locks when making any changes, and so can be safely used from
// multiple goroutines.
type File struct {
	mu *sync.RWMutex
	file
}

func (f *File) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.Read(p)
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.file.ReadAt(p, off)
}

func (f *File) ReadByte() (byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.ReadByte()
}

func (f *File) UnreadByte() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.UnreadByte()
}

func (f *File) ReadRune() (rune, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.ReadRune()
}

func (f *File) UnreadRune() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.UnreadRune()
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opRead, true); err != nil {
		return 0, err
	}

	data := f.data[f.pos:]

	n, err := w.Write(append(make([]byte, 0, len(data)), data...))
	f.pos += int64(n)
	f.lastRead = 0

	return int64(n), err
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.Seek(offset, whence)
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.file.Close()
}

func (f *File) grow(size int) {
	if size > len(f.data) {
		if size < cap(f.data) {
			f.data = (f.data)[:size]
		} else {
			var newData []byte

			const simpleGrowLimit = 512

			if len(f.data) < simpleGrowLimit {
				newData = make([]byte, size, size<<1)
			} else {
				const growShift = 2

				newData = make([]byte, size, size+(size>>growShift))
			}

			copy(newData, f.data)
			f.data = newData
		}
	}
}

func (f *File) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite, false); err != nil {
		return 0, err
	}

	f.grow(int(f.pos) + len(p))

	n := copy(f.data[f.pos:], p)
	f.pos += int64(n)
	f.lastRead = 0
	f.modtime = time.Now()

	return n, nil
}

func (f *File) WriteAt(p []byte, off int64) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite|opSeek, false); err != nil {
		return 0, err
	}

	f.grow(int(off) + len(p))

	n := copy(f.data[off:], p)
	f.modtime = time.Now()

	return n, nil
}

func (f *File) WriteString(str string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite, false); err != nil {
		return 0, err
	}

	f.grow(int(f.pos) + len(str))

	n := copy(f.data[f.pos:], str)
	f.pos += int64(n)
	f.lastRead = 0
	f.modtime = time.Now()

	return n, nil
}

func (f *File) WriteByte(c byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite, false); err != nil {
		return err
	}

	f.grow(int(f.pos) + 1)

	f.data[f.pos] = c
	f.pos++
	f.lastRead = 0
	f.modtime = time.Now()

	return nil
}

func (f *File) WriteRune(r rune) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite, false); err != nil {
		return 0, err
	}

	p := utf8.AppendRune([]byte{}, r)

	f.grow(int(f.pos) + len(p))

	n := copy(f.data[f.pos:], p)
	f.pos += int64(n)
	f.lastRead = 0
	f.modtime = time.Now()

	return n, nil
}

func (f *File) ReadFrom(r io.Reader) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.validTo(opWrite, false); err != nil {
		return 0, err
	}

	var count int64

	for {
		f.grow(int(f.pos + 1))

		n, err := r.Read(f.data[f.pos:cap(f.data)])

		count += int64(n)
		f.pos += int64(n)
		f.data = f.data[:f.pos]

		if errors.Is(err, io.EOF) {
			return count, nil
		}

		if err != nil {
			return count, err
		}
	}
}

func (f *File) handleOpenMode(mode Mode) {
	if mode&Truncate != 0 {
		f.data = f.data[:0]
		f.modtime = time.Now()
	}

	if mode&Append != 0 {
		f.pos = int64(len(f.data))
	}
}

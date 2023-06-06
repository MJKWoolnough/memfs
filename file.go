package memfs

import (
	"io"
	"io/fs"
	"time"
)

type file struct{}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Read(p []byte) (int, error) {
	return 0, nil
}

func (f *file) ReadAt(p []byte, off int64) (int, error) {
	return 0, nil
}

func (f *file) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

func (f *file) ReadByte() (byte, error) {
	return 0, nil
}

func (f *file) UnreadByte() error {
	return nil
}

func (f *file) ReadRune() (rune, int, error) {
	return 0, 0, nil
}

func (f *file) UnreadRune() error {
	return nil
}

func (f *file) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
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

# memfs
--
    import "vimagination.zapto.org/memfs"

Package memfs contains both ReadOnly and ReadWrite implementations of an in
memory FileSystem, supporting all of the FS interfaces and more.

## Usage

#### type FS

```go
type FS struct {
}
```

FS represents an in-memory fs.FS implementation, with additional methods for a
more 'OS' like experience.

#### func  New

```go
func New() *FS
```

#### func (*FS) Chmod

```go
func (f *FS) Chmod(path string, mode fs.FileMode) error
```

#### func (*FS) Chown

```go
func (f *FS) Chown(path string, _, _ int) error
```

#### func (*FS) Chtimes

```go
func (f *FS) Chtimes(path string, atime time.Time, mtime time.Time) error
```

#### func (*FS) Create

```go
func (f *FS) Create(path string) (*File, error)
```

#### func (*FS) LStat

```go
func (f *FS) LStat(path string) (fs.FileInfo, error)
```

#### func (*FS) Lchown

```go
func (f *FS) Lchown(path string, _, _ int) error
```

#### func (*FS) Lchtimes

```go
func (f *FS) Lchtimes(path string, atime time.Time, mtime time.Time) error
```

#### func (*FS) Link

```go
func (f *FS) Link(oldPath, newPath string) error
```

#### func (*FS) Mkdir

```go
func (f *FS) Mkdir(path string, perm fs.FileMode) error
```

#### func (*FS) MkdirAll

```go
func (f *FS) MkdirAll(p string, perm fs.FileMode) error
```

#### func (*FS) Open

```go
func (f *FS) Open(p string) (fs.File, error)
```

#### func (*FS) OpenFile

```go
func (f *FS) OpenFile(path string, mode Mode, perm fs.FileMode) (*File, error)
```

#### func (*FS) ReadDir

```go
func (f *FS) ReadDir(path string) ([]fs.DirEntry, error)
```

#### func (*FS) ReadFile

```go
func (f *FS) ReadFile(path string) ([]byte, error)
```

#### func (*FS) Readlink

```go
func (f *FS) Readlink(path string) (string, error)
```

#### func (*FS) Remove

```go
func (f *FS) Remove(path string) error
```

#### func (*FS) RemoveAll

```go
func (f *FS) RemoveAll(path string) error
```

#### func (*FS) Rename

```go
func (f *FS) Rename(oldPath, newPath string) error
```

#### func (*FS) Seal

```go
func (f *FS) Seal() FSRO
```
Seal converts the Read-Write FS into a Read-only one.

The resulting FSRO cannot be changed, and has no locking. As the current
implementation doesn't copy any data, it destroys the current FS in order to
remove the need for locks on the resulting FSRO.

#### func (*FS) Stat

```go
func (f *FS) Stat(path string) (fs.FileInfo, error)
```

#### func (*FS) Sub

```go
func (f *FS) Sub(path string) (fs.FS, error)
```

#### func (*FS) Symlink

```go
func (f *FS) Symlink(oldPath, newPath string) error
```

#### type FSRO

```go
type FSRO interface {
	fs.FS
	fs.ReadDirFS
	fs.ReadFileFS
	fs.StatFS
	fs.SubFS
	LStat(path string) (fs.FileInfo, error)
	Readlink(path string) (string, error)
}
```

FSRO represents all of the methods on a read-only FS implementation.

#### type File

```go
type File struct {
}
```

File represents an open file, that can be used for reading and writing
(depending on how it was opened).

The file locks when making any changes, and so can be safely used from multiple
goroutines.

#### func (*File) Close

```go
func (f *File) Close() error
```

#### func (*File) Info

```go
func (f *File) Info() (fs.FileInfo, error)
```

#### func (*File) Name

```go
func (f *File) Name() string
```

#### func (*File) Read

```go
func (f *File) Read(p []byte) (int, error)
```

#### func (*File) ReadAt

```go
func (f *File) ReadAt(p []byte, off int64) (int, error)
```

#### func (*File) ReadByte

```go
func (f *File) ReadByte() (byte, error)
```

#### func (*File) ReadFrom

```go
func (f *File) ReadFrom(r io.Reader) (int64, error)
```

#### func (*File) ReadRune

```go
func (f *File) ReadRune() (rune, int, error)
```

#### func (*File) Seek

```go
func (f *File) Seek(offset int64, whence int) (int64, error)
```

#### func (*File) Stat

```go
func (f *File) Stat() (fs.FileInfo, error)
```

#### func (*File) String

```go
func (f *File) String() string
```

#### func (*File) Sys

```go
func (f *File) Sys() any
```

#### func (*File) UnreadByte

```go
func (f *File) UnreadByte() error
```

#### func (*File) UnreadRune

```go
func (f *File) UnreadRune() error
```

#### func (*File) Write

```go
func (f *File) Write(p []byte) (int, error)
```

#### func (*File) WriteAt

```go
func (f *File) WriteAt(p []byte, off int64) (int, error)
```

#### func (*File) WriteByte

```go
func (f *File) WriteByte(c byte) error
```

#### func (*File) WriteRune

```go
func (f *File) WriteRune(r rune) (int, error)
```

#### func (*File) WriteString

```go
func (f *File) WriteString(str string) (int, error)
```

#### func (*File) WriteTo

```go
func (f *File) WriteTo(w io.Writer) (int64, error)
```

#### type Mode

```go
type Mode uint8
```

Mode is used to determine how a file is opened.

Each value of Mode matches the intention of its similarly named OS counterpart.

```go
const (
	ReadOnly Mode = 1 << iota
	WriteOnly
	Append
	Create
	Excl
	Truncate

	ReadWrite = ReadOnly | WriteOnly
)
```

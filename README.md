# memfs
--
    import "vimagination.zapto.org/memfs"


## Usage

#### type FS

```go
type FS struct {
}
```


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
func (f *FS) Chown(path string, uid, gid int) error
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
func (f *FS) Lchown(path string, uid, gid int) error
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
func (f *FS) MkdirAll(path string, perm fs.FileMode) error
```

#### func (*FS) Open

```go
func (f *FS) Open(path string) (fs.File, error)
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


#### type File

```go
type File struct {
}
```


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

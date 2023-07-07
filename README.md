# memfs
--
    import "vimagination.zapto.org/memfs"


## Usage

#### type FS

```go
type FS struct {
}
```


#### func (*FS) LStat

```go
func (f *FS) LStat(path string) (fs.FileInfo, error)
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

#### func (*FS) Stat

```go
func (f *FS) Stat(path string) (fs.FileInfo, error)
```

#### type FSRW

```go
type FSRW struct {
	FS
}
```


#### func  New

```go
func New() *FSRW
```

#### func (*FSRW) Chmod

```go
func (f *FSRW) Chmod(path string, mode fs.FileMode) error
```

#### func (*FSRW) Chown

```go
func (f *FSRW) Chown(path string, uid, gid int) error
```

#### func (*FSRW) Chtimes

```go
func (f *FSRW) Chtimes(path string, atime time.Time, mtime time.Time) error
```

#### func (*FSRW) Create

```go
func (f *FSRW) Create(path string) (File, error)
```

#### func (*FSRW) LStat

```go
func (f *FSRW) LStat(path string) (fs.FileInfo, error)
```

#### func (*FSRW) Lchown

```go
func (f *FSRW) Lchown(path string, uid, gid int) error
```

#### func (*FSRW) Lchtimes

```go
func (f *FSRW) Lchtimes(path string, atime time.Time, mtime time.Time) error
```

#### func (*FSRW) Link

```go
func (f *FSRW) Link(oldPath, newPath string) error
```

#### func (*FSRW) Mkdir

```go
func (f *FSRW) Mkdir(path string, perm fs.FileMode) error
```

#### func (*FSRW) MkdirAll

```go
func (f *FSRW) MkdirAll(path string, perm fs.FileMode) error
```

#### func (*FSRW) ReadDir

```go
func (f *FSRW) ReadDir(path string) ([]fs.DirEntry, error)
```

#### func (*FSRW) ReadFile

```go
func (f *FSRW) ReadFile(path string) ([]byte, error)
```

#### func (*FSRW) Readlink

```go
func (f *FSRW) Readlink(path string) (string, error)
```

#### func (*FSRW) Remove

```go
func (f *FSRW) Remove(path string) error
```

#### func (*FSRW) RemoveAll

```go
func (f *FSRW) RemoveAll(path string) error
```

#### func (*FSRW) Rename

```go
func (f *FSRW) Rename(oldPath, newPath string) error
```

#### func (*FSRW) Stat

```go
func (f *FSRW) Stat(path string) (fs.FileInfo, error)
```

#### func (*FSRW) Symlink

```go
func (f *FSRW) Symlink(oldPath, newPath string) error
```

#### type File

```go
type File interface {
	fs.File
	Write([]byte) (int, error)
}
```

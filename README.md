# memfs

[![CI](https://github.com/MJKWoolnough/memfs/actions/workflows/go-checks.yml/badge.svg)](https://github.com/MJKWoolnough/memfs/actions)
[![Go Reference](https://pkg.go.dev/badge/vimagination.zapto.org/memfs.svg)](https://pkg.go.dev/vimagination.zapto.org/memfs)
[![Go Report Card](https://goreportcard.com/badge/vimagination.zapto.org/memfs)](https://goreportcard.com/report/vimagination.zapto.org/memfs)

--
    import "vimagination.zapto.org/memfs"

Package memfs contains both ReadOnly and ReadWrite implementations of an in-memory FileSystem, supporting all of the FS interfaces and more.

## Highlights

 - Full in-memory file system implementation.
 - Implements `fs.FS`, `fs.ReadDirFS`, `fs.ReadDirFile`, `fs.ReadFileFS`, `fs.StatFS`, and `fs.SubFS`.
 - Supports symlinks and hardlinks.
 - Handles metadata operations and permissions.
 - Can convert a read-write FS into a read-only one to remove locking and related overhead.

## Usage

```go
package main

import (
	"fmt"
	"io"

	"vimagination.zapto.org/memfs"
)

func main() {
	fs := memfs.New()

	fs.Mkdir("example", 0700)

	file, _ := fs.Create("example/file.txt")
	io.WriteString(file, "Hello, World!")
	file.Close()

	contents, _ := fs.ReadFile("example/file.txt")

	fmt.Printf("read: %q", contents)

	// Output:
	// read: "Hello, World!"
}
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/vimagination.zapto.org/memfs

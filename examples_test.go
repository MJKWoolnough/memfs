package memfs_test

import (
	"fmt"
	"io"

	"vimagination.zapto.org/memfs"
)

func Example() {
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

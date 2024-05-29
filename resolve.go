package memfs

import (
	"io/fs"
	"path"
	"strings"
)

type resolver struct {
	fullPath, path     string
	cutAt              int
	redirectsRemaining uint8
}

func (f *fsRO) getEntryWithoutCheck(path string) (directoryEntry, error) {
	r := resolver{
		fullPath:           path,
		path:               path,
		redirectsRemaining: maxRedirects,
	}

	return r.resolve(f.de)
}

func (r *resolver) resolve(root directoryEntry) (directoryEntry, error) {
	curr := root

	for r.path != "" {
		if curr.Mode()&modeRead == 0 {
			return nil, fs.ErrPermission
		} else if name := r.splitOffNamePart(); isEmptyName(name) {
			continue
		} else if next, err := curr.getEntry(name); err != nil {
			return nil, err
		} else if next.Mode()&fs.ModeSymlink == 0 {
			curr = next.directoryEntry

			continue
		} else if err = r.handleSymlink(next); err != nil {
			return nil, err
		}

		curr = root
	}

	return curr, nil
}

func (r *resolver) splitOffNamePart() string {
	slashPos := strings.Index(r.path, "/")

	var name string

	if slashPos == -1 {
		name, r.path = r.path, ""
	} else {
		name, r.path = r.path[:slashPos], r.path[slashPos+1:]
		r.cutAt += slashPos + 1
	}

	return name
}

func isEmptyName(name string) bool {
	return name == "" || name == "."
}

func (r *resolver) handleSymlink(sym *dirEnt) error {
	r.redirectsRemaining--
	if r.redirectsRemaining == 0 {
		return fs.ErrInvalid
	}

	symPath, err := sym.string()
	if err != nil {
		return err
	} else if strings.HasPrefix(symPath, "/") {
		r.fullPath = path.Clean(symPath)[1:]
	} else if r.path == "" {
		r.fullPath = path.Join(r.fullPath[:r.cutAt], symPath)
	} else {
		r.fullPath = path.Join(r.fullPath[:r.cutAt-len(sym.name)-1], symPath, r.path)
	}

	r.path = r.fullPath
	r.cutAt = 0

	return nil
}

package searchpath

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrAbort can be returned from a visitation func to terminate iteration.
var ErrAbort = errors.New("abort visitation")

func visitFile(path string, visit func(f *os.File) error) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	return visit(f)
}

// Visit visits all searchpaths looking for extant files at relative path relpath, opening each file and calling visitFn
// on it (in searchpath order). The files are closed by Visit.
// If visitFn returns ErrAbort, Visit stops traversal and returns nil.
func Visit(relpath string, visitFn func(f *os.File) error) error {
	for _, prefix := range searchPaths {
		if err := visitFile(filepath.Join(prefix, relpath), visitFn); err != nil {
			if err == ErrAbort {
				err = nil
			}
			return err
		}
	}
	return nil
}

func visitFilepath(path string, visit func(fp string) error) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return visit(path)
}

// VisitPaths is as Visit, but only passes the filepath to visitFn, without opening it.
func VisitPaths(relath string, visitFn func(fp string) error) error {
	for _, prefix := range searchPaths {
		if err := visitFilepath(filepath.Join(prefix, relath), visitFn); err != nil {
			if err == ErrAbort {
				err = nil
			}
			return err
		}
	}
	return nil
}

var searchPaths []string

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, home)
	}
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, cwd)
	}
}

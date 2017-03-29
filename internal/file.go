package internal

import (
	"io/ioutil"
	"path/filepath"
)

// IsArcFile returns true if a file has the .arc file extension.
func IsArcFile(filename string) bool {
	return filepath.Ext(filename) == ".arc"
}

// ReadCurDir has the same signature and functionality as ReadDir, but reads
// from the current working directory.
func ReadCurDir() ([]string, error) {
	return ReadDir(".")
}

// ReadDir reads a directory and returns a list of ARC files in that directory.
func ReadDir(dirname string) ([]string, error) {
	list, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, file := range list {
		if !file.IsDir() && IsArcFile(file.Name()) {
			files = append(files, file.Name())
		}
	}
	return files, nil
}

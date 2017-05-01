package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// IsArcFile returns true if a file has the .arc file extension.
func IsArcFile(filename string) bool {
	return filepath.Ext(filename) == ".arc"
}

// IsDirectory returns true if the path is a directory.
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

// ReadCurDir has the same signature and functionality as ReadDir, but reads
// from the current working directory.
func ReadCurDir() ([]string, error) {
	return ReadDir(".")
}

// ReadDir reads a directory and returns a list of ARC files in that directory.
// An error is returned if ioutil.ReadDir() fails.
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

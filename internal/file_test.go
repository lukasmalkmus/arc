package internal

import (
	"os"
	"testing"
)

func TestIsArcFile(t *testing.T) {
	tests := []struct {
		filename  string
		isArcFile bool
	}{
		{"file.arc", true},
		{"file.arc.arc", true},
		{"file.ar", false},
		{"file.arc.ar", false},
		{"file", false},
	}

	for _, tt := range tests {
		equals(t, tt.isArcFile, IsArcFile(tt.filename))
	}
}

func TestIsDirectory(t *testing.T) {
	err := os.Chdir("../testdata")
	if err != nil {
		t.Error("could not switch to testdata directory")
	}

	tests := []struct {
		filename string
		isDir    bool
	}{
		{".", true},
		{"./", true},
		{".", true},
		{"sub", true},
		{"./sub", true},
		{"./sub/", true},
		{"valid.arc", false},
		{"./valid.arc", false},
	}

	for _, tt := range tests {
		is, err := IsDirectory(tt.filename)
		ok(t, err)
		equals(t, tt.isDir, is)
	}
}

func TestReadDir(t *testing.T) {
	err := os.Chdir("../testdata")
	if err != nil {
		t.Error("could not switch to testdata directory")
	}

	tests := []struct {
		folder string
		files  []string
	}{
		{".", []string{"valid.arc"}},
		{"sub", []string{}},
	}

	for _, tt := range tests {
		files, err := ReadDir(tt.folder)
		ok(t, err)
		equals(t, tt.files, files)
	}
}

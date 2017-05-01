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
	err := os.Chdir("..")
	if err != nil {
		panic(err)
	}

	tests := []struct {
		filename string
		isDir    bool
	}{
		{"testdata", true},
		{"testdata/", true},
		{"testdata/valid.arc", false},
	}

	for _, tt := range tests {
		is, err := IsDirectory(tt.filename)
		ok(t, err)
		equals(t, tt.isDir, is)
	}
}

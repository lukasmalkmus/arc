package internal

import (
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

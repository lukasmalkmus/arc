package token

import (
	"testing"
)

func TestPosition_String(t *testing.T) {
	tests := []struct {
		str string
		pos Pos
	}{
		{"INVALID POSITION", Pos{}},
		{"INVALID POSITION", Pos{Line: 1}},
		{"INVALID POSITION", Pos{Char: 1}},
		{"token.go", Pos{Filename: "token.go"}},
		{"token.go:1:0", Pos{Filename: "token.go", Line: 1}},
		{"token.go:0:1", Pos{Filename: "token.go", Char: 1}},
		{"1:1", Pos{Filename: "", Line: 1, Char: 1}},
		{"token.go:2:3", Pos{Filename: "token.go", Line: 2, Char: 3}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.String())
	}
}

func TestPosition_NoFile(t *testing.T) {
	tests := []struct {
		str string
		pos Pos
	}{
		{"0:0", Pos{}},
		{"1:0", Pos{Line: 1}},
		{"0:1", Pos{Char: 1}},
		{"0:0", Pos{Filename: "token.go"}},
		{"1:0", Pos{Filename: "token.go", Line: 1}},
		{"0:1", Pos{Filename: "token.go", Char: 1}},
		{"1:1", Pos{Filename: "", Line: 1, Char: 1}},
		{"2:3", Pos{Filename: "token.go", Line: 2, Char: 3}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.NoFile())
	}
}

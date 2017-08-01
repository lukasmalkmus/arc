package token_test

import (
	"testing"

	"github.com/lukasmalkmus/arc/token"
)

func TestPosition_String(t *testing.T) {
	tests := []struct {
		str string
		pos token.Pos
	}{
		{"INVALID POSITION", token.Pos{}},
		{"1:0", token.Pos{Line: 1}},
		{"0:1", token.Pos{Char: 1}},
		{"token.go", token.Pos{Filename: "token.go"}},
		{"token.go:1:0", token.Pos{Filename: "token.go", Line: 1}},
		{"token.go:0:1", token.Pos{Filename: "token.go", Char: 1}},
		{"1:1", token.Pos{Filename: "", Line: 1, Char: 1}},
		{"token.go:2:3", token.Pos{Filename: "token.go", Line: 2, Char: 3}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.String())
	}
}

func TestPosition_NoFile(t *testing.T) {
	tests := []struct {
		str string
		pos token.Pos
	}{
		{"0:0", token.Pos{}},
		{"1:0", token.Pos{Line: 1}},
		{"0:1", token.Pos{Char: 1}},
		{"0:0", token.Pos{Filename: "token.go"}},
		{"1:0", token.Pos{Filename: "token.go", Line: 1}},
		{"0:1", token.Pos{Filename: "token.go", Char: 1}},
		{"1:1", token.Pos{Filename: "", Line: 1, Char: 1}},
		{"2:3", token.Pos{Filename: "token.go", Line: 2, Char: 3}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.NoFile())
	}
}

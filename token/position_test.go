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
		{"1:1", Pos{Filename: "", Line: 1, Char: 1}},
		{"token.go:2:3", Pos{Filename: "token.go", Line: 2, Char: 3}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.String())
	}
}

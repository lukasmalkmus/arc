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
		{"line 1", Pos{Filename: "", Line: 1}},
		{"token.go: line 1", Pos{Filename: "token.go", Line: 1}},
	}

	for _, tt := range tests {
		equals(t, tt.str, tt.pos.String())
	}
}

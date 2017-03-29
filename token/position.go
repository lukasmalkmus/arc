package token

import "fmt"

// Pos is the position of a token in a source string.
type Pos struct {
	Filename string
	Line     int
	Char     int
}

// String returns a string representation of the Position.
func (p Pos) String() string {
	if p.Line == 0 || p.Char == 0 {
		return fmt.Sprintf("INVALID POSITION")
	}
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Char)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Char)
}

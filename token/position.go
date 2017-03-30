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
	if (p.Line == 0 || p.Char == 0) && p.Filename == "" {
		return fmt.Sprintf("INVALID POSITION")
	} else if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Char)
	} else if p.Line == 0 && p.Char == 0 && p.Filename != "" {
		return p.Filename
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Char)
}

// NoFile returns a string representation of the Position without the filename.
func (p Pos) NoFile() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Char)
}

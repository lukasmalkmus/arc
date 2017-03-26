package fmt

import (
	"io"
	"io/ioutil"

	"github.com/LukasMa/arc/ast"
	"github.com/LukasMa/arc/parser"
)

// Formater formats ARC source code.
type Formater struct {
	prog *ast.Program
}

// New returns a new ARC formater. It operates on the AST of an ARC program.
func New(prog *ast.Program) *Formater {
	return &Formater{
		prog: prog,
	}
}

// Format will format ARC source code. The function takes the source from an
// io.Reader as parameter. It returns the formated program as a slice of bytes.
// An error is returned if formating fails.
func Format(src io.Reader) ([]byte, error) {
	// Parse source.
	prog, err := parser.New(src).Parse()
	if err != nil {
		return nil, err
	}

	return New(prog).Format()
}

// FormatFile will format an ARC source file. The function takes a filename as
// parameter. The formated program will be written back to the source file. The
// function returns an error if formating fails.
func FormatFile(filename string) error {
	// Parse source file.
	prog, err := parser.ParseFile(filename)
	if err != nil {
		return err
	}

	code, err := New(prog).Format()
	if err != nil {
		return err
	}

	// Write formated code back to source file.
	return ioutil.WriteFile(filename, code, 0644)
}

// Format will format ARC source code. The function returns the formated program
// as a slice of bytes. An error is returned if formating fails.
func (f *Formater) Format() ([]byte, error) {
	return []byte(f.prog.String()), nil
}

/*
Package fmt provides an ARC source code formater. The formater operates on the
AST of an ARC program and therefore relies on the parser. This means, it formats
an ARC statement and returns its string representation. So formating doesn't
mean modification of the actual source code. It is a complete rewriting of the
program.
NOTE: At the moment, it is not possible to format an invalid ARC program. The
parser must not return an error to start formatting.
*/
package fmt

import (
	"io"
	"io/ioutil"

	"github.com/lukasmalkmus/arc/ast"
	"github.com/lukasmalkmus/arc/internal"
	"github.com/lukasmalkmus/arc/parser"
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
	errs := internal.MultiError{}

	// TODO: If the parser can handle invalid source code, we can continue and
	// format the invalid program, keeping the invalid code segment intact for
	// the user to correct.
	// Parse source.
	prog, err := parser.New(src).Parse()
	//errs.Add(err)
	if err != nil {
		return nil, err
	}

	code, err := New(prog).Format()
	if err != nil {
		errs.Add(err)
		return nil, errs.Return()
	}

	return code, errs.Return()
}

// FormatFile will format an ARC source file. The function takes a filename as
// parameter. The formated program will be written back to the source file. The
// function returns an error if formating fails.
func FormatFile(filename string) error {
	errs := internal.MultiError{}

	// TODO: If the parser can handle invalid source code, we can continue and
	// format the invalid program, keeping the invalid code segment intact for
	// the user to correct.
	// Parse source file.
	prog, err := parser.ParseFile(filename)
	//errs.Add(err)
	if err != nil {
		return err
	}

	code, err := New(prog).Format()
	if err != nil {
		errs.Add(err)
		return errs.Return()
	}

	// Write formated code back to source file.
	errs.Add(ioutil.WriteFile(filename, code, 0644))
	return errs.Return()
}

// Format will format ARC source code. The function returns the formated program
// as a slice of bytes. An error is returned if formating fails.
func (f *Formater) Format() ([]byte, error) {
	return []byte(f.prog.String()), nil
}

package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/LukasMa/arc/ast"
	"github.com/LukasMa/arc/parser"
)

// Options are configuration values for the Assembler.
type Options struct {
	// Log is where log messages will be written to.
	Log io.Writer

	// Verbose enables more verbose output.
	Verbose bool
}

// Assembler assembles ARC source code into machine code. It operates on the AST
// of an ARC program.
type Assembler struct {
	opts *Options
	prog *ast.Program
}

// New returns a new ARC assembler. It takes the source code as io.Reader as
// first parameter.
func New(prog *ast.Program, options *Options) *Assembler {
	a := &Assembler{
		opts: options,
		prog: prog,
	}

	// Set defaults.
	if a.opts == nil {
		a.opts = &Options{}
	}
	if a.opts.Log == nil {
		a.opts.Log = os.Stdout
	}

	return a
}

// Assemble will transform ARC source code into machine code. The function takes
// an io.Reader as source and an verbosity switch as parameters. The function
// returns the assembled program as a slice of bytes. An error is returned if
// assembling fails.
func Assemble(src io.Reader, options *Options) ([]byte, error) {
	// Parse source.
	prog, err := parser.New(src).Parse()
	if err != nil {
		return nil, err
	}

	return New(prog, options).Assemble()
}

// AssembleFile will transform an ARC source file into machine code. The
// function takes a filename and an switch for increased verbosity as
// parameters. It returns an error if assembling fails.
func AssembleFile(filename string, options *Options) error {
	// Parse source file.
	prog, err := parser.ParseFile(filename)
	if err != nil {
		return err
	}

	// Assemble source file.
	code, err := New(prog, options).Assemble()
	if err != nil {
		return err
	}

	// Evaluate destination file and write program to file.
	ext := filepath.Ext(filename)
	dest := filename[0 : len(filename)-len(ext)]
	return ioutil.WriteFile(dest, code, 0644)
}

// Assemble will transform ARC source code into machine code. The function
// returns the assembled program as a slice of bytes. An error is returned if
// assembling fails.
func (a *Assembler) Assemble() ([]byte, error) {
	return []byte(a.prog.String()), nil
}

// log is a helper function providing shorter and faster logging. It only logs
// when the verbose option is enabled.
func (a *Assembler) log(text string) {
	if a.opts.Verbose {
		fmt.Fprintln(a.opts.Log, text)
	}
}

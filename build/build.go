/*
Package build provides an ARC assembler. The assembler operates on the AST of an
ARC program and therefore relies on the parser.
*/
package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lukasmalkmus/arc/ast"
	"github.com/lukasmalkmus/arc/internal"
	"github.com/lukasmalkmus/arc/parser"
	"github.com/lukasmalkmus/arc/token"
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
	asm, err := New(prog, options).Assemble()
	if err != nil {
		return err
	}

	// Evaluate destination file and write program to file.
	ext := filepath.Ext(filename)
	dest := filename[0 : len(filename)-len(ext)]
	return ioutil.WriteFile(dest, asm, 0644)
}

// Assemble will transform ARC source code into machine code. The function
// returns the assembled program as a slice of bytes. An error is returned if
// assembling fails.
func (a *Assembler) Assemble() ([]byte, error) {
	// Reserve 33 bytes of memory per statement (32bit instruction where one bit
	// is represented by an ASCII char + 1 byte newline char).
	prog := make([]byte, 0, len(a.prog.Statements)*33)
	errs := internal.MultiError{}

	// Assemble the program line by line.
	for _, stmt := range a.prog.Statements {
		asm, err := a.AssembleStatement(stmt)
		if err != nil {
			errs.Add(err)
			continue
		}
		prog = append(prog, asm...)
		prog = append(prog, '\n')
	}

	return prog, errs.Return()
}

// AssembleStatement will assemble a Statement AST object into ARC assembly.
func (a *Assembler) AssembleStatement(stmt ast.Statement) ([]byte, error) {
	// Evaluate which statement to parse.
	switch stmt.(type) {
	case *ast.LoadStatement:
		return a.AssembleLoadStatement(stmt.(*ast.LoadStatement))
	}

	return nil, &AssemblerError{fmt.Sprintf("no assemble instructions defined for %q", stmt.Tok()), stmt.Pos()}
}

// AssembleLoadStatement will assemble a LoadStatement AST object into ARC
// assembly.
func (a *Assembler) AssembleLoadStatement(stmt *ast.LoadStatement) ([]byte, error) {
	asm := make([]byte, 0, 32)

	op, ok := LookupOpCode(stmt)
	if !ok {
		return nil, &AssemblerError{fmt.Sprintf("missing operation code in lookup table for %q", stmt.Tok()), stmt.Pos()}
	}
	asm = append(asm, op...)

	format, ok := LookupInstructionFormat(stmt)
	if !ok {
		return nil, &AssemblerError{fmt.Sprintf("missing instruction format in lookup table for %q", stmt.Tok()), stmt.Pos()}
	}
	asm = append(asm, format...)

	return asm, nil
}

// log is a helper function providing shorter and faster logging. It only logs
// when the verbose option is enabled.
func (a *Assembler) log(text string) {
	if a.opts.Verbose {
		fmt.Fprintln(a.opts.Log, text)
	}
}

// AssemblerError represents an error that occurred during parsing.
type AssemblerError struct {
	Message string
	Pos     token.Pos
}

// Error returns the string representation of the error. It implements the error
// interface.
func (e AssemblerError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

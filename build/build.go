package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/LukasMa/arc/parser"
)

// Assembler assembles ARC source code into machine code.
type Assembler struct {
	parser  *parser.Parser
	verbose bool
}

// New returns a new ARC assembler. It takes the source code as io.Reader as
// first parameter. Furthermore, more verbose output can be enabled by passing
// true as second parameter.
func New(src io.Reader, verbose bool) *Assembler {
	return &Assembler{
		parser:  parser.New(src),
		verbose: verbose,
	}
}

// AssembleFile will transform an ARC source file into machine code. The
// function takes a filename and an switch for increased verbosity as
// parameters. It returns an error if assembling fails.
func AssembleFile(srcFile string, verbose bool) error {
	// Read source file.
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error reading file %s: %e", srcFile, err)
	}
	defer src.Close()

	// Assemble source file.
	code, err := Assemble(src, verbose)
	if err != nil {
		return fmt.Errorf("error assembling file %s: %e", srcFile, err)
	}

	// Evaluate destination file and write program to file.
	ext := filepath.Ext(srcFile)
	destFile := srcFile[0 : len(srcFile)-len(ext)]
	return ioutil.WriteFile(destFile, code, 0644)
}

// Assemble will transform ARC source code into machine code. The function takes
// an io.Reader as source and an verbosity switch as parameters. The function
// returns the assembled program as a slice of bytes. An error is returned if
// assembling fails.
func Assemble(src io.Reader, verbose bool) ([]byte, error) {
	return New(src, verbose).Assemble()
}

// Assemble will transform ARC source code into machine code. The function
// returns the assembled program as a slice of bytes. An error is returned if
// assembling fails.
func (a *Assembler) Assemble() ([]byte, error) {
	prog, err := a.parser.Parse()
	if err != nil {
		return nil, err
	}
	return []byte(prog.String()), nil
}

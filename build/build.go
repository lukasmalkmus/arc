package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/LukasMa/arc/parser"
)

// Options are configuration values for the Assembler.
type Options struct {
	Verbose bool
}

// Assembler assembles ARC source code into machine code.
type Assembler struct {
	opts   *Options
	parser *parser.Parser
	dest   bytes.Buffer
}

// New returns a new ARC assembler. It takes the source code as io.Reader as
// first parameter.
func New(src io.Reader, options *Options) *Assembler {
	a := &Assembler{
		opts:   options,
		parser: parser.New(src),
	}

	// Set defaults.
	if a.opts == nil {
		a.opts = &Options{}
	}

	return a
}

// AssembleFile will transform an ARC source file into machine code. The
// function takes a filename and an switch for increased verbosity as
// parameters. It returns an error if assembling fails.
func AssembleFile(srcFile string, options *Options) error {
	// Read source file.
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error reading file %s: %e", srcFile, err)
	}
	defer src.Close()

	// Assemble source file.
	code, err := Assemble(src, options)
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
func Assemble(src io.Reader, options *Options) ([]byte, error) {
	return New(src, options).Assemble()
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

// log is a helper function providing shorter and faster logging. It only logs
// when the verbose option is enabled.
func (a *Assembler) log(text string) {
	if a.opts.Verbose {
		fmt.Println(text)
	}
}

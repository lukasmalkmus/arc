package fmt

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/LukasMa/arc/parser"
)

// Formater formats ARC source code.
type Formater struct {
	parser *parser.Parser
}

// New returns a new ARC formater. It takes the source code from an io.Reader as
// first parameter.
func New(src io.Reader) *Formater {
	return &Formater{
		parser: parser.New(src),
	}
}

// FormatFile will format an ARC source file. The function takes a filename as
// parameter. It returns an error if formating fails.
func FormatFile(srcFile string) error {
	// Read source file.
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error reading file %s:\n%s", srcFile, err.Error())
	}
	defer src.Close()

	// Assemble source file.
	code, err := Format(src)
	if err != nil {
		return fmt.Errorf("error formating file %s:\n%s", srcFile, err.Error())
	}

	// Write formated code to same file.
	return ioutil.WriteFile(src.Name(), code, 0644)
}

// Format will format ARC source code. The function takes the source from an
// io.Reader as parameter. It returns the formated program as a slice of bytes.
// An error is returned if formating fails.
func Format(src io.Reader) ([]byte, error) {
	return New(src).Format()
}

// Format will format ARC source code. The function returns the formated program
// as a slice of bytes. An error is returned if formating fails.
func (f *Formater) Format() ([]byte, error) {
	prog, err := f.parser.Parse()
	if err != nil {
		return nil, err
	}
	return []byte(prog.String()), nil
}

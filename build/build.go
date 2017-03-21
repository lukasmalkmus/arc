package build

import (
	"bytes"
	"io"

	"github.com/LukasMa/arc/parser"
)

// Assemble will transform ARC source code into machine code. The function takes
// an io.Reader as source for the ARC source code. It writes the assembled
// program to an io.Writer destination.
func Assemble(src io.Reader) (io.Writer, error) {
	p := parser.New(src)
	prog, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return bytes.NewBufferString(prog.String()), nil
}

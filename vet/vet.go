package vet

import (
	"io"

	"github.com/LukasMa/arc/parser"
)

// Options are configuration values for the Vet.
type Options struct {
	Fix bool
}

// Vet examines ARC source code and reports suspicious language constructs.
type Vet struct {
	opts   *Options
	parser *parser.Parser
}

// New returns a new ARC Vet. It takes the source code as io.Reader as first
// parameter.
func New(src io.Reader, options *Options) *Vet {
	a := &Vet{
		opts:   options,
		parser: parser.New(src),
	}

	// Set defaults.
	if a.opts == nil {
		a.opts = &Options{}
	}

	return a
}

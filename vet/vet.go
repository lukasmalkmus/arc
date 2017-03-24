package vet

import (
	"fmt"
	"io"
	"os"

	"github.com/LukasMa/arc/parser"
	"github.com/LukasMa/arc/vet/check"
	_ "github.com/LukasMa/arc/vet/check/begEndCheck" // BeginEndCheck
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

// CheckFile will check an ARC source file. The function takes a filename as
// parameter. It returns an error if checking fails.
func CheckFile(srcFile string, options *Options) ([]string, error) {
	// Read source file.
	src, err := os.Open(srcFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %e", srcFile, err)
	}
	defer src.Close()

	// Check source file.
	res, err := Check(src, options)
	if err != nil {
		return nil, fmt.Errorf("error formating file %s: %e", srcFile, err)
	}

	// Return results
	return res, nil
}

// Check performs multiple checks on the ARC AST. It takes the source code from
// an io.Reader as parameter. Results are returned as a slice of strings. An
// error is returned if parsing of the source file or a check fails.
func Check(src io.Reader, options *Options) ([]string, error) {
	return New(src, options).Check()
}

// Check performs multiple checks on the ARC AST. Results are returned as a
// slice of strings. An error is returned if parsing of the source file or a
// check fails.
func (v *Vet) Check() ([]string, error) {
	_, err := v.parser.Parse()
	if err != nil {
		return nil, err
	}
	res := []string{}
	for name, check := range check.Checks() {
		r, err := check.Run()
		if err != nil {
			return nil, fmt.Errorf("check %s failed: %e", name, err)
		}
		res = append(res, r...)
	}
	return res, nil
}

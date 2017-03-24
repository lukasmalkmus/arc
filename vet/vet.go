package vet

import (
	"fmt"
	"io"
	"os"

	"github.com/LukasMa/arc/parser"
	"github.com/LukasMa/arc/vet/check"
)

// Options are configuration values for the Vet.
type Options struct {
	// Fix enables applying fixes on the source code.
	Fix bool

	// Checks is a slice of strings representing the checks to run on the source
	// code.
	Checks []string
}

// Vet examines ARC source code and reports suspicious language constructs.
type Vet struct {
	opts   *Options
	parser *parser.Parser
	checks map[string]check.Check
}

// New returns a new ARC Vet. It takes the source code as io.Reader as first
// parameter.
func New(src io.Reader, options *Options) (*Vet, error) {
	v := &Vet{
		opts:   options,
		parser: parser.New(src),
		checks: make(map[string]check.Check),
	}

	// Init empty config.
	if v.opts == nil {
		v.opts = &Options{}
	}

	// Empty slice means run all checks.
	if len(v.opts.Checks) == 0 {
		v.opts.Checks = check.List()
	}

	// Resolve enabled checks.
	for _, name := range v.opts.Checks {
		c, err := check.Get(name)
		if err != nil {
			return nil, err
		}
		v.checks[name] = c
	}

	return v, nil
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
// error is returned if the New() function, parsing of the source file or a
// check fails.
func Check(src io.Reader, options *Options) ([]string, error) {
	v, err := New(src, options)
	if err != nil {
		return nil, err
	}
	return v.Check()
}

// Check performs multiple checks on the ARC AST. Results are returned as a
// slice of strings. An error is returned if parsing of the source file or a
// check fails.
func (v *Vet) Check() ([]string, error) {
	// Parse source code.
	_, err := v.parser.Parse()
	if err != nil {
		return nil, err
	}

	// Run every enabled check.
	res := []string{}
	for name, check := range v.checks {
		// Run check.
		r, err := check.Run()
		if err != nil {
			return nil, fmt.Errorf("check %s failed: %e", name, err)
		}
		res = append(res, r...)
	}

	return res, nil
}

// EnabledChecks returns a slice of the enabled checks.
func (v Vet) EnabledChecks() []string {
	return v.opts.Checks
}

/*
Package vet examines ARC source code and reports suspicious language constructs.
It operates on the AST of an ARC program and therefore relies on the parser.
NOTE: At the moment, it is not possible to format an invalid ARC program. The
parser must not return an error to start formatting.
Different checks are implemented in the subdirectory vet/check and must register
themselves by calling vet.Register().
*/
package vet

import (
	"fmt"
	"io"
	"sort"

	"github.com/lukasmalkmus/arc/ast"
	"github.com/lukasmalkmus/arc/internal"
	"github.com/lukasmalkmus/arc/parser"
	"github.com/lukasmalkmus/arc/vet/check"
)

// Options are configuration values for the Vet.
type Options struct {
	// Checks is a slice of strings representing the checks to run on the source
	// code.
	Checks []string
	// Sort enables sorting vet results.
	Sort bool
}

// Vet examines ARC source code and reports suspicious language constructs. It
// operates on the AST of an ARC program.
type Vet struct {
	opts   *Options
	prog   *ast.Program
	checks map[string]check.Check
}

// New returns a new ARC Vet. It takes the source code as io.Reader as first
// parameter.
func New(prog *ast.Program, options *Options) (*Vet, error) {
	v := &Vet{
		opts:   options,
		prog:   prog,
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

// Check performs multiple checks on the ARC AST. It takes the source code from
// an io.Reader as parameter. Results are returned as a slice of strings. An
// error is returned if the New() function, parsing of the file or a check
// fails.
func Check(src io.Reader, options *Options) ([]string, error) {
	errs := internal.MultiError{}

	// Parse source. Abort if we don't have a program.
	prog, err := parser.New(src).Parse()
	if prog == nil {
		return nil, err
	}
	errs.Add(err)

	// Create new vet instance.
	v, err := New(prog, options)
	if err != nil {
		errs.Add(err)
		return nil, errs
	}

	// Vet program (run checks).
	res, err := v.Check()
	errs.Add(err)

	return res, errs
}

// CheckFile performs multiple checks on the ARC AST. It takes a filename as
// parameter. Results are returned as a slice of strings. An error is returned
// if the New() function, parsing of the file or a check fails.
func CheckFile(filename string, options *Options) ([]string, error) {
	errs := internal.MultiError{}

	// Parse source. Abort if we don't have a program.
	prog, err := parser.ParseFile(filename)
	if prog == nil {
		return nil, err
	}
	errs.Add(err)

	// Create new vet instance.
	v, err := New(prog, options)
	if err != nil {
		errs.Add(err)
		return nil, errs.Return()
	}

	// Vet program (run checks).
	res, err := v.Check()
	errs.Add(err)

	return res, errs.Return()
}

// Check performs multiple checks on the ARC AST. Results are returned as a
// slice of strings. An error is returned if parsing of the source file or a
// check fails.
func (v *Vet) Check() ([]string, error) {
	errs := internal.MultiError{}
	res := []string{}

	// Run every enabled check.
	for name, check := range v.checks {
		// Run check.
		r, err := check.Run(v.prog)
		if err != nil {
			errs.Add(fmt.Errorf("check %s failed: %e", name, err))
		}
		res = append(res, r...)
	}

	// Sort results if enabled.
	if v.opts.Sort {
		sort.Strings(res)
	}

	return res, errs.Return()
}

// EnabledChecks returns a slice of the enabled checks.
func (v Vet) EnabledChecks() []string {
	return v.opts.Checks
}

package vet

import (
	"fmt"
	"io"

	"github.com/LukasMa/arc/ast"
	"github.com/LukasMa/arc/parser"
	"github.com/LukasMa/arc/util"
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
	errs := util.MultiError{}

	// Parse source.
	prog, err := parser.New(src).Parse()
	errs.Add(err)

	// Create new vet.
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
	errs := util.MultiError{}

	// Parse source.
	prog, err := parser.ParseFile(filename)
	errs.Add(err)

	// Create new vet.
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
	errs := util.MultiError{}
	res := []string{}

	// Run every enabled check.
	for name, check := range v.checks {
		// Run check.
		r, err := check.Run()
		if err != nil {
			errs.Add(fmt.Errorf("check %s failed: %e", name, err))
		}
		res = append(res, r...)
	}

	return res, errs.Return()
}

// EnabledChecks returns a slice of the enabled checks.
func (v Vet) EnabledChecks() []string {
	return v.opts.Checks
}

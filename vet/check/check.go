/*
Package check provides an interface for checks as well as some generic helper
functions. Checks must satisfy the Check interface and register themselfs
by calling vet.Register().
*/
package check

import (
	"fmt"
	"sort"

	"github.com/lukasmalkmus/arc/ast"
	"github.com/lukasmalkmus/arc/token"
)

// Check is the interface that must implemented by checks.
type Check interface {
	// Desc returns a description of the check.
	Desc() string
	// Name returns the name of the check.
	Name() string
	// Run will execute the given check and return a slice of results. An error
	// is returned if the check fails.
	Run(*ast.Program) ([]string, error)
}

var checks = make(map[string]Check)

// Register makes a check available by the provided name. If Register is called
// twice with the same name or if check is nil, it panics.
func Register(check Check) {
	if check == nil {
		panic("check: Register check is nil")
	}
	if _, dup := checks[check.Name()]; dup {
		panic("check: Register called twice for check " + check.Name())
	}
	checks[check.Name()] = check
}

// Get looks up a registered check by its name. It returns an error if no check
// is registered on that name.
func Get(name string) (Check, error) {
	check, prs := checks[name]
	if !prs {
		return nil, fmt.Errorf("no check registered named %q", name)
	}
	return check, nil
}

// Desc returns a slice of all registered checks and their description.
func Desc() (res []string) {
	for name, check := range checks {
		res = append(res, fmt.Sprintf("%s - %s", name, check.Desc()))
	}
	sort.Strings(res)
	return res
}

// List returns a slice of all registered checks by their name.
func List() (res []string) {
	for name := range checks {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}

// buildMsg builds an appropriate message including the calling checks name.
func buildMsg(check Check, pos token.Pos, msg string) string {
	return fmt.Sprintf("%s: %s (%s)", pos, msg, check.Name())
}

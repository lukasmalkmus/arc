/*
Package check provides an interface for checks as well as some generic helper
functions. Checks must satisfy the Check interface and register themselfs
by calling vet.Register().
*/
package check

import (
	"fmt"
	"sort"
)

// Check is the interface that must implemented by checks.
type Check interface {
	// Desc returns a description of the check.
	Desc() string

	// Run will execute the given check and return a slice of results. An error
	// is returned if the check fails.
	Run() ([]string, error)
}

var checks = make(map[string]Check)

// Register makes a check available by the provided name. If Register is called
// twice with the same name or if check is nil, it panics.
func Register(name string, check Check) {
	if check == nil {
		panic("check: Register check is nil")
	}
	if _, dup := checks[name]; dup {
		panic("check: Register called twice for check " + name)
	}
	checks[name] = check
}

// Get looks up a registered check by its name. It returns an error if no check
// is registered on that name.
func Get(name string) (Check, error) {
	check, ok := checks[name]
	if !ok {
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

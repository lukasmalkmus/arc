package util

import "strings"

// MultiError is a collection of multiple errors. It implements the error
// interface.
type MultiError struct {
	errs []error
}

func (m MultiError) Error() string {
	errs := []string{}
	for _, err := range m.errs {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "\n")
}

// Add adds one or more errors.
func (m *MultiError) Add(es ...error) {
	for _, e := range es {
		if e != nil {
			m.errs = append(m.errs, e)
		}
	}
}

// Return returns the MultiError itself if errors are set, otherwise nil.
func (m *MultiError) Return() error {
	if len(m.errs) > 0 {
		return m
	}
	return nil
}

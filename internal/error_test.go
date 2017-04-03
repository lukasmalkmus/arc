package internal

import (
	"fmt"
	"testing"
)

func TestMultiError_Add(t *testing.T) {
	tests := []struct {
		errs []error
		len  int
	}{
		{[]error{fmt.Errorf("first error"), nil}, 1},
		{[]error{fmt.Errorf("first error"), fmt.Errorf("second error")}, 2},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, tt.len, len(me.errs))
	}
}

func TestMultiError_Error(t *testing.T) {
	tests := []struct {
		errs []error
		res  string
	}{
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
			res:  "first error\nsecond error",
		},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, tt.res, me.Error())
	}
}

func TestMultiError_Errors(t *testing.T) {
	tests := []struct {
		errs []error
		res  []error
	}{
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
			res:  []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
		},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, tt.res, me.Errors())
	}
}

func TestMultiError_Return(t *testing.T) {
	tests := []struct {
		errs []error
		err  error
	}{
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
			err:  MultiError{errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")}},
		},
		{
			errs: []error{},
			err:  nil,
		},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, tt.err, me.Return())
	}
}

func TestMultiError_Sort(t *testing.T) {
	tests := []struct {
		errs []error
		err  string
	}{
		{
			errs: []error{fmt.Errorf(`4:13: unresolved IDENTIFIER "y"`), fmt.Errorf(`3:8: unresolved IDENTIFIER "x"`)},
			err: `3:8: unresolved IDENTIFIER "x"
4:13: unresolved IDENTIFIER "y"`,
		},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		me.Sort()
		equals(t, tt.err, me.Error())
	}
}

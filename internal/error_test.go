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

func TestMultiError_Return(t *testing.T) {
	tests := []struct {
		errs []error
		err  error
	}{
		{errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
			err: &MultiError{errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")}},
		},
		{},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, tt.err, me.Return())
	}
}

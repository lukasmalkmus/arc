package util

import (
	"fmt"
	"testing"
)

func TestMultiError_Add(t *testing.T) {
	tests := []struct {
		errs []error
	}{
		{[]error{fmt.Errorf("first error"), fmt.Errorf("second error")}},
	}

	for _, tt := range tests {
		me := MultiError{}
		me.Add(tt.errs...)
		equals(t, len(tt.errs), len(me.errs))
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

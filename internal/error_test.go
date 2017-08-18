package internal

import (
	"errors"
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
		t.Run("", func(t *testing.T) {
			me := MultiError{}
			me.Add(tt.errs...)
			equals(t, tt.len, len(me.errs))
		})
	}
}

func TestMultiError_Error(t *testing.T) {
	tests := []struct {
		errs []error
		res  string
	}{
		{
			errs: []error{fmt.Errorf("first error")},
			res:  "first error",
		},
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
			res:  "first error\nsecond error",
		},
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error"), fmt.Errorf("third error")},
			res:  "first error\nsecond error\nthird error",
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			me := MultiError{}
			me.Add(tt.errs...)
			equals(t, tt.res, me.Error())
		})
	}
}

func TestMultiError_Errors(t *testing.T) {
	tests := []struct {
		errs []error
	}{
		{
			errs: []error{fmt.Errorf("first error"), fmt.Errorf("second error")},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			me := MultiError{}
			me.Add(tt.errs...)
			equals(t, tt.errs, me.Errors())
		})
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
		t.Run("", func(t *testing.T) {
			me := MultiError{}
			me.Add(tt.errs...)
			equals(t, tt.err, me.Return())
		})
	}
}

func TestMultiError_Sort(t *testing.T) {
	tests := []struct {
		errs []error
		err  string
	}{
		{
			errs: []error{fmt.Errorf(`4:13: unresolved IDENTIFIER "y"`), fmt.Errorf(`3:8: unresolved IDENTIFIER "x"`)},
			err:  "3:8: unresolved IDENTIFIER \"x\"\n4:13: unresolved IDENTIFIER \"y\"",
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			me := MultiError{}
			me.Add(tt.errs...)
			me.Sort()
			equals(t, tt.err, me.Error())
		})
	}
}

func BenchmarkMultiError_Sort(b *testing.B) {
	errs := []error{
		errors.New("4:13 first error"),
		errors.New("5:10 fifth error"),
		errors.New("1:11 ninth error"),
		errors.New("5:34 second error"),
		errors.New("56:23 sixth error"),
		errors.New("1:1 tenth error"),
		errors.New("9:34 thrid error"),
		errors.New("2:7 seventh error"),
		errors.New("8:3 fourth error"),
		errors.New("53:11 eigth error"),
	}
	for i := 0; i < b.N; i++ {
		me := MultiError{}
		me.Add(errs...)
		me.Sort()
	}
}

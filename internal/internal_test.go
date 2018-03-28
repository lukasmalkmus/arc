package internal

import (
	"reflect"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf("\033[31m "+msg+"\033[39m\n\n", v...)
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\033[31m unexpected error: %s\033[39m\n\n", err.Error())
	}
}

// equals fails the test if got is not equal to want.
func equals(tb testing.TB, got, want interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(got, want) {
		tb.Fatalf("\033[31m\n\n\tgot: %#v\n\n\twant: %#v\033[39m\n\n", got, want)
	}
}

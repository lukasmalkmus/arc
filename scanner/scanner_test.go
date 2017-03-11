package scanner

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/LukasMa/arc/token"
)

func TestScanner_Scan(t *testing.T) {
	tests := []struct {
		str     string
		Token   token.Token
		Literal string
	}{
		// Special tokens
		{"#", token.ILLEGAL, "#"},
		{"_", token.ILLEGAL, "_"},
		{"_123", token.ILLEGAL, "_"},
		{".", token.ILLEGAL, "."},
		{".123", token.ILLEGAL, "."},
		{"", token.EOF, ""},
		{" ", token.WS, " "},
		{"   ", token.WS, "   "},
		{"   x", token.WS, "   "},
		{"\t", token.WS, "\t"},
		{"\n", token.NL, "\n"},
		{"\r", token.NL, "\r"},
		{"\n\r", token.NL, "\n\r"},
		{"\r\n", token.NL, "\r\n"},
		{"\n\n", token.NL, "\n\n"},
		{"\r\r", token.NL, "\r\r"},
		{"\nx", token.NL, "\n"},
		{"!", token.COMMENT, "!"},
		{"! My comment", token.COMMENT, "! My comment"},
		{"!    My second comment", token.COMMENT, "!    My second comment"},

		// Identifiers
		{"x", token.IDENT, "x"},
		{"foo ", token.IDENT, "foo"},
		{"x9", token.IDENT, "x9"},
		{"r1", token.IDENT, "r1"},
		{"r10", token.IDENT, "r10"},
		{"r31", token.IDENT, "r31"},
		{"4", token.INT, "4"},
		{"8", token.INT, "8"},
		{"12", token.INT, "12"},
		{"16", token.INT, "16"},
		{"128", token.INT, "128"},
		{"123x", token.INT, "123"},

		// Operators
		{"+", token.PLUS, "+"},
		{"+4", token.PLUS, "+"},
		{"-", token.MINUS, "-"},
		{"-4", token.MINUS, "-"},

		// Misc characters
		{"[", token.LBRACKET, "["},
		{"]", token.RBRACKET, "]"},
		{",", token.COMMA, ","},
		{":", token.COLON, ":"},

		// Keywords
		{"ld", token.LOAD, "ld"},
		{"LD", token.LOAD, "LD"},
		{"st", token.STORE, "st"},
		{"ST", token.STORE, "ST"},
		{"add", token.ADD, "add"},
		{"ADD", token.ADD, "ADD"},
		{"sub", token.SUB, "sub"},
		{"SUB", token.SUB, "SUB"},

		// Directives
		{".begin", token.BEGIN, ".begin"},
		{".end", token.END, ".end"},
		{".org", token.ORG, ".org"},
	}

	for _, tt := range tests {
		s := New(strings.NewReader(tt.str))
		tok, lit := s.Scan()
		equals(t, tt.Token, tok)
		equals(t, tt.Literal, lit)
	}
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

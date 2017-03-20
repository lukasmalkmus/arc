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
		Line    int
	}{
		// Special tokens
		{"#", token.ILLEGAL, "#", 1},
		{"_", token.ILLEGAL, "_", 1},
		{"_123", token.ILLEGAL, "_", 1},
		{".", token.ILLEGAL, ".", 1},
		{".123", token.ILLEGAL, ".", 1},
		{"", token.EOF, "", 1},
		{" ", token.WS, " ", 1},
		{"   ", token.WS, "   ", 1},
		{"   x", token.WS, "   ", 1},
		{"\t", token.WS, "\t", 1},
		{"\n", token.NL, "\n", 1},         // Single newline (LF)
		{"\r\n", token.NL, "\n", 1},       // Single newline (CRLF)
		{"\n\n", token.NL, "\n\n", 2},     // Double newline (LF + LF)
		{"\r\n\r\n", token.NL, "\n\n", 2}, // Double newline (CRLF + CRLF)
		{"\nx", token.NL, "\n", 1},
		{"!", token.COMMENT, "!", 1},
		{"! My comment", token.COMMENT, "! My comment", 1},
		{"!    My second comment", token.COMMENT, "!    My second comment", 1},

		// Identifiers
		{"x", token.IDENT, "x", 1},
		{"foo ", token.IDENT, "foo", 1},
		{"x9", token.IDENT, "x9", 1},
		{"r1", token.IDENT, "r1", 1},
		{"r10", token.IDENT, "r10", 1},
		{"r31", token.IDENT, "r31", 1},
		{"4", token.INT, "4", 1},
		{"8", token.INT, "8", 1},
		{"12", token.INT, "12", 1},
		{"16", token.INT, "16", 1},
		{"128", token.INT, "128", 1},
		{"123x", token.INT, "123", 1},

		// Operators
		{"+", token.PLUS, "+", 1},
		{"+4", token.PLUS, "+", 1},
		{"-", token.MINUS, "-", 1},
		{"-4", token.MINUS, "-", 1},

		// Misc characters
		{"[", token.LBRACKET, "[", 1},
		{"]", token.RBRACKET, "]", 1},
		{",", token.COMMA, ",", 1},
		{":", token.COLON, ":", 1},

		// Keywords
		{"ld", token.LOAD, "ld", 1},
		{"LD", token.LOAD, "LD", 1},
		{"st", token.STORE, "st", 1},
		{"ST", token.STORE, "ST", 1},
		{"add", token.ADD, "add", 1},
		{"ADD", token.ADD, "ADD", 1},
		{"sub", token.SUB, "sub", 1},
		{"SUB", token.SUB, "SUB", 1},

		// Directives
		{".begin", token.BEGIN, ".begin", 1},
		{".end", token.END, ".end", 1},
		{".org", token.ORG, ".org", 1},
	}

	for tc, tt := range tests {
		s := New(strings.NewReader(tt.str))
		tok, lit, pos := s.Scan()
		equals(t, tc, tt.Token, tok)
		equals(t, tc, tt.Literal, lit)
		equals(t, tc, tt.Line, pos.Line)
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
func equals(tb testing.TB, tc int, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\n\n\t(test case %d)\033[39m\n\n", filepath.Base(file), line, exp, act, tc+1)
		tb.FailNow()
	}
}

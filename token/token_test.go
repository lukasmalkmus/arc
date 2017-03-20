package token

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestToken(t *testing.T) {
	tests := []struct {
		str    string
		tok    Token
		isSpec bool
		isLit  bool
		isOp   bool
		isKey  bool
		isDir  bool
	}{
		// Special tokens
		{"ILLEGAL", ILLEGAL, true, false, false, false, false},
		{"EOF", EOF, true, false, false, false, false},
		{"WHITESPACE", WS, true, false, false, false, false},
		{"NEWLINE", NL, true, false, false, false, false},
		{"COMMENT", COMMENT, true, false, false, false, false},

		// Identifiers and type literals
		{"IDENT", IDENT, false, true, false, false, false},
		{"INTEGER", INT, false, true, false, false, false},

		// Operators
		{"+", PLUS, false, false, true, false, false},
		{"-", MINUS, false, false, true, false, false},

		// Misc characters
		{"[", LBRACKET, false, false, false, false, false},
		{"]", RBRACKET, false, false, false, false, false},
		{",", COMMA, false, false, false, false, false},
		{":", COLON, false, false, false, false, false},

		// Keywords
		{"ld", LOAD, false, false, false, true, false},
		{"st", STORE, false, false, false, true, false},
		{"add", ADD, false, false, false, true, false},
		{"sub", SUB, false, false, false, true, false},

		// Directives
		{".begin", BEGIN, false, false, false, false, true},
		{".end", END, false, false, false, false, true},
		{".org", ORG, false, false, false, false, true},
	}

	for _, tt := range tests {
		equals(t, tt.tok.String(), tt.str)
		equals(t, tt.tok.IsSpecial(), tt.isSpec)
		equals(t, tt.tok.IsKeyword(), tt.isKey)
		equals(t, tt.tok.IsLiteral(), tt.isLit)
		equals(t, tt.tok.IsOperator(), tt.isOp)
		equals(t, tt.tok.IsDirective(), tt.isDir)
	}
}

func TestKeywords(t *testing.T) {
	for _, tok := range Keywords() {
		assert(t, tok.IsKeyword(), "Returned token isn't a keyword!", tok)
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		str   string
		isKey bool
		isDir bool
	}{
		// Identifiers
		{"abc", false, false},
		{"abc", false, false},

		// Keywords
		{"ld", true, false},
		{"st", true, false},
		{"add", true, false},
		{"sub", true, false},

		// Directives
		{".begin", false, true},
		{".end", false, true},
		{".org", false, true},
	}

	for _, tt := range tests {
		tok := Lookup(tt.str)
		equals(t, tt.isKey, tok.IsKeyword())
		equals(t, tt.isDir, tok.IsDirective())
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
		fmt.Printf("\033[31m%s:%d: unttected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if tt is not equal to act.
func equals(tb testing.TB, tt, act interface{}) {
	if !reflect.DeepEqual(tt, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, tt, act)
		tb.FailNow()
	}
}

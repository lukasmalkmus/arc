package token_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/lukasmalkmus/arc/token"
)

func TestToken(t *testing.T) {
	tests := []struct {
		str    string
		tok    token.Token
		isSpec bool
		isLit  bool
		isOp   bool
		isKey  bool
		isDir  bool
	}{
		// Special tokens
		{"ILLEGAL", token.ILLEGAL, true, false, false, false, false},
		{"EOF", token.EOF, true, false, false, false, false},
		{"WHITESPACE", token.WS, true, false, false, false, false},
		{"NEWLINE", token.NL, true, false, false, false, false},
		{"COMMENT", token.COMMENT, true, false, false, false, false},

		// Identifiers and type literals
		{"IDENTIFIER", token.IDENT, false, true, false, false, false},
		{"REGISTER", token.REG, false, true, false, false, false},
		{"INTEGER", token.INT, false, true, false, false, false},

		// Operators
		{"+", token.PLUS, false, false, true, false, false},
		{"-", token.MINUS, false, false, true, false, false},

		// Misc characters
		{"[", token.LBRACKET, false, false, false, false, false},
		{"]", token.RBRACKET, false, false, false, false, false},
		{",", token.COMMA, false, false, false, false, false},
		{":", token.COLON, false, false, false, false, false},

		// Keywords
		{"ld", token.LOAD, false, false, false, true, false},
		{"st", token.STORE, false, false, false, true, false},
		{"add", token.ADD, false, false, false, true, false},
		{"addcc", token.ADDCC, false, false, false, true, false},
		{"sub", token.SUB, false, false, false, true, false},
		{"subcc", token.SUBCC, false, false, false, true, false},
		{"and", token.AND, false, false, false, true, false},
		{"andcc", token.ANDCC, false, false, false, true, false},
		{"or", token.OR, false, false, false, true, false},
		{"orcc", token.ORCC, false, false, false, true, false},
		{"orn", token.ORN, false, false, false, true, false},
		{"orncc", token.ORNCC, false, false, false, true, false},
		{"xor", token.XOR, false, false, false, true, false},
		{"xorcc", token.XORCC, false, false, false, true, false},
		{"sll", token.SLL, false, false, false, true, false},
		{"sra", token.SRA, false, false, false, true, false},
		{"be", token.BE, false, false, false, true, false},
		{"bne", token.BNE, false, false, false, true, false},
		{"bneg", token.BNEG, false, false, false, true, false},
		{"bpos", token.BPOS, false, false, false, true, false},
		{"ba", token.BA, false, false, false, true, false},
		{"call", token.CALL, false, false, false, true, false},
		{"jmpl", token.JMPL, false, false, false, true, false},

		// Directives
		{".begin", token.BEGIN, false, false, false, false, true},
		{".end", token.END, false, false, false, false, true},
		{".org", token.ORG, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			t.Run("tokstr", func(t *testing.T) {
				equals(t, tt.tok.String(), tt.str)
			})
			t.Run("spec", func(t *testing.T) {
				equals(t, tt.tok.IsSpecial(), tt.isSpec)
			})
			t.Run("key", func(t *testing.T) {
				equals(t, tt.tok.IsKeyword(), tt.isKey)
			})
			t.Run("lit", func(t *testing.T) {
				equals(t, tt.tok.IsLiteral(), tt.isLit)
			})
			t.Run("op", func(t *testing.T) {
				equals(t, tt.tok.IsOperator(), tt.isOp)
			})
			t.Run("dir", func(t *testing.T) {
				equals(t, tt.tok.IsDirective(), tt.isDir)
			})
		})
	}
}

func TestDirectives(t *testing.T) {
	for _, tok := range token.Directives() {
		assert(t, tok.IsDirective(), "Returned token isn't a directive!", tok)
	}
}

func TestKeywords(t *testing.T) {
	for _, tok := range token.Keywords() {
		assert(t, tok.IsKeyword(), "Returned token isn't a keyword!", tok)
	}
}

// TestLookup makes sure that Lookup returns either the right keyword or IDENT
// for non keywords, like directives or identifiers.
func TestLookup(t *testing.T) {
	tests := []struct {
		str   string
		isKey bool
		isDir bool
	}{
		// Identifiers
		{"abc", false, false},
		{"123", false, false},
		{"%r1", false, false},

		// Keywords
		{"ld", true, false},
		{"st", true, false},
		{"add", true, false},
		{"addcc", true, false},
		{"sub", true, false},
		{"subcc", true, false},
		{"and", true, false},
		{"andcc", true, false},
		{"or", true, false},
		{"orcc", true, false},
		{"orn", true, false},
		{"orncc", true, false},
		{"xor", true, false},
		{"xorcc", true, false},
		{"sll", true, false},
		{"sra", true, false},
		{"be", true, false},
		{"bne", true, false},
		{"bneg", true, false},
		{"bpos", true, false},
		{"ba", true, false},
		{"call", true, false},
		{"jmpl", true, false},

		// Directives
		{".begin", false, true},
		{".end", false, true},
		{".org", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			tok := token.Lookup(tt.str)
			t.Run("key", func(t *testing.T) {
				equals(t, tt.isKey, tok.IsKeyword())
			})
			t.Run("dir", func(t *testing.T) {
				equals(t, tt.isDir, tok.IsDirective())
			})
		})
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

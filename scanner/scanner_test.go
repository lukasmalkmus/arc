package scanner

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/lukasmalkmus/arc/token"
)

func TestScanner_Scan(t *testing.T) {
	tests := []struct {
		str  string
		tok  token.Token
		lit  string
		line int
	}{
		// Special tokens
		{"#", token.ILLEGAL, "#", 1},
		{"_", token.ILLEGAL, "_", 1},
		{"_x", token.ILLEGAL, "_", 1},      // Underscore can't prefix identifier
		{"_123", token.ILLEGAL, "_", 1},    // Underscore can't prefix integer
		{"foo_", token.ILLEGAL, "foo_", 1}, // Underscore can't suffix identifier
		{".", token.ILLEGAL, ".", 1},
		{".x", token.ILLEGAL, ".x", 1},       // Dot can't prefix identifier, reserved for directive
		{".123", token.ILLEGAL, ".", 1},      // Dot can't prefix integer/integer can't suffix dot (reserved for directive)
		{"123x", token.ILLEGAL, "123x", 1},   // Illegal integer (wrong hex representation)
		{"08", token.ILLEGAL, "08", 1},       // Octal out of range
		{"0xx08", token.ILLEGAL, "0xx08", 1}, // Illegal hex syntax
		{"%", token.ILLEGAL, "%", 1},         // No ident after register char
		{"%%", token.ILLEGAL, "%", 1},        // No ident after register char
		{"%2", token.ILLEGAL, "%2", 1},       // First ident char is not a letter
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
		{"!", token.COMMENT, "!", 1}, // Empty comment
		{"! My comment", token.COMMENT, "! My comment", 1},
		{"!    My second comment   ", token.COMMENT, "!    My second comment   ", 1},

		// Identifiers
		{"x", token.IDENT, "x", 1},
		{"foo ", token.IDENT, "foo", 1},
		{"foo_bar", token.IDENT, "foo_bar", 1},
		{"r1", token.IDENT, "r1", 1},
		{"r10", token.IDENT, "r10", 1},
		{"r31", token.IDENT, "r31", 1},
		{"%r1", token.REG, "%r1", 1},
		{"%r10", token.REG, "%r10", 1},
		{"%r31", token.REG, "%r31", 1},

		// Integers
		{"4", token.INT, "4", 1},
		{"8", token.INT, "8", 1},
		{"12", token.INT, "12", 1},
		{"16", token.INT, "16", 1},
		{"128", token.INT, "128", 1},
		{"07", token.INT, "07", 1},     // Octal
		{"0x08", token.INT, "0x08", 1}, // Hex
		{"0X08", token.INT, "0x08", 1}, // X will get transformed to lower case

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
		{"addcc", token.ADDCC, "addcc", 1},
		{"ADDCC", token.ADDCC, "ADDCC", 1},
		{"sub", token.SUB, "sub", 1},
		{"SUB", token.SUB, "SUB", 1},
		{"subcc", token.SUBCC, "subcc", 1},
		{"SUBCC", token.SUBCC, "SUBCC", 1},
		{"and", token.AND, "and", 1},
		{"AND", token.AND, "AND", 1},
		{"andcc", token.ANDCC, "andcc", 1},
		{"ANDCC", token.ANDCC, "ANDCC", 1},
		{"or", token.OR, "or", 1},
		{"OR", token.OR, "OR", 1},
		{"orcc", token.ORCC, "orcc", 1},
		{"ORCC", token.ORCC, "ORCC", 1},
		{"orn", token.ORN, "orn", 1},
		{"ORN", token.ORN, "ORN", 1},
		{"orncc", token.ORNCC, "orncc", 1},
		{"ORNCC", token.ORNCC, "ORNCC", 1},
		{"xor", token.XOR, "xor", 1},
		{"XOR", token.XOR, "XOR", 1},
		{"xorcc", token.XORCC, "xorcc", 1},
		{"XORCC", token.XORCC, "XORCC", 1},
		{"sll", token.SLL, "sll", 1},
		{"SLL", token.SLL, "SLL", 1},
		{"sra", token.SRA, "sra", 1},
		{"SRA", token.SRA, "SRA", 1},
		{"be", token.BE, "be", 1},
		{"BE", token.BE, "BE", 1},
		{"bne", token.BNE, "bne", 1},
		{"BNE", token.BNE, "BNE", 1},
		{"bneg", token.BNEG, "bneg", 1},
		{"BNEG", token.BNEG, "BNEG", 1},
		{"bpos", token.BPOS, "bpos", 1},
		{"BPOS", token.BPOS, "BPOS", 1},
		{"bpos", token.BPOS, "bpos", 1},
		{"BPOS", token.BPOS, "BPOS", 1},
		{"ba", token.BA, "ba", 1},
		{"BA", token.BA, "BA", 1},
		{"call", token.CALL, "call", 1},
		{"CALL", token.CALL, "CALL", 1},
		{"jmpl", token.JMPL, "jmpl", 1},
		{"JMPL", token.JMPL, "JMPL", 1},

		// Directives
		{".begin", token.BEGIN, ".begin", 1},
		{".end", token.END, ".end", 1},
		{".org", token.ORG, ".org", 1},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			s := New(strings.NewReader(tt.str))
			tok, lit, pos := s.Scan()
			t.Run("tok", func(t *testing.T) {
				equals(t, tt.tok.String(), tok.String())
			})
			t.Run("lit", func(t *testing.T) {
				equals(t, tt.lit, lit)
			})
			t.Run("pos", func(t *testing.T) {
				equals(t, tt.line, pos.Line)
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

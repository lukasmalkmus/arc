package parser

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/LukasMa/arc/ast"
)

var errExp = errors.New("Expecting error")

func TestParse(t *testing.T) {
	tests := []struct {
		prog string
	}{
		{"ld %r1, %r2"},
		{"ld %r1, %r2\nld %r2, %r3"},
		{"\nld %r1, %r2\nld %r2, %r3"},
	}

	for _, tt := range tests {
		_, err := Parse(tt.prog)
		ok(t, err)
	}
}

func TestParseLoadStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "ld %r1, %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.MemoryLocation{Base: &ast.Identifier{Name: "%r1"}, Mode: ast.Indirect},
				Destination: &ast.Identifier{Name: "%r2"},
			},
		},
		{
			str: "ld [x], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.MemoryLocation{Base: &ast.Identifier{Name: "x"}, Mode: ast.Direct},
				Destination: &ast.Identifier{Name: "%r2"},
			},
		},
		{
			str: "ld [%r1+8191], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.MemoryLocation{Base: &ast.Identifier{Name: "%r1"}, Operator: "+", Offset: 8191, Mode: ast.Offset},
				Destination: &ast.Identifier{Name: "%r2"},
			},
		},
		{
			str:  "l %r1, %r2",
			stmt: nil,
			err:  `found IDENT ("l"), expected "ld", "st", "add", "sub"`,
		},
		{
			str:  "ld ld, %r2",
			stmt: nil,
			err:  `found "ld", expected IDENT`,
		},
		{
			str:  "ld %r1 %r2",
			stmt: nil,
			err:  `found IDENT ("%r2"), expected ","`,
		},
		{
			str:  "ld %r1, ld",
			stmt: nil,
			err:  `found "ld", expected IDENT`,
		},
		{
			str:  "ld %r1, %r2, %r3",
			stmt: nil,
			err:  `found ",", expected NEWLINE, EOF`,
		},
		{
			str:  "\nld %r1, %r2",
			stmt: nil,
			err:  `found NEWLINE, expected "ld", "st", "add", "sub"`,
		},
		{
			str:  "ld [%r1*4], %r2",
			stmt: nil,
			err:  `found ILLEGAL ("*"), expected "+", "-"`,
		},
		{
			str:  "ld %r1+4], %r2",
			stmt: nil,
			err:  `found "+", expected ","`,
		},
		{
			str:  "ld [%r1 + 4, %r2",
			stmt: nil,
			err:  `found ",", expected "]"`,
		},
		{
			str:  "ld %r1, [%r2]",
			stmt: nil,
			err:  `found "[", expected IDENT`,
		},
		{
			str:  "ld [%r1+x], %r2",
			stmt: nil,
			err:  `found IDENT ("x"), expected INT`,
		},
		{
			str:  "ld [%r1+10000], %r2",
			stmt: nil,
			err:  `found INT "10000" is not a valid SIMM13`,
		},
		{
			str:  "ld [10000], %r2",
			stmt: nil,
			err:  `found INT ("10000"), expected IDENT`,
		},
	}

	for _, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if loadStmt, valid := tt.stmt.(*ast.LoadStatement); valid {
			ok(t, err)
			fmt.Println(stmt)
			equals(t, loadStmt, stmt)
		} else {
			equals(t, tt.err, err.Error())
		}
	}
}

// Test if the correct token is returned after an unscan().
func TestParserBuffer(t *testing.T) {
	test := `ld %r1, %r2`
	p := New(strings.NewReader(test))
	p.scan()
	tok, lit := p.tok, p.lit
	p.unscan()
	bufTok, bufLit := p.buf.tok, p.buf.lit
	equals(t, tok, bufTok)
	equals(t, lit, bufLit)
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

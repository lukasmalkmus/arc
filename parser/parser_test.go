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

// Test if the correct token is returned after an unscan().
func TestParserBuffer(t *testing.T) {
	test := `ld %r1, %r2`
	p := New(strings.NewReader(test))
	p.scan()
	tok, lit := p.tok, p.lit
	p.unscan()
	bufTok, bufLit := p.buf.tok, p.buf.lit
	equals(t, 0, tok, bufTok)
	equals(t, 0, lit, bufLit)
}

// TestParse will validate that linebreaks, etc. don't break the parser.
func TestParse(t *testing.T) {
	tests := []struct {
		prog string
	}{
		{"ld %r1, %r2"},
		{"ld %r1, %r2\nld %r2, %r3"},
		{"\nld %r1, %r2\nld %r2, %r3"},
		{"\nld %r1, %r2\n\n\nld %r2, %r3"},
	}

	for tc, tt := range tests {
		_, err := Parse(tt.prog)
		ok(t, tc, err)
	}
}

// TestParseLoadStatement validates the correct parsing of ld commands.
func TestParseLoadStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "ld %r1, %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Identifier{Value: "%r1"},
				Destination: &ast.Identifier{Value: "%r2"},
			},
		},
		{
			str: "ld [x], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Ident: &ast.Identifier{Value: "x"}},
				Destination: &ast.Identifier{Value: "%r2"},
			},
		},
		{
			str: "ld [%r1+8191], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 8191},
				Destination: &ast.Identifier{Value: "%r2"},
			},
		},
		{
			str: "ld [%r1+0], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 0},
				Destination: &ast.Identifier{Value: "%r2"},
			},
		},
		{
			str: "l %r1, %r2",
			err: `found IDENT ("%r1"), expected ":"`,
		},
		{
			str: "ld ld, %r2",
			err: `found "ld", expected "[", IDENT`,
		},
		{
			str: "ld %r1 %r2",
			err: `found IDENT ("%r2"), expected ","`,
		},
		{
			str: "ld %r1, ld",
			err: `found "ld", expected IDENT`,
		},
		{
			str: "ld %r1, %r2, %r3",
			err: `found ",", expected NEWLINE, EOF`,
		},
		{
			str: "\nld %r1, %r2",
			err: `found NEWLINE, expected "ld", "st", "add", "sub"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if loadStmt, valid := tt.stmt.(*ast.LoadStatement); valid {
			ok(t, tc, err)
			equals(t, tc, loadStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseStoreStatement validates the correct parsing of st commands.
func TestParseStoreStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "st %r2, %r1",
			stmt: &ast.StoreStatement{
				Source:      &ast.Identifier{Value: "%r2"},
				Destination: &ast.Identifier{Value: "%r1"},
			},
		},
		{
			str: "st %r2, [x]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Identifier{Value: "%r2"},
				Destination: &ast.Expression{Ident: &ast.Identifier{Value: "x"}},
			},
		},
		{
			str: "st %r2, [%r1+8191]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Identifier{Value: "%r2"},
				Destination: &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 8191},
			},
		},
		{
			str: "st %r2, [%r1+0]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Identifier{Value: "%r2"},
				Destination: &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 0},
			},
		},
		{
			str: "s %r2, %r1",
			err: `found IDENT ("%r2"), expected ":"`,
		},
		{
			str: "st st, %r1",
			err: `found "st", expected IDENT`,
		},
		{
			str: "st %r2 %r1",
			err: `found IDENT ("%r1"), expected ","`,
		},
		{
			str: "st %r2, st",
			err: `found "st", expected "[", IDENT`,
		},
		{
			str: "st %r2, %r1, %r3",
			err: `found ",", expected NEWLINE, EOF`,
		},
		{
			str: "\nst %r2, %r1",
			err: `found NEWLINE, expected "ld", "st", "add", "sub"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if loadStmt, valid := tt.stmt.(*ast.StoreStatement); valid {
			ok(t, tc, err)
			equals(t, tc, loadStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseLabelStatement validates the correct parsing of st commands.
func TestParseLabelStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str:  "x: 25",
			stmt: &ast.LabelStatement{Ident: &ast.Identifier{Value: "x"}, Reference: ast.Integer(25)},
		},
		{
			str: "mylabel: ld %r1, %r2",
			stmt: &ast.LabelStatement{Ident: &ast.Identifier{Value: "mylabel"},
				Reference: &ast.LoadStatement{
					Source:      &ast.Identifier{Value: "%r1"},
					Destination: &ast.Identifier{Value: "%r2"},
				}},
		},
		{str: "x: y: 25", err: `found IDENT ("y"), expected INTEGER, "ld", "st", "add", "sub"`},
		{str: "x: 25;", err: `found ILLEGAL (";"), expected NEWLINE, EOF`},
		{str: "x: ld", err: `found EOF, expected "[", IDENT`},
		{str: "X: 90000000000000", err: `integer 90000000000000 overflows 32 bit integer`},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if labelStmt, valid := tt.stmt.(*ast.LabelStatement); valid {
			ok(t, tc, err)
			equals(t, tc, labelStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseExpression verifies the correct parsing of expressions.
func TestParseExpression(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Expression
		err string
	}{
		{str: "[%r1+8191]", obj: &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 8191}},
		{str: "[%r1+0]", obj: &ast.Expression{Ident: &ast.Identifier{Value: "%r1"}, Operator: "+", Offset: 0}},
		{str: "[x]", obj: &ast.Expression{Ident: &ast.Identifier{Value: "x"}, Operator: "", Offset: 0}},
		{str: "x]", err: `found IDENT ("x"), expected "["`},
		{str: "[+8191]", err: `found "+", expected IDENT`},
		{str: "[0+8191]", err: `found INTEGER ("0"), expected IDENT`},
		{str: "[%r1 8191]", err: `found INTEGER ("8191"), expected "+", "-", "]"`},
		{str: "[%r1*8191]", err: `found ILLEGAL ("*"), expected "+", "-", "]"`},
		{str: "[%r1+]", err: `found "]", expected INTEGER`},
		{str: "[%r1+45", err: `found EOF, expected "]"`},
	}

	for i, tt := range tests {
		exp, err := New(strings.NewReader(tt.str)).parseExpression()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, exp)
		} else {
			equals(t, i, tt.err, err.Error())
		}
	}
}

// TestParseIdent verifies the correct parsing of identifiers.
func TestParseIdent(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Identifier
		err string
	}{
		{str: "x", obj: &ast.Identifier{Value: "x"}},
		{str: "%r1", obj: &ast.Identifier{Value: "%r1"}},
		{str: "mylabel", obj: &ast.Identifier{Value: "mylabel"}},
		{str: ":x", err: `found ":", expected IDENT`},
		{str: "123", err: `found INTEGER ("123"), expected IDENT`},
	}

	for i, tt := range tests {
		ident, err := New(strings.NewReader(tt.str)).parseIdent()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, ident)
		} else {
			equals(t, i, tt.err, err.Error())
		}
	}
}

// TestParseInteger verifies the correct parsing of integers.
func TestParseInteger(t *testing.T) {
	tests := []struct {
		str string
		obj ast.Integer
		err string
	}{
		{str: "0", obj: ast.Integer(0)},
		{str: "100", obj: ast.Integer(100)},
		{str: "001", obj: ast.Integer(1)},
		{str: "0", obj: ast.Integer(0)},
		{str: "90000000000000", err: `integer 90000000000000 overflows 32 bit integer`},
		{str: "x", err: `found IDENT ("x"), expected INTEGER`},
	}

	for i, tt := range tests {
		integer, err := New(strings.NewReader(tt.str)).parseInteger()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, integer)
		} else {
			equals(t, i, tt.err, err.Error())
		}
	}
}

// TestParseMemoryLocation verifies the correct parsing of memory locations.
func TestParseMemoryLocation(t *testing.T) {
	tests := []struct {
		str string
		obj ast.MemoryLocation
		err string
	}{
		{str: "[x]", obj: &ast.Expression{Ident: &ast.Identifier{Value: "x"}}},
		{str: "x", obj: &ast.Identifier{Value: "x"}},
		{str: "123", err: `found INTEGER ("123"), expected "[", IDENT`},
	}

	for i, tt := range tests {
		loc, err := New(strings.NewReader(tt.str)).parseMemoryLocation()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, loc)
		} else {
			equals(t, i, tt.err, err.Error())
		}
	}
}

// TestExpectStatementEnd verifies the correct detection of statement ends.
func TestExpectStatementEnd(t *testing.T) {
	tests := []struct {
		str string
		err string
	}{
		{str: "\r\n"},
		{str: "\n\r"},
		{str: "\n\n"},
		{str: ""},
		{str: " "},
		{str: string(rune(0))},
		{str: "\t", err: `found ILLEGAL, expected NEWLINE, EOF`},
		{str: ";", err: `found ILLEGAL (";"), expected NEWLINE, EOF`},
	}

	for i, tt := range tests {
		err := New(strings.NewReader(tt.str)).expectStatementEnd()
		if err == nil {
			ok(t, i, err)
		} else {
			equals(t, i, tt.err, err.Error())
		}
	}
}

// TestParseSIMM13 verifies the correct parsing of integers.
func TestParseSIMM13(t *testing.T) {
	tests := []struct {
		str string
		obj ast.Integer
		err string
	}{
		{str: "0", obj: ast.Integer(0)},
		{str: "100", obj: ast.Integer(100)},
		{str: "001", obj: ast.Integer(1)},
		{str: "0", obj: ast.Integer(0)},
		{str: "8191", obj: ast.Integer(8191)},
		{str: "8192", err: `integer 8192 is not a valid SIMM13`},
		{str: "-1", err: `found "-", expected INTEGER`},
	}

	for i, tt := range tests {
		integer, err := New(strings.NewReader(tt.str)).parseSIMM13()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, integer)
		} else {
			equals(t, i, tt.err, err.Error())
		}
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
func ok(tb testing.TB, tc int, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error in test case %d: %s\033[39m\n\n", filepath.Base(file), line, tc+1, err.Error())
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

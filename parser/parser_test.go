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

// TestParserBuffer tests if the correct token is returned after an unscan().
func TestParserBuffer(t *testing.T) {
	test := `ld %r1, %r2`
	p := New(strings.NewReader(test))
	p.scan()
	tok, lit, pos := p.tok, p.lit, p.pos
	p.unscan()
	bufTok, bufLit, bufPos := p.buf.tok, p.buf.lit, p.buf.pos
	equals(t, 0, tok, bufTok)
	equals(t, 0, lit, bufLit)
	equals(t, 0, pos, bufPos)
}

// TestParse will validate the correct parsing of a complete program.
func TestParse(t *testing.T) {
	tests := []struct {
		prog string
		err  string
	}{
		{
			prog: `.begin
		ld %r1, %r2
		.end`,
			err: ``,
		},
		{
			prog: `
		.begin
		! A comment above the statement
		ld %r1, %r2

		st %r2, %r3 ! A comment behind the statement


		! Another comment above the comment
		ld %r3, %r4 ! Another comment behind the statement

		.end

		! This is valid
		st %r4, %r5

		`,
			err: ``,
		},
		{
			prog: `.begin
		ld %r1 %r2
		.end`,
			err: `line 2: found REGISTER ("r2"), expected ","`,
		},
	}

	for tc, tt := range tests {
		prog, err := Parse(tt.prog)
		if prog != nil {
			ok(t, tc, err)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseCommentStatement validates the correct parsing of the begin directive.
func TestParseCommentStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: "!  This is a comment  ", stmt: &ast.CommentStatement{Text: "This is a comment"}},
		{str: "This is not a comment", err: `line 1: found IDENT ("is"), expected ":"`},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if commentStmt, valid := tt.stmt.(*ast.CommentStatement); valid {
			ok(t, tc, err)
			equals(t, tc, commentStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseBeginStatement validates the correct parsing of the begin directive.
func TestParseBeginStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".begin", stmt: &ast.BeginStatement{}},
		{str: ".beg", err: `line 1: found ILLEGAL (".beg"), expected COMMENT, IDENT, ".begin", ".end", ".org", "ld", "st", "add", "sub"`},
		{str: "begin", err: `line 1: found EOF, expected ":"`},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if beginStmt, valid := tt.stmt.(*ast.BeginStatement); valid {
			ok(t, tc, err)
			equals(t, tc, beginStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseEndStatement validates the correct parsing of the end directive.
func TestParseEndStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".end", stmt: &ast.EndStatement{}},
		{str: ".ed", err: `line 1: found ILLEGAL (".ed"), expected COMMENT, IDENT, ".begin", ".end", ".org", "ld", "st", "add", "sub"`},
		{str: "end", err: `line 1: found EOF, expected ":"`},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if endStmt, valid := tt.stmt.(*ast.EndStatement); valid {
			ok(t, tc, err)
			equals(t, tc, endStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseOrgStatement validates the correct parsing of the org directive.
func TestParseOrgStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".org 2048", stmt: &ast.OrgStatement{Value: ast.Integer(2048)}},
		{str: ".org 2048 128", err: `line 1: found INTEGER ("128"), expected COMMENT, NEWLINE, EOF`},
		{str: ".org", err: `line 1: found EOF, expected INTEGER`},
		{str: ".og", err: `line 1: found ILLEGAL (".og"), expected COMMENT, IDENT, ".begin", ".end", ".org", "ld", "st", "add", "sub"`},
		{str: "org", err: `line 1: found EOF, expected ":"`},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if orgStmt, valid := tt.stmt.(*ast.OrgStatement); valid {
			ok(t, tc, err)
			equals(t, tc, orgStmt, stmt)
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
			stmt: &ast.LabelStatement{Ident: &ast.Identifier{Name: "x"}, Reference: ast.Integer(25)},
		},
		{
			str: "mylabel: ld %r1, %r2",
			stmt: &ast.LabelStatement{Ident: &ast.Identifier{Name: "mylabel"},
				Reference: &ast.LoadStatement{
					Source:      &ast.Register{Name: "r1"},
					Destination: &ast.Register{Name: "r2"},
				}},
		},
		{str: "x: y: 25", err: `line 1: found IDENT ("y"), expected INTEGER, "ld", "st", "add", "sub"`},
		{str: "x: 25;", err: `line 1: found ILLEGAL (";"), expected COMMENT, NEWLINE, EOF`},
		{str: "x: ld", err: `line 1: found EOF, expected "[", REGISTER`},
		{str: "X: 90000000000000", err: `line 1: integer 90000000000000 overflows 32 bit integer`},
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
				Source:      &ast.Register{Name: "r1"},
				Destination: &ast.Register{Name: "r2"},
			},
		},
		{
			str: "ld [x], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Base: &ast.Identifier{Name: "x"}},
				Destination: &ast.Register{Name: "r2"},
			},
		},
		{
			str: "ld [%r1+8191], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 8191},
				Destination: &ast.Register{Name: "r2"},
			},
		},
		{
			str: "ld [%r1+0], %r2",
			stmt: &ast.LoadStatement{
				Source:      &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 0},
				Destination: &ast.Register{Name: "r2"},
			},
		},
		{
			str: "l %r1, %r2",
			err: `line 1: found REGISTER ("r1"), expected ":"`,
		},
		{
			str: "ld ld, %r2",
			err: `line 1: found "ld", expected "[", REGISTER`,
		},
		{
			str: "ld %r1 %r2",
			err: `line 1: found REGISTER ("r2"), expected ","`,
		},
		{
			str: "ld %r1, ld",
			err: `line 1: found "ld", expected REGISTER`,
		},
		{
			str: "ld %r1, %r2, %r3",
			err: `line 1: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nld %r1, %r2",
			err: `line 1: found NEWLINE, expected COMMENT, IDENT, ".begin", ".end", ".org", "ld", "st", "add", "sub"`,
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
				Source:      &ast.Register{Name: "r2"},
				Destination: &ast.Register{Name: "r1"},
			},
		},
		{
			str: "st %r2, [x]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Register{Name: "r2"},
				Destination: &ast.Expression{Base: &ast.Identifier{Name: "x"}},
			},
		},
		{
			str: "st %r2, [%r1+8191]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Register{Name: "r2"},
				Destination: &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 8191},
			},
		},
		{
			str: "st %r2, [%r1+0]",
			stmt: &ast.StoreStatement{
				Source:      &ast.Register{Name: "r2"},
				Destination: &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 0},
			},
		},
		{
			str: "s %r2, %r1",
			err: `line 1: found REGISTER ("r2"), expected ":"`,
		},
		{
			str: "st st, %r1",
			err: `line 1: found "st", expected REGISTER`,
		},
		{
			str: "st %r2 %r1",
			err: `line 1: found REGISTER ("r1"), expected ","`,
		},
		{
			str: "st %r2, st",
			err: `line 1: found "st", expected "[", REGISTER`,
		},
		{
			str: "st %r2, %r1, %r3",
			err: `line 1: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nst %r2, %r1",
			err: `line 1: found NEWLINE, expected COMMENT, IDENT, ".begin", ".end", ".org", "ld", "st", "add", "sub"`,
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

// TestParseExpression verifies the correct parsing of expressions.
func TestParseExpression(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Expression
		err string
	}{
		{str: "[%r1+8191]", obj: &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 8191}},
		{str: "[%r1+0]", obj: &ast.Expression{Base: &ast.Register{Name: "r1"}, Operator: "+", Offset: 0}},
		{str: "[x]", obj: &ast.Expression{Base: &ast.Identifier{Name: "x"}, Operator: "", Offset: 0}},
		{str: "x]", err: `line 1: found IDENT ("x"), expected "["`},
		{str: "[+8191]", err: `line 1: found "+", expected IDENT, REGISTER`},
		{str: "[0+8191]", err: `line 1: found INTEGER ("0"), expected IDENT, REGISTER`},
		{str: "[%r1 8191]", err: `line 1: found INTEGER ("8191"), expected "+", "-", "]"`},
		{str: "[%r1*8191]", err: `line 1: found ILLEGAL ("*"), expected "+", "-", "]"`},
		{str: "[%r1+]", err: `line 1: found "]", expected INTEGER`},
		{str: "[%r1+45", err: `line 1: found EOF, expected "]"`},
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
		{str: "x", obj: &ast.Identifier{Name: "x"}},
		{str: "mylabel", obj: &ast.Identifier{Name: "mylabel"}},
		{str: ":x", err: `line 1: found ":", expected IDENT`},
		{str: "123", err: `line 1: found INTEGER ("123"), expected IDENT`},
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
		{str: "90000000000000", err: `line 1: integer 90000000000000 overflows 32 bit integer`},
		{str: "x", err: `line 1: found IDENT ("x"), expected INTEGER`},
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
		{str: "[x]", obj: &ast.Expression{Base: &ast.Identifier{Name: "x"}}},
		{str: "%r1", obj: &ast.Register{Name: "r1"}},
		{str: "x", err: `line 1: found IDENT ("x"), expected "[", REGISTER`},
		{str: "123", err: `line 1: found INTEGER ("123"), expected "[", REGISTER`},
		{str: "[x+]", err: `line 1: found "]", expected INTEGER`},
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
		{str: "\t", err: `line 1: found ILLEGAL, expected NEWLINE, EOF`},
		{str: ";", err: `line 1: found ILLEGAL (";"), expected NEWLINE, EOF`},
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
		{str: "8192", err: `line 1: integer 8192 is not a valid SIMM13`},
		{str: "-1", err: `line 1: found "-", expected INTEGER`},
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

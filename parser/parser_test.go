package parser

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/lukasmalkmus/arc/ast"
	"github.com/lukasmalkmus/arc/token"
)

var errExp = errors.New("Expecting error")
var testPos = token.Pos{Line: 1, Char: 1}

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
		},
		{
			prog: `.begin
		ld %r1 %r2
		.end`,
			err: `2:10: found REGISTER "%r2", expected ","`,
		},
		{
			prog: `.begin
		.org 2048
		ld ld, %r2
		st %r2, %r3
		.org 3000
		x: 25
		y: x: z
		ld %r3, %r4
		.end`,
			err: `3:6: found "ld", expected "[", REGISTER
7:6: found IDENTIFIER "x", expected INTEGER, "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
		{
			prog: `
			x: 25
			x: ld %r1, %r2
			x: st %r2, %r3`,
			err: `3:4: label "x" already declared: previous declaration at 2:4
4:4: label "x" already declared: previous declaration at 2:4`,
		},
		{
			prog: `
			.begin
			ld [x], %r1
			st %r1, [y]
			.end`,
			err: `3:8: unresolved IDENTIFIER "x"
4:13: unresolved IDENTIFIER "y"`,
		},
	}

	for tc, tt := range tests {
		_, err := Parse(tt.prog)
		if tt.err == "" {
			ok(t, tc, err)
		} else {
			if err == nil {
				t.Fatalf("Expected error but got nil!\n(test case %d)", tc)
			}
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
		{str: "!  This is a comment  ", stmt: &ast.CommentStatement{Token: 5, Position: testPos, Text: "!  This is a comment  "}},
		{str: "This is not a comment", err: `1:6: found IDENTIFIER "is", expected ":"`},
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
		{str: ".begin", stmt: &ast.BeginStatement{Token: 39, Position: testPos}},
		{str: ".beg", err: `1:1: found ILLEGAL ".beg", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`},
		{str: "begin", err: `1:6: found EOF, expected ":"`},
		{str: ".begin 123", err: `1:8: found INTEGER "123", expected COMMENT, NEWLINE, EOF`},
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
		{str: ".end", stmt: &ast.EndStatement{Token: 40, Position: testPos}},
		{str: ".ed", err: `1:1: found ILLEGAL ".ed", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`},
		{str: "end", err: `1:4: found EOF, expected ":"`},
		{str: ".end 123", err: `1:6: found INTEGER "123", expected COMMENT, NEWLINE, EOF`},
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
		{str: ".org 2048", stmt: &ast.OrgStatement{Token: 41, Position: testPos, Value: &ast.Integer{Token: 10, Position: posAfter(6), Value: 2048, Literal: "2048"}}},
		{str: ".org 2048 128", err: `1:11: found INTEGER "128", expected COMMENT, NEWLINE, EOF`},
		{str: ".org", err: `1:5: found EOF, expected INTEGER`},
		{str: ".og", err: `1:1: found ILLEGAL ".og", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`},
		{str: "org", err: `1:4: found EOF, expected ":"`},
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
			str: "x: 25",
			stmt: &ast.LabelStatement{
				Token:     8,
				Position:  testPos,
				Ident:     &ast.Identifier{Token: 8, Position: testPos, Name: "x"},
				Reference: &ast.Integer{Token: 10, Position: posAfter(4), Value: 25, Literal: "25"},
			},
		},
		{
			str: "mylabel: ld %r1, %r2",
			stmt: &ast.LabelStatement{
				Token:    8,
				Position: testPos,
				Ident:    &ast.Identifier{Token: 8, Position: testPos, Name: "mylabel"},
				Reference: &ast.LoadStatement{
					Token:       21,
					Position:    posAfter(10),
					Source:      &ast.Register{Name: "%r1"},
					Destination: &ast.Register{Name: "%r2"},
				},
			},
		},
		{str: "x: y: 25", err: `1:4: found IDENTIFIER "y", expected INTEGER, "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`},
		{str: "x: 25;", err: `1:6: found ILLEGAL ";", expected COMMENT, NEWLINE, EOF`},
		{str: "x: ld", err: `1:6: found EOF, expected "[", REGISTER`},
		{str: "X: 90000000000000", err: `1:4: INTEGER "90000000000000" out of 32 bit range`},
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

// TestParseLoadStatement validates the correct parsing of load commands.
func TestParseLoadStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "ld %r1, %r2",
			stmt: &ast.LoadStatement{
				Token:       21,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "ld [x], %r2",
			stmt: &ast.LoadStatement{
				Token:    21,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base: &ast.Identifier{Token: 8,
						Position: posAfter(5),
						Name:     "x",
					},
				},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "ld [%r1+8191], %r2",
			stmt: &ast.LoadStatement{
				Token:    21,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: 10, Position: posAfter(9), Value: 8191, Literal: "8191"},
				},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "ld [%r1+0], %r2",
			stmt: &ast.LoadStatement{
				Token:    21,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: 10, Position: posAfter(9), Value: 0, Literal: "0"},
				},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "l %r1, %r2",
			err: `1:3: found REGISTER "%r1", expected ":"`,
		},
		{
			str: "ld ld, %r2",
			err: `1:4: found "ld", expected "[", REGISTER`,
		},
		{
			str: "ld %r1 %r2",
			err: `1:8: found REGISTER "%r2", expected ","`,
		},
		{
			str: "ld %r1, ld",
			err: `1:9: found "ld", expected REGISTER`,
		},
		{
			str: "ld %r1, %r2, %r3",
			err: `1:12: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nld %r1, %r2",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
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

// TestParseStoreStatement validates the correct parsing of store commands.
func TestParseStoreStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "st %r2, %r1",
			stmt: &ast.StoreStatement{
				Token:       22,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r1"},
			},
		},
		{
			str: "st %r2, [x]",
			stmt: &ast.StoreStatement{
				Token:    22,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base: &ast.Identifier{Token: 8,
						Position: posAfter(10),
						Name:     "x",
					},
				},
			},
		},
		{
			str: "st %r2, [%r1+8191]",
			stmt: &ast.StoreStatement{
				Token:    22,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: 10, Position: posAfter(14), Value: 8191, Literal: "8191"},
				},
			},
		},
		{
			str: "st %r2, [%r1+0]",
			stmt: &ast.StoreStatement{
				Token:    22,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: 10, Position: posAfter(14), Value: 0, Literal: "0"},
				},
			},
		},
		{
			str: "s %r2, %r1",
			err: `1:3: found REGISTER "%r2", expected ":"`,
		},
		{
			str: "st st, %r1",
			err: `1:4: found "st", expected REGISTER`,
		},
		{
			str: "st %r2 %r1",
			err: `1:8: found REGISTER "%r1", expected ","`,
		},
		{
			str: "st %r2, st",
			err: `1:9: found "st", expected "[", REGISTER`,
		},
		{
			str: "st %r2, %r1, %r3",
			err: `1:12: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nst %r2, %r1",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if storeStmt, valid := tt.stmt.(*ast.StoreStatement); valid {
			ok(t, tc, err)
			equals(t, tc, storeStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseAddStatement validates the correct parsing of add commands.
func TestParseAddStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "add %r1, %r2, %r3",
			stmt: &ast.AddStatement{
				Token:       23,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "add %r1, 32, %r3",
			stmt: &ast.AddStatement{
				Token:       23,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "add %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "add %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "add %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "add %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "add x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "add %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "add 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "and %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nadd %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if addStmt, valid := tt.stmt.(*ast.AddStatement); valid {
			ok(t, tc, err)
			equals(t, tc, addStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseAddCCStatement validates the correct parsing of addcc commands.
func TestParseAddCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "addcc %r1, %r2, %r3",
			stmt: &ast.AddCCStatement{
				Token:       24,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "addcc %r1, 32, %r3",
			stmt: &ast.AddCCStatement{
				Token:       24,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(12), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "addcc %r1 %r2, %r3",
			err: `1:11: found REGISTER "%r2", expected ","`,
		},
		{
			str: "addcc %r1, %r2",
			err: `1:15: found EOF, expected ","`,
		},
		{
			str: "addcc %r1, %r2, 32",
			err: `1:17: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "addcc %r1, x, %r3",
			err: `1:12: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "addcc x, %r2, %r3",
			err: `1:7: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "addcc %r1, %r2, %r3, %r4",
			err: `1:20: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "addcc 32, %r2, %r3",
			err: `1:7: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "andcc %r1, 90000000000, %r3",
			err: `1:12: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\naddcc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if addCCStmt, valid := tt.stmt.(*ast.AddCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, addCCStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseSubStatement validates the correct parsing of sub commands.
func TestParseSubStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sub %r1, %r2, %r3",
			stmt: &ast.SubStatement{
				Token:       25,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sub %r1, 32, %r3",
			stmt: &ast.SubStatement{
				Token:       25,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sub %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "sub %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "sub %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sub %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "sub x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "sub %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "sub 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sub %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nsub %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if subStmt, valid := tt.stmt.(*ast.SubStatement); valid {
			ok(t, tc, err)
			equals(t, tc, subStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseSubCCStatement validates the correct parsing of subcc commands.
func TestParseSubCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "subcc %r1, %r2, %r3",
			stmt: &ast.SubCCStatement{
				Token:       26,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "subcc %r1, 32, %r3",
			stmt: &ast.SubCCStatement{
				Token:       26,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(12), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "subcc %r1 %r2, %r3",
			err: `1:11: found REGISTER "%r2", expected ","`,
		},
		{
			str: "subcc %r1, %r2",
			err: `1:15: found EOF, expected ","`,
		},
		{
			str: "subcc %r1, %r2, 32",
			err: `1:17: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "subcc %r1, x, %r3",
			err: `1:12: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "subcc x, %r2, %r3",
			err: `1:7: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "subcc %r1, %r2, %r3, %r4",
			err: `1:20: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "subcc 32, %r2, %r3",
			err: `1:7: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "subcc %r1, 90000000000, %r3",
			err: `1:12: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nsubcc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if subCCStmt, valid := tt.stmt.(*ast.SubCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, subCCStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseAndStatement validates the correct parsing of and commands.
func TestParseAndStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "and %r1, %r2, %r3",
			stmt: &ast.AndStatement{
				Token:       27,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "and %r1, 32, %r3",
			stmt: &ast.AndStatement{
				Token:       27,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "and %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "and %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "and %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "and %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "and x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "and %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "and 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "and %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nand %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if andStmt, valid := tt.stmt.(*ast.AndStatement); valid {
			ok(t, tc, err)
			equals(t, tc, andStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseAndCCStatement validates the correct parsing of andcc commands.
func TestParseAndCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "andcc %r1, %r2, %r3",
			stmt: &ast.AndCCStatement{
				Token:       28,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "andcc %r1, 32, %r3",
			stmt: &ast.AndCCStatement{
				Token:       28,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(12), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "andcc %r1 %r2, %r3",
			err: `1:11: found REGISTER "%r2", expected ","`,
		},
		{
			str: "andcc %r1, %r2",
			err: `1:15: found EOF, expected ","`,
		},
		{
			str: "andcc %r1, %r2, 32",
			err: `1:17: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "andcc %r1, x, %r3",
			err: `1:12: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "andcc x, %r2, %r3",
			err: `1:7: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "andcc %r1, %r2, %r3, %r4",
			err: `1:20: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "andcc 32, %r2, %r3",
			err: `1:7: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "andcc %r1, 90000000000, %r3",
			err: `1:12: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nandcc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if andCCStmt, valid := tt.stmt.(*ast.AndCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, andCCStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseOrStatement validates the correct parsing of or commands.
func TestParseOrStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "or %r1, %r2, %r3",
			stmt: &ast.OrStatement{
				Token:       29,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "or %r1, 32, %r3",
			stmt: &ast.OrStatement{
				Token:       29,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(9), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "or %r1 %r2, %r3",
			err: `1:8: found REGISTER "%r2", expected ","`,
		},
		{
			str: "or %r1, %r2",
			err: `1:12: found EOF, expected ","`,
		},
		{
			str: "or %r1, %r2, 32",
			err: `1:14: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "or %r1, x, %r3",
			err: `1:9: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "or x, %r2, %r3",
			err: `1:4: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "or %r1, %r2, %r3, %r4",
			err: `1:17: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "or 32, %r2, %r3",
			err: `1:4: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "or %r1, 90000000000, %r3",
			err: `1:9: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nor %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if orStmt, valid := tt.stmt.(*ast.OrStatement); valid {
			ok(t, tc, err)
			equals(t, tc, orStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseOrCCStatement validates the correct parsing of orcc commands.
func TestParseOrCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orcc %r1, %r2, %r3",
			stmt: &ast.OrCCStatement{
				Token:       30,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orcc %r1, 32, %r3",
			stmt: &ast.OrCCStatement{
				Token:       30,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(11), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orcc %r1 %r2, %r3",
			err: `1:10: found REGISTER "%r2", expected ","`,
		},
		{
			str: "orcc %r1, %r2",
			err: `1:14: found EOF, expected ","`,
		},
		{
			str: "orcc %r1, %r2, 32",
			err: `1:16: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orcc %r1, x, %r3",
			err: `1:11: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "orcc x, %r2, %r3",
			err: `1:6: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "orcc %r1, %r2, %r3, %r4",
			err: `1:19: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "orcc 32, %r2, %r3",
			err: `1:6: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orcc %r1, 90000000000, %r3",
			err: `1:11: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\norcc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if orCCStmt, valid := tt.stmt.(*ast.OrCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, orCCStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseOrnStatement validates the correct parsing of orn commands.
func TestParseOrnStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orn %r1, %r2, %r3",
			stmt: &ast.OrnStatement{
				Token:       31,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orn %r1, 32, %r3",
			stmt: &ast.OrnStatement{
				Token:       31,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orn %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "orn %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "orn %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orn %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "orn x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "orn %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "orn 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orn %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\norn %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if ornStmt, valid := tt.stmt.(*ast.OrnStatement); valid {
			ok(t, tc, err)
			equals(t, tc, ornStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseOrnCCStatement validates the correct parsing of orncc commands.
func TestParseOrnCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orncc %r1, %r2, %r3",
			stmt: &ast.OrnCCStatement{
				Token:       32,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orncc %r1, 32, %r3",
			stmt: &ast.OrnCCStatement{
				Token:       32,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(12), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orncc %r1 %r2, %r3",
			err: `1:11: found REGISTER "%r2", expected ","`,
		},
		{
			str: "orncc %r1, %r2",
			err: `1:15: found EOF, expected ","`,
		},
		{
			str: "orncc %r1, %r2, 32",
			err: `1:17: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orncc %r1, x, %r3",
			err: `1:12: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "orncc x, %r2, %r3",
			err: `1:7: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "orncc %r1, %r2, %r3, %r4",
			err: `1:20: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "orncc 32, %r2, %r3",
			err: `1:7: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "orncc %r1, 90000000000, %r3",
			err: `1:12: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\norncc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if ornCCStmt, valid := tt.stmt.(*ast.OrnCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, ornCCStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseXorStatement validates the correct parsing of xor commands.
func TestParseXorStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "xor %r1, %r2, %r3",
			stmt: &ast.XorStatement{
				Token:       33,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xor %r1, 32, %r3",
			stmt: &ast.XorStatement{
				Token:       33,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xor %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "xor %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "xor %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "xor %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "xor x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "xor %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "xor 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "xor %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nxor %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if xorStmt, valid := tt.stmt.(*ast.XorStatement); valid {
			ok(t, tc, err)
			equals(t, tc, xorStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseXorCCStatement validates the correct parsing of xorcc commands.
func TestParseXorCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "xorcc %r1, %r2, %r3",
			stmt: &ast.XorCCStatement{
				Token:       34,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xorcc %r1, 32, %r3",
			stmt: &ast.XorCCStatement{
				Token:       34,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(12), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xorcc %r1 %r2, %r3",
			err: `1:11: found REGISTER "%r2", expected ","`,
		},
		{
			str: "xorcc %r1, %r2",
			err: `1:15: found EOF, expected ","`,
		},
		{
			str: "xorcc %r1, %r2, 32",
			err: `1:17: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "xorcc %r1, x, %r3",
			err: `1:12: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "xorcc x, %r2, %r3",
			err: `1:7: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "xorcc %r1, %r2, %r3, %r4",
			err: `1:20: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "xorcc 32, %r2, %r3",
			err: `1:7: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "xorcc %r1, 90000000000, %r3",
			err: `1:12: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nxorcc %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if xorccStmt, valid := tt.stmt.(*ast.XorCCStatement); valid {
			ok(t, tc, err)
			equals(t, tc, xorccStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseSLLStatement validates the correct parsing of sll commands.
func TestParseSLLStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sll %r1, %r2, %r3",
			stmt: &ast.SLLStatement{
				Token:       35,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sll %r1, 32, %r3",
			stmt: &ast.SLLStatement{
				Token:       35,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sll %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "sll %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "sll %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sll %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "sll x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "sll %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "sll 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sll %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nsll %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if sllStmt, valid := tt.stmt.(*ast.SLLStatement); valid {
			ok(t, tc, err)
			equals(t, tc, sllStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
		}
	}
}

// TestParseSRAStatement validates the correct parsing of sra commands.
func TestParseSRAStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sra %r1, %r2, %r3",
			stmt: &ast.SRAStatement{
				Token:       36,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sra %r1, 32, %r3",
			stmt: &ast.SRAStatement{
				Token:       36,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: 10, Position: posAfter(10), Value: 32, Literal: "32"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sra %r1 %r2, %r3",
			err: `1:9: found REGISTER "%r2", expected ","`,
		},
		{
			str: "sra %r1, %r2",
			err: `1:13: found EOF, expected ","`,
		},
		{
			str: "sra %r1, %r2, 32",
			err: `1:15: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sra %r1, x, %r3",
			err: `1:10: found IDENTIFIER "x", expected INTEGER, REGISTER`,
		},
		{
			str: "sra x, %r2, %r3",
			err: `1:5: found IDENTIFIER "x", expected REGISTER`,
		},
		{
			str: "sra %r1, %r2, %r3, %r4",
			err: `1:18: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "sra 32, %r2, %r3",
			err: `1:5: found INTEGER "32", expected REGISTER`,
		},
		{
			str: "sra %r1, 90000000000, %r3",
			err: `1:10: INTEGER "90000000000" out of 32 bit range`,
		},
		{
			str: "\nsra %r1, %r2, %r3",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra"`,
		},
	}

	for tc, tt := range tests {
		stmt, err := ParseStatement(tt.str)
		if sllStmt, valid := tt.stmt.(*ast.SRAStatement); valid {
			ok(t, tc, err)
			equals(t, tc, sllStmt, stmt)
		} else {
			equals(t, tc, tt.err, err.Error())
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
		{str: "x", obj: &ast.Identifier{Token: 8, Position: testPos, Name: "x"}},
		{str: "mylabel", obj: &ast.Identifier{Token: 8, Position: testPos, Name: "mylabel"}},
		{str: ":x", err: `1:1: found ":", expected IDENTIFIER`},
		{str: "123", err: `1:1: found INTEGER "123", expected IDENTIFIER`},
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

// TestParseRegister verifies the correct parsing of registers.
func TestParseRegister(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Register
		err string
	}{
		{str: "%r1", obj: &ast.Register{Name: "%r1"}},
		{str: "r1", err: `1:1: found IDENTIFIER "r1", expected REGISTER`},
	}

	for i, tt := range tests {
		ident, err := New(strings.NewReader(tt.str)).parseRegister()
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
		obj *ast.Integer
		err string
	}{
		{str: "0", obj: &ast.Integer{Token: 10, Position: testPos, Value: 0, Literal: "0"}},
		{str: "100", obj: &ast.Integer{Token: 10, Position: testPos, Value: 100, Literal: "100"}},
		{str: "001", obj: &ast.Integer{Token: 10, Position: testPos, Value: 1, Literal: "001"}},
		{str: "0", obj: &ast.Integer{Token: 10, Position: testPos, Value: 0, Literal: "0"}},
		{str: "0x800", obj: &ast.Integer{Token: 10, Position: testPos, Value: 2048, Literal: "0x800"}},
		{str: "90000000000000", err: `1:1: INTEGER "90000000000000" out of 32 bit range`},
		{str: "x", err: `1:1: found IDENTIFIER "x", expected INTEGER`},
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

// TestParseSIMM13 verifies the correct parsing of SIMM13 integers.
func TestParseSIMM13(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Integer
		err string
	}{
		{str: "0", obj: &ast.Integer{Token: 10, Position: testPos, Value: 0, Literal: "0"}},
		{str: "100", obj: &ast.Integer{Token: 10, Position: testPos, Value: 100, Literal: "100"}},
		{str: "001", obj: &ast.Integer{Token: 10, Position: testPos, Value: 1, Literal: "001"}},
		{str: "0", obj: &ast.Integer{Token: 10, Position: testPos, Value: 0, Literal: "0"}},
		{str: "8191", obj: &ast.Integer{Token: 10, Position: testPos, Value: 8191, Literal: "8191"}},
		{str: "0x800", obj: &ast.Integer{Token: 10, Position: testPos, Value: 2048, Literal: "0x800"}},
		{str: "8192", err: `1:1: INTEGER "8192" is not a valid SIMM13`},
		{str: "-1", err: `1:1: found "-", expected INTEGER`},
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

// TestParseExpression verifies the correct parsing of expressions.
func TestParseExpression(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Expression
		err string
	}{
		{str: "[%r1+8191]", obj: &ast.Expression{Base: &ast.Register{Name: "%r1"}, Operator: "+", Offset: &ast.Integer{Token: 10, Position: posAfter(6), Value: 8191, Literal: "8191"}}},
		{str: "[%r1+0]", obj: &ast.Expression{Base: &ast.Register{Name: "%r1"}, Operator: "+", Offset: &ast.Integer{Token: 10, Position: posAfter(6), Value: 0, Literal: "0"}}},
		{str: "[x]", obj: &ast.Expression{Base: &ast.Identifier{Token: 8, Position: posAfter(2), Name: "x"}, Operator: "", Offset: nil}},
		{str: "x]", err: `1:1: found IDENTIFIER "x", expected "["`},
		{str: "[+8191]", err: `1:2: found "+", expected IDENTIFIER, REGISTER`},
		{str: "[0+8191]", err: `1:2: found INTEGER "0", expected IDENTIFIER, REGISTER`},
		{str: "[%r1 8191]", err: `1:6: found INTEGER "8191", expected "+", "-", "]"`},
		{str: "[%r1*8191]", err: `1:5: found ILLEGAL "*", expected "+", "-", "]"`},
		{str: "[%r1+]", err: `1:6: found "]", expected INTEGER`},
		{str: "[%r1+45", err: `1:8: found EOF, expected "]"`},
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

// TestParseMemoryLocation verifies the correct parsing of operands.
func TestParseOperand(t *testing.T) {
	tests := []struct {
		str string
		obj ast.Operand
		err string
	}{
		{str: "64", obj: &ast.Integer{Token: 10, Position: testPos, Value: 64, Literal: "64"}},
		{str: "%r1", obj: &ast.Register{Name: "%r1"}},
		{str: "x", err: `1:1: found IDENTIFIER "x", expected INTEGER, REGISTER`},
	}

	for i, tt := range tests {
		loc, err := New(strings.NewReader(tt.str)).parseOperand()
		if err == nil {
			ok(t, i, err)
			equals(t, i, tt.obj, loc)
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
		{
			str: "[x]", obj: &ast.Expression{
				Position: testPos,
				Base: &ast.Identifier{Token: 8,
					Position: token.Pos{Line: 1, Char: 2},
					Name:     "x",
				},
			},
		},
		{str: "%r1", obj: &ast.Register{Name: "%r1"}},
		{str: "x", err: `1:1: found IDENTIFIER "x", expected "[", REGISTER`},
		{str: "123", err: `1:1: found INTEGER "123", expected "[", REGISTER`},
		{str: "[x+]", err: `1:4: found "]", expected INTEGER`},
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
		{str: "\t", err: `1:1: found ILLEGAL, expected NEWLINE, EOF`},
		{str: ";", err: `1:1: found ILLEGAL ";", expected NEWLINE, EOF`},
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

func posAfter(char int) token.Pos {
	return token.Pos{Line: 1, Char: char}
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

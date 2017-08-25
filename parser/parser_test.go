package parser

import (
	"errors"
	"fmt"
	"os"
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
var (
	validProg = `
! main.arc
! This is a valid ARC sample program.
.begin
.org 0x800
main:   ld [x], %r1				! Load x.
        ld [y], %r2				! Load y.
        add %r1, %r2, %r3
		subcc %r3, 2, %r4
		sll %r4, 3, %r5
        st %r5, [z]
        ba exit					! Always branch to exit routine.
exit:	ld [z], %r6				! jmpl %r15 + 4, %r6

! Start data section at 0x1000.
.org 0x1000
x: 2
y: 4
z: 0
.end

`

	arraySum = `
! ------------------------------------------------------- !
! This program sums the elements from array that is       !
! located starting with 3000.                             !
! ------------------------------------------------------- !
! Used registers                                          !
! ==============                                          !
! r1: length                                              !
! r2: start (3000)                                        !
! r3: sum of the elements (is initialized with zero)      !
! r4: the current element                                 !
! ==============                                          !
! r1, r2 and r4 are set back to 0 after the loop is done  !
! ------------------------------------------------------- !

        .begin
        .org 2048
        call init_r
        call loop

init_r: ld [length], %r1
        ld [start], %r2
        ld [zero], %r3
        jmpl %r15+4, %r0

loop:   ld %r2, %r4
        addcc %r2, 4, %r2
        addcc %r3, %r4, %r3
        addcc %r1, -1, %r1
        be done
        ba loop

done:   ld [zero], %r1
        ld [zero], %r2
        ld [zero], %r4
        jmpl %r15+4, %r0

start:  3000
length: 4
zero:   0

        .org 3000
        10
        20
        -0xa
        aH
        .end
`
)

// TestParserBuffer tests if the correct token is returned after an unscan().
func TestParserBuffer(t *testing.T) {
	// Scan and save token, literal and token position.
	test := `ld %r1, %r2`
	p := New(strings.NewReader(test))
	p.scan()
	tok, lit, pos := p.tok, p.lit, p.pos

	// Unscan and check buffer content.
	p.unscan()
	bufTok, bufLit, bufPos := p.buf.tok, p.buf.lit, p.buf.pos
	equals(t, tok, bufTok)
	equals(t, lit, bufLit)
	equals(t, pos, bufPos)
}

// TestFeed tests if the parser is feed with the new data on top of the
// previously parsed data.
func TestFeed(t *testing.T) {
	stmt1, stmt2 := "x: 25", "ld [x], %r2"

	// Parse first statement.
	p := New(strings.NewReader(stmt1))
	prog, err := p.Parse()
	ok(t, err)
	equals(t, 1, len(prog.Statements))

	// Parse next statement and check the error. Since "x" was already resolved
	// the error should be nil.
	p.Feed(stmt2)
	prog, err = p.Parse()
	ok(t, err)
	equals(t, 1, len(prog.Statements))
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
			prog: validProg,
			// err:  `stuff`,
		},
		{
			prog: arraySum,
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
			err: `3:6: found KEYWORD "ld", expected "[", REGISTER
7:6: found IDENTIFIER "x", expected INTEGER, "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
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

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			_, err := Parse(tt.prog)
			if tt.err == "" {
				ok(t, err)
			} else {
				// if err == nil {
				// 	t.Fatalf("expected error but got nil")
				// }
				assert(t, err != nil, "expected error but got nil")
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParseFile will validate the correct parsing of a file containing a
// complete program.
func TestParseFile(t *testing.T) {
	err := os.Chdir("../testdata")
	if err != nil {
		t.Error("could not switch to testdata directory")
	}

	tests := []struct {
		file string
		err  string
	}{
		{file: "valid.arc"},
		{file: "notExisting.arc", err: "open notExisting.arc: no such file or directory"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			_, err := ParseFile(tt.file)
			if tt.err == "" {
				ok(t, err)
			} else {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseCommentStatement validates the correct parsing of the begin directive.
func TestParser_ParseCommentStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: "!  This is a comment  ", stmt: &ast.CommentStatement{Token: token.COMMENT, Position: testPos, Text: "!  This is a comment  "}},
		{str: "This is not a comment", err: `1:6: found IDENTIFIER "is", expected ":"`},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if commentStmt, valid := tt.stmt.(*ast.CommentStatement); valid {
				ok(t, err)
				equals(t, commentStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBeginStatement validates the correct parsing of the begin directive.
func TestParser_ParseBeginStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".begin", stmt: &ast.BeginStatement{Token: token.BEGIN, Position: testPos}},
		{str: ".beg", err: `1:1: found ILLEGAL ".beg", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`},
		{str: "begin", err: `1:6: found EOF, expected ":"`},
		{str: ".begin 123", err: `1:8: found INTEGER "123", expected COMMENT, NEWLINE, EOF`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if beginStmt, valid := tt.stmt.(*ast.BeginStatement); valid {
				ok(t, err)
				equals(t, beginStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseEndStatement validates the correct parsing of the end directive.
func TestParser_ParseEndStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".end", stmt: &ast.EndStatement{Token: token.END, Position: testPos}},
		{str: ".ed", err: `1:1: found ILLEGAL ".ed", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`},
		{str: "end", err: `1:4: found EOF, expected ":"`},
		{str: ".end 123", err: `1:6: found INTEGER "123", expected COMMENT, NEWLINE, EOF`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if endStmt, valid := tt.stmt.(*ast.EndStatement); valid {
				ok(t, err)
				equals(t, endStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseOrgStatement validates the correct parsing of the org directive.
func TestParser_ParseOrgStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{str: ".org 2048", stmt: &ast.OrgStatement{Token: token.ORG, Position: testPos, Value: &ast.Integer{Token: token.INT, Position: posAfter(6), Value: 2048, Literal: "2048"}}},
		{str: ".org 2048 128", err: `1:11: found INTEGER "128", expected COMMENT, NEWLINE, EOF`},
		{str: ".org", err: `1:5: found EOF, expected INTEGER`},
		{str: ".og", err: `1:1: found ILLEGAL ".og", expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`},
		{str: "org", err: `1:4: found EOF, expected ":"`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if orgStmt, valid := tt.stmt.(*ast.OrgStatement); valid {
				ok(t, err)
				equals(t, orgStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseLabelStatement validates the correct parsing of st commands.
func TestParser_ParseLabelStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "x: 25",
			stmt: &ast.LabelStatement{
				Token:     token.IDENT,
				Position:  testPos,
				Ident:     &ast.Identifier{Token: token.IDENT, Position: testPos, Name: "x"},
				Reference: &ast.Integer{Token: token.INT, Position: posAfter(4), Value: 25, Literal: "25"},
			},
		},
		{
			str: "mylabel: ld %r1, %r2",
			stmt: &ast.LabelStatement{
				Token:    token.IDENT,
				Position: testPos,
				Ident:    &ast.Identifier{Token: token.IDENT, Position: testPos, Name: "mylabel"},
				Reference: &ast.LoadStatement{
					Token:       token.LOAD,
					Position:    posAfter(10),
					Source:      &ast.Register{Name: "%r1"},
					Destination: &ast.Register{Name: "%r2"},
				},
			},
		},
		{str: "x: y: 25", err: `1:4: found IDENTIFIER "y", expected INTEGER, "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`},
		{str: "x: 25;", err: `1:6: found ILLEGAL ";", expected COMMENT, NEWLINE, EOF`},
		{str: "x: ld", err: `1:6: found EOF, expected "[", REGISTER`},
		{str: "X: 90000000000000", err: `1:4: INTEGER "90000000000000" out of 32 bit range`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if labelStmt, valid := tt.stmt.(*ast.LabelStatement); valid {
				ok(t, err)
				equals(t, labelStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseLoadStatement validates the correct parsing of load commands.
func TestParser_ParseLoadStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "ld %r1, %r2",
			stmt: &ast.LoadStatement{
				Token:       token.LOAD,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "ld [x], %r2",
			stmt: &ast.LoadStatement{
				Token:    token.LOAD,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base: &ast.Identifier{Token: token.IDENT,
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
				Token:    token.LOAD,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: token.INT, Position: posAfter(9), Value: 8191, Literal: "8191"},
				},
				Destination: &ast.Register{Name: "%r2"},
			},
		},
		{
			str: "ld [%r1+0], %r2",
			stmt: &ast.LoadStatement{
				Token:    token.LOAD,
				Position: testPos,
				Source: &ast.Expression{
					Position: posAfter(4),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: token.INT, Position: posAfter(9), Value: 0, Literal: "0"},
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
			err: `1:4: found KEYWORD "ld", expected "[", REGISTER`,
		},
		{
			str: "ld %r1, ld",
			err: `1:9: found KEYWORD "ld", expected REGISTER`,
		},
		{
			str: "ld %r1 %r2",
			err: `1:8: found REGISTER "%r2", expected ","`,
		},
		{
			str: "ld %r1, %r2, %r3",
			err: `1:12: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nld %r1, %r2",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if loadStmt, valid := tt.stmt.(*ast.LoadStatement); valid {
				ok(t, err)
				equals(t, loadStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseStoreStatement validates the correct parsing of store commands.
func TestParser_ParseStoreStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "st %r2, %r1",
			stmt: &ast.StoreStatement{
				Token:       token.STORE,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r1"},
			},
		},
		{
			str: "st %r2, [x]",
			stmt: &ast.StoreStatement{
				Token:    token.STORE,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base: &ast.Identifier{Token: token.IDENT,
						Position: posAfter(10),
						Name:     "x",
					},
				},
			},
		},
		{
			str: "st %r2, [%r1+8191]",
			stmt: &ast.StoreStatement{
				Token:    token.STORE,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: token.INT, Position: posAfter(14), Value: 8191, Literal: "8191"},
				},
			},
		},
		{
			str: "st %r2, [%r1+0]",
			stmt: &ast.StoreStatement{
				Token:    token.STORE,
				Position: testPos,
				Source:   &ast.Register{Name: "%r2"},
				Destination: &ast.Expression{
					Position: posAfter(9),
					Base:     &ast.Register{Name: "%r1"},
					Operator: "+",
					Offset:   &ast.Integer{Token: token.INT, Position: posAfter(14), Value: 0, Literal: "0"},
				},
			},
		},
		{
			str: "s %r2, %r1",
			err: `1:3: found REGISTER "%r2", expected ":"`,
		},
		{
			str: "st st, %r1",
			err: `1:4: found KEYWORD "st", expected REGISTER`,
		},
		{
			str: "st %r2, st",
			err: `1:9: found KEYWORD "st", expected "[", REGISTER`,
		},
		{
			str: "st %r2 %r1",
			err: `1:8: found REGISTER "%r1", expected ","`,
		},
		{
			str: "st %r2, %r1, %r3",
			err: `1:12: found ",", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nst %r2, %r1",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if storeStmt, valid := tt.stmt.(*ast.StoreStatement); valid {
				ok(t, err)
				equals(t, storeStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseAddStatement validates the correct parsing of add commands.
func TestParser_ParseAddStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "add %r1, %r2, %r3",
			stmt: &ast.AddStatement{
				Token:       token.ADD,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "add %r1, 32, %r3",
			stmt: &ast.AddStatement{
				Token:       token.ADD,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if addStmt, valid := tt.stmt.(*ast.AddStatement); valid {
				ok(t, err)
				equals(t, addStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseAddCCStatement validates the correct parsing of addcc commands.
func TestParser_ParseAddCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "addcc %r1, %r2, %r3",
			stmt: &ast.AddCCStatement{
				Token:       token.ADDCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "addcc %r1, 32, %r3",
			stmt: &ast.AddCCStatement{
				Token:       token.ADDCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(12), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if addCCStmt, valid := tt.stmt.(*ast.AddCCStatement); valid {
				ok(t, err)
				equals(t, addCCStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseSubStatement validates the correct parsing of sub commands.
func TestParser_ParseSubStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sub %r1, %r2, %r3",
			stmt: &ast.SubStatement{
				Token:       token.SUB,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sub %r1, 32, %r3",
			stmt: &ast.SubStatement{
				Token:       token.SUB,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if subStmt, valid := tt.stmt.(*ast.SubStatement); valid {
				ok(t, err)
				equals(t, subStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseSubCCStatement validates the correct parsing of subcc commands.
func TestParser_ParseSubCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "subcc %r1, %r2, %r3",
			stmt: &ast.SubCCStatement{
				Token:       token.SUBCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "subcc %r1, 32, %r3",
			stmt: &ast.SubCCStatement{
				Token:       token.SUBCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(12), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if subCCStmt, valid := tt.stmt.(*ast.SubCCStatement); valid {
				ok(t, err)
				equals(t, subCCStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseAndStatement validates the correct parsing of and commands.
func TestParser_ParseAndStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "and %r1, %r2, %r3",
			stmt: &ast.AndStatement{
				Token:       token.AND,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "and %r1, 32, %r3",
			stmt: &ast.AndStatement{
				Token:       token.AND,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if andStmt, valid := tt.stmt.(*ast.AndStatement); valid {
				ok(t, err)
				equals(t, andStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseAndCCStatement validates the correct parsing of andcc commands.
func TestParser_ParseAndCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "andcc %r1, %r2, %r3",
			stmt: &ast.AndCCStatement{
				Token:       token.ANDCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "andcc %r1, 32, %r3",
			stmt: &ast.AndCCStatement{
				Token:       token.ANDCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(12), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if andCCStmt, valid := tt.stmt.(*ast.AndCCStatement); valid {
				ok(t, err)
				equals(t, andCCStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseOrStatement validates the correct parsing of or commands.
func TestParser_ParseOrStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "or %r1, %r2, %r3",
			stmt: &ast.OrStatement{
				Token:       token.OR,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "or %r1, 32, %r3",
			stmt: &ast.OrStatement{
				Token:       token.OR,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(9), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if orStmt, valid := tt.stmt.(*ast.OrStatement); valid {
				ok(t, err)
				equals(t, orStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseOrCCStatement validates the correct parsing of orcc commands.
func TestParser_ParseOrCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orcc %r1, %r2, %r3",
			stmt: &ast.OrCCStatement{
				Token:       token.ORCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orcc %r1, 32, %r3",
			stmt: &ast.OrCCStatement{
				Token:       token.ORCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(11), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if orCCStmt, valid := tt.stmt.(*ast.OrCCStatement); valid {
				ok(t, err)
				equals(t, orCCStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseOrnStatement validates the correct parsing of orn commands.
func TestParser_ParseOrnStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orn %r1, %r2, %r3",
			stmt: &ast.OrnStatement{
				Token:       token.ORN,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orn %r1, 32, %r3",
			stmt: &ast.OrnStatement{
				Token:       token.ORN,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if ornStmt, valid := tt.stmt.(*ast.OrnStatement); valid {
				ok(t, err)
				equals(t, ornStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseOrnCCStatement validates the correct parsing of orncc commands.
func TestParser_ParseOrnCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "orncc %r1, %r2, %r3",
			stmt: &ast.OrnCCStatement{
				Token:       token.ORNCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "orncc %r1, 32, %r3",
			stmt: &ast.OrnCCStatement{
				Token:       token.ORNCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(12), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if ornCCStmt, valid := tt.stmt.(*ast.OrnCCStatement); valid {
				ok(t, err)
				equals(t, ornCCStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseXorStatement validates the correct parsing of xor commands.
func TestParser_ParseXorStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "xor %r1, %r2, %r3",
			stmt: &ast.XorStatement{
				Token:       token.XOR,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xor %r1, 32, %r3",
			stmt: &ast.XorStatement{
				Token:       token.XOR,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if xorStmt, valid := tt.stmt.(*ast.XorStatement); valid {
				ok(t, err)
				equals(t, xorStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseXorCCStatement validates the correct parsing of xorcc commands.
func TestParser_ParseXorCCStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "xorcc %r1, %r2, %r3",
			stmt: &ast.XorCCStatement{
				Token:       token.XORCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "xorcc %r1, 32, %r3",
			stmt: &ast.XorCCStatement{
				Token:       token.XORCC,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(12), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if xorccStmt, valid := tt.stmt.(*ast.XorCCStatement); valid {
				ok(t, err)
				equals(t, xorccStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseSLLStatement validates the correct parsing of sll commands.
func TestParser_ParseSLLStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sll %r1, %r2, %r3",
			stmt: &ast.SLLStatement{
				Token:       token.SLL,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sll %r1, 32, %r3",
			stmt: &ast.SLLStatement{
				Token:       token.SLL,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if sllStmt, valid := tt.stmt.(*ast.SLLStatement); valid {
				ok(t, err)
				equals(t, sllStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseSRAStatement validates the correct parsing of sra commands.
func TestParser_ParseSRAStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "sra %r1, %r2, %r3",
			stmt: &ast.SRAStatement{
				Token:       token.SRA,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Register{Name: "%r2"},
				Destination: &ast.Register{Name: "%r3"},
			},
		},
		{
			str: "sra %r1, 32, %r3",
			stmt: &ast.SRAStatement{
				Token:       token.SRA,
				Position:    testPos,
				Source:      &ast.Register{Name: "%r1"},
				Operand:     &ast.Integer{Token: token.INT, Position: posAfter(10), Value: 32, Literal: "32"},
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
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if sllStmt, valid := tt.stmt.(*ast.SRAStatement); valid {
				ok(t, err)
				equals(t, sllStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBEStatement validates the correct parsing of be commands.
func TestParser_ParseBEStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "be x",
			stmt: &ast.BEStatement{
				Token:    token.BE,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(4), Name: "x"},
			},
		},
		{
			str: "be main",
			stmt: &ast.BEStatement{
				Token:    token.BE,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(4), Name: "main"},
			},
		},
		{
			str: "be %r1",
			err: `1:4: found REGISTER "%r1", expected IDENTIFIER`,
		},
		{
			str: "be 123",
			err: `1:4: found INTEGER "123", expected IDENTIFIER`,
		},
		{
			str: "be be",
			err: `1:4: found KEYWORD "be", expected IDENTIFIER`,
		},
		{
			str: "be main x",
			err: `1:9: found IDENTIFIER "x", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nbe x",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if beStmt, valid := tt.stmt.(*ast.BEStatement); valid {
				ok(t, err)
				equals(t, beStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBNEStatement validates the correct parsing of bne commands.
func TestParser_ParseBNEStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "bne x",
			stmt: &ast.BNEStatement{
				Token:    token.BNE,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(5), Name: "x"},
			},
		},
		{
			str: "bne main",
			stmt: &ast.BNEStatement{
				Token:    token.BNE,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(5), Name: "main"},
			},
		},
		{
			str: "bne %r1",
			err: `1:5: found REGISTER "%r1", expected IDENTIFIER`,
		},
		{
			str: "bne 123",
			err: `1:5: found INTEGER "123", expected IDENTIFIER`,
		},
		{
			str: "bne bne",
			err: `1:5: found KEYWORD "bne", expected IDENTIFIER`,
		},
		{
			str: "bne main x",
			err: `1:10: found IDENTIFIER "x", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nbne x",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if bneStmt, valid := tt.stmt.(*ast.BNEStatement); valid {
				ok(t, err)
				equals(t, bneStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBNEGStatement validates the correct parsing of bneg commands.
func TestParser_ParseBNEGStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "bneg x",
			stmt: &ast.BNEGStatement{
				Token:    token.BNEG,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(6), Name: "x"},
			},
		},
		{
			str: "bneg main",
			stmt: &ast.BNEGStatement{
				Token:    token.BNEG,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(6), Name: "main"},
			},
		},
		{
			str: "bneg %r1",
			err: `1:6: found REGISTER "%r1", expected IDENTIFIER`,
		},
		{
			str: "bneg 123",
			err: `1:6: found INTEGER "123", expected IDENTIFIER`,
		},
		{
			str: "bneg bneg",
			err: `1:6: found KEYWORD "bneg", expected IDENTIFIER`,
		},
		{
			str: "bneg main x",
			err: `1:11: found IDENTIFIER "x", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nbneg x",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if bnegStmt, valid := tt.stmt.(*ast.BNEGStatement); valid {
				ok(t, err)
				equals(t, bnegStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBPOSStatement validates the correct parsing of bpos commands.
func TestParser_ParseBPOSStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "bpos x",
			stmt: &ast.BPOSStatement{
				Token:    token.BPOS,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(6), Name: "x"},
			},
		},
		{
			str: "bpos main",
			stmt: &ast.BPOSStatement{
				Token:    token.BPOS,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(6), Name: "main"},
			},
		},
		{
			str: "bpos %r1",
			err: `1:6: found REGISTER "%r1", expected IDENTIFIER`,
		},
		{
			str: "bpos 123",
			err: `1:6: found INTEGER "123", expected IDENTIFIER`,
		},
		{
			str: "bpos bpos",
			err: `1:6: found KEYWORD "bpos", expected IDENTIFIER`,
		},
		{
			str: "bpos main x",
			err: `1:11: found IDENTIFIER "x", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nbneg x",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if bposStmt, valid := tt.stmt.(*ast.BPOSStatement); valid {
				ok(t, err)
				equals(t, bposStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseBAStatement validates the correct parsing of ba commands.
func TestParser_ParseBAStatement(t *testing.T) {
	tests := []struct {
		str  string
		stmt ast.Statement
		err  string
	}{
		{
			str: "ba x",
			stmt: &ast.BAStatement{
				Token:    token.BA,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(4), Name: "x"},
			},
		},
		{
			str: "ba main",
			stmt: &ast.BAStatement{
				Token:    token.BA,
				Position: testPos,
				Target:   &ast.Identifier{Token: token.IDENT, Position: posAfter(4), Name: "main"},
			},
		},
		{
			str: "ba %r1",
			err: `1:4: found REGISTER "%r1", expected IDENTIFIER`,
		},
		{
			str: "ba 123",
			err: `1:4: found INTEGER "123", expected IDENTIFIER`,
		},
		{
			str: "ba ba",
			err: `1:4: found KEYWORD "ba", expected IDENTIFIER`,
		},
		{
			str: "ba main x",
			err: `1:9: found IDENTIFIER "x", expected COMMENT, NEWLINE, EOF`,
		},
		{
			str: "\nbe x",
			err: `1:1: found NEWLINE, expected COMMENT, IDENTIFIER, ".begin", ".end", ".org", "ld", "st", "add", "addcc", "sub", "subcc", "and", "andcc", "or", "orcc", "orn", "orncc", "xor", "xorcc", "sll", "sra", "be", "bne", "bneg", "bpos", "ba", "call", "jmpl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			stmt, err := ParseStatement(tt.str)
			if baStmt, valid := tt.stmt.(*ast.BAStatement); valid {
				ok(t, err)
				equals(t, baStmt, stmt)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseIdent verifies the correct parsing of identifiers.
func TestParser_ParseIdent(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Identifier
		err string
	}{
		{str: "x", obj: &ast.Identifier{Token: token.IDENT, Position: testPos, Name: "x"}},
		{str: "mylabel", obj: &ast.Identifier{Token: token.IDENT, Position: testPos, Name: "mylabel"}},
		{str: ":x", err: `1:1: found ":", expected IDENTIFIER`},
		{str: "123", err: `1:1: found INTEGER "123", expected IDENTIFIER`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			ident, err := New(strings.NewReader(tt.str)).parseIdent()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, ident)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseRegister verifies the correct parsing of registers.
func TestParser_ParseRegister(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Register
		err string
	}{
		{str: "%r1", obj: &ast.Register{Name: "%r1"}},
		{str: "r1", err: `1:1: found IDENTIFIER "r1", expected REGISTER`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			ident, err := New(strings.NewReader(tt.str)).parseRegister()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, ident)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseInteger verifies the correct parsing of integers.
func TestParser_ParseInteger(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Integer
		err string
	}{
		{str: "0", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 0, Literal: "0"}},
		{str: "100", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 100, Literal: "100"}},
		{str: "001", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 1, Literal: "001"}},
		{str: "0", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 0, Literal: "0"}},
		{str: "0x800", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 2048, Literal: "0x800"}},
		{str: "90000000000000", err: `1:1: INTEGER "90000000000000" out of 32 bit range`},
		{str: "x", err: `1:1: found IDENTIFIER "x", expected INTEGER`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			integer, err := New(strings.NewReader(tt.str)).parseInteger()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, integer)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseSIMM13 verifies the correct parsing of SIMM13 integers.
func TestParser_ParseSIMM13(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Integer
		err string
	}{
		{str: "0", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 0, Literal: "0"}},
		{str: "100", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 100, Literal: "100"}},
		{str: "001", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 1, Literal: "001"}},
		{str: "0", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 0, Literal: "0"}},
		{str: "8191", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 8191, Literal: "8191"}},
		{str: "0x800", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 2048, Literal: "0x800"}},
		{str: "8192", err: `1:1: INTEGER "8192" is not a valid SIMM13`},
		{str: "-1", err: `1:1: found "-", expected INTEGER`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			integer, err := New(strings.NewReader(tt.str)).parseSIMM13()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, integer)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseExpression verifies the correct parsing of expressions.
func TestParser_ParseExpression(t *testing.T) {
	tests := []struct {
		str string
		obj *ast.Expression
		err string
	}{
		{str: "[%r1+8191]", obj: &ast.Expression{Base: &ast.Register{Name: "%r1"}, Operator: "+", Offset: &ast.Integer{Token: token.INT, Position: posAfter(6), Value: 8191, Literal: "8191"}}},
		{str: "[%r1+0]", obj: &ast.Expression{Base: &ast.Register{Name: "%r1"}, Operator: "+", Offset: &ast.Integer{Token: token.INT, Position: posAfter(6), Value: 0, Literal: "0"}}},
		{str: "[x]", obj: &ast.Expression{Base: &ast.Identifier{Token: token.IDENT, Position: posAfter(2), Name: "x"}, Operator: "", Offset: nil}},
		{str: "x]", err: `1:1: found IDENTIFIER "x", expected "["`},
		{str: "[+8191]", err: `1:2: found "+", expected IDENTIFIER, REGISTER`},
		{str: "[0+8191]", err: `1:2: found INTEGER "0", expected IDENTIFIER, REGISTER`},
		{str: "[%r1 8191]", err: `1:6: found INTEGER "8191", expected "+", "-", "]"`},
		{str: "[%r1*8191]", err: `1:5: found ILLEGAL "*", expected "+", "-", "]"`},
		{str: "[%r1+]", err: `1:6: found "]", expected INTEGER`},
		{str: "[%r1+45", err: `1:8: found EOF, expected "]"`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			exp, err := New(strings.NewReader(tt.str)).parseExpression()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, exp)

			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseMemoryLocation verifies the correct parsing of operands.
func TestParser_ParseOperand(t *testing.T) {
	tests := []struct {
		str string
		obj ast.Operand
		err string
	}{
		{str: "64", obj: &ast.Integer{Token: token.INT, Position: testPos, Value: 64, Literal: "64"}},
		{str: "%r1", obj: &ast.Register{Name: "%r1"}},
		{str: "x", err: `1:1: found IDENTIFIER "x", expected INTEGER, REGISTER`},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			loc, err := New(strings.NewReader(tt.str)).parseOperand()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, loc)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// TestParser_ParseMemoryLocation verifies the correct parsing of memory locations.
func TestParser_ParseMemoryLocation(t *testing.T) {
	tests := []struct {
		str string
		obj ast.MemoryLocation
		err string
	}{
		{
			str: "[x]", obj: &ast.Expression{
				Position: testPos,
				Base: &ast.Identifier{Token: token.IDENT,
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

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			loc, err := New(strings.NewReader(tt.str)).parseMemoryLocation()
			if err == nil {
				ok(t, err)
				equals(t, tt.obj, loc)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
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

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			err := New(strings.NewReader(tt.str)).expectStatementEnd()
			if err == nil {
				ok(t, err)
			} else {
				equals(t, tt.err, err.Error())
			}
		})
	}
}

// BenchmarkParse benchmarks the overall parsing performance for a valid ARC
// program.
func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(validProg)
	}
}

func posAfter(char int) token.Pos {
	return token.Pos{Line: 1, Char: char}
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	tb.Helper()
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unttected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if tt is not equal to act.
func equals(tb testing.TB, tt, act interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(tt, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, tt, act)
		tb.FailNow()
	}
}

/*
Package ast implements the abstract syntax tree of the ARC assembly language. It
provides generic interfaces, ast objects, data types and helper methods.
*/
package ast

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/lukasmalkmus/arc/token"
)

// Statement is an ARC assembly statement.
type Statement interface {
	// stmt is unexported to ensure implementations of Statement can only
	// originate in this package.
	stmt()
	Pos() token.Pos
	String() string
}

func (*CommentStatement) stmt() {}
func (*BeginStatement) stmt()   {}
func (*EndStatement) stmt()     {}
func (*OrgStatement) stmt()     {}
func (*LabelStatement) stmt()   {}
func (*LoadStatement) stmt()    {}
func (*StoreStatement) stmt()   {}
func (*AddStatement) stmt()     {}
func (*AddCCStatement) stmt()   {}
func (*SubStatement) stmt()     {}
func (*SubCCStatement) stmt()   {}
func (*AndStatement) stmt()     {}
func (*AndCCStatement) stmt()   {}
func (*OrStatement) stmt()      {}
func (*OrCCStatement) stmt()    {}
func (*OrnStatement) stmt()     {}
func (*OrnCCStatement) stmt()   {}
func (*XorStatement) stmt()     {}
func (*XorCCStatement) stmt()   {}
func (*SLLStatement) stmt()     {}
func (*SRAStatement) stmt()     {}

// Reference is implemented by types which can be referenced by a label. These
// are statements and identifiers.
type Reference interface {
	// ref is unexported to ensure implementations of Reference can only
	// originate in this package.
	ref()
	String() string
}

func (Integer) ref()         {}
func (*LoadStatement) ref()  {}
func (*StoreStatement) ref() {}
func (*AddStatement) ref()   {}
func (*AddCCStatement) ref() {}
func (*SubStatement) ref()   {}
func (*SubCCStatement) ref() {}
func (*AndStatement) ref()   {}
func (*AndCCStatement) ref() {}
func (*OrStatement) ref()    {}
func (*OrCCStatement) ref()  {}
func (*OrnStatement) ref()   {}
func (*OrnCCStatement) ref() {}
func (*XorStatement) ref()   {}
func (*XorCCStatement) ref() {}
func (*SLLStatement) ref()   {}
func (*SRAStatement) ref()   {}

// MemoryLocation is implemented by types which can be addressed as locations in
// memory. Expressions can be addressed as well as registers.
type MemoryLocation interface {
	// memLoc is unexported to ensure implementations of Memory can only
	// originate in this package.
	memLoc()
	String() string
}

func (*Expression) memLoc() {}
func (*Register) memLoc()   {}

// ExpressionBase is implemented by types which can be the base values of
// expressions. These are identifiers and registers.
type ExpressionBase interface {
	// epb is unexported to ensure implementations of Reference can only
	// originate in this package.
	epb()
	String() string
}

func (*Identifier) epb() {}
func (*Register) epb()   {}

// Operand is implemented by types which can be used as operands in arithmetic
// operations.
type Operand interface {
	// op is unexported to ensure implementations of Reference can only
	// originate in this package.
	op()
	String() string
}

func (Integer) op()   {}
func (*Register) op() {}

// Statements is a list of statements.
type Statements []Statement

func (s Statements) String() string {
	var str []string
	for _, stmt := range s {
		str = append(str, stmt.String())
	}
	return strings.Join(str, "\n")
}

// Program represents a collection of statements.
type Program struct {
	// Filename is the name of the file containing the programs source code.
	Filename token.Pos
	// Statements is the list of statements building the program.
	Statements Statements
}

func (p Program) String() string { return p.Statements.String() }

// AddStatement adds one or more Statements to the Program.
func (p *Program) AddStatement(stmts ...Statement) {
	for _, stmt := range stmts {
		if stmt != nil {
			p.Statements = append(p.Statements, stmt)
		}
	}
}

// CommentStatement represents a comment.
type CommentStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Text is the actual text of the comment.
	Text string
}

// Pos returns the statements position.
func (stmt CommentStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt CommentStatement) String() string {
	return "! " + strings.TrimSpace(stmt.Text[1:])
}

// BeginStatement marks the beginning of an ARC program.
type BeginStatement struct {
	// Position is the position in the source.
	Position token.Pos
}

// Pos returns the statements position.
func (stmt BeginStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt BeginStatement) String() string {
	return ".begin"
}

// EndStatement marks the end of an ARC program.
type EndStatement struct {
	// Position is the position in the source.
	Position token.Pos
}

// Pos returns the statements position.
func (stmt EndStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt EndStatement) String() string {
	return ".end"
}

// OrgStatement marks a new section of data in memory.
type OrgStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Value is the memory location.
	Value Integer
}

// Pos returns the statements position.
func (stmt OrgStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt OrgStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString(".org ")
	buf.WriteString(stmt.Value.String())
	return buf.String()
}

// LabelStatement represents a label.
type LabelStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Ident is the labels identifier.
	Ident *Identifier
	// Reference is an Identifier, Integer or the Statement the label addresses.
	Reference Reference
}

// Pos returns the statements position.
func (stmt LabelStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt LabelStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString(stmt.Ident.String())
	buf.WriteString(": ")
	buf.WriteString(stmt.Reference.String())
	return buf.String()
}

// LoadStatement represents a load command (ld).
type LoadStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is the memory location where the value is loaded from.
	Source MemoryLocation
	// Destination is the register where the value is loaded to.
	Destination *Register
}

// Pos returns the statements position.
func (stmt LoadStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt LoadStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("ld ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// StoreStatement represents a store command (st).
type StoreStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is the register where the value is stored from.
	Source *Register
	// Destination is the memory location where the value is stored to.
	Destination MemoryLocation
}

// Pos returns the statements position.
func (stmt StoreStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt StoreStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("st ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// AddStatement represents an add command (add).
type AddStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the arithmetic
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt AddStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt AddStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("add ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// AddCCStatement represents an add (conditional codes set) command (addcc).
type AddCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the arithmetic
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt AddCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt AddCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("addcc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// SubStatement represents a sub command (sub).
type SubStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the arithmetic
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt SubStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt SubStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("sub ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// SubCCStatement represents a sub (conditional codes set) command (subcc).
type SubCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the arithmetic
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt SubCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt SubCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("subcc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// AndStatement represents an and command (and).
type AndStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt AndStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt AndStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("and ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// AndCCStatement represents an and (conditional codes set) command (andcc).
type AndCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt AndCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt AndCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("andcc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// OrStatement represents an or command (or).
type OrStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt OrStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt OrStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("or ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// OrCCStatement represents an or (conditional codes set) command (orcc).
type OrCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt OrCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt OrCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("orcc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// OrnStatement represents a orn command (orn).
type OrnStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt OrnStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt OrnStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("orn ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// OrnCCStatement represents a orn (conditional codes set) command (orncc).
type OrnCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt OrnCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt OrnCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("orncc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// XorStatement represents a xor command (xor).
type XorStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt XorStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt XorStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("xor ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// XorCCStatement represents a xor (conditional codes set) command (xorcc).
type XorCCStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt XorCCStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt XorCCStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("xorcc ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// SLLStatement represents a shift left logical command (sll).
type SLLStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the logical
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt SLLStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt SLLStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("orn ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// SRAStatement represents a shift right arithmetic command (sra).
type SRAStatement struct {
	// Position is the position in the source.
	Position token.Pos
	// Source is a register acting as first operand.
	Source *Register
	// Operand is the second one of the two operands.
	Operand Operand
	// Destination is the target register receiving the result of the arithmetic
	// operation.
	Destination *Register
}

// Pos returns the statements position.
func (stmt SRAStatement) Pos() token.Pos {
	return stmt.Position
}

func (stmt SRAStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("orn ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Operand.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// Expression is an expression which bundles an identifier with an offset. In
// ARC an expression is delimited by an opening and a closing square bracket.
type Expression struct {
	Position token.Pos
	// Base is the register or identifer used as base in the expression.
	Base ExpressionBase
	// Operator is the operator which is used in the expression.
	Operator string
	// Offset is the second operand.
	Offset Integer
}

// Pos returns the statements position.
func (e Expression) Pos() token.Pos {
	return e.Position
}

func (e Expression) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.WriteString(e.Base.String())
	if e.Operator != "" {
		buf.WriteString(e.Operator)
		buf.WriteString(strconv.FormatInt(int64(e.Offset), 10))
	}
	buf.WriteString("]")
	return buf.String()
}

// Identifier is a named identifier.
type Identifier struct {
	// Position is the tokens position in the source.
	Position token.Pos
	// Name is the name of the identifier.
	Name string
}

// Pos returns the statements position.
func (i Identifier) Pos() token.Pos {
	return i.Position
}

func (i Identifier) String() string {
	return i.Name
}

// Register is an ARC Register.
type Register struct {
	// Name is the name/identifier of the register.
	Name string
}

func (r Register) String() string {
	return r.Name
}

// Integer represents a 32 bit integer value.
type Integer int32

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

/*
// AddressingMode is the addressing mode used by memory operations.
type AddressingMode int

func (am AddressingMode) String() string {
	switch am {
	case Direct:
		return "DIRECT"
	case Indirect:
		return "INDIRECT"
	case Offset:
		return "OFFSET"
	}
	return ""
}

const (
	// Direct loads directly from a memory location to a source register.
	Direct = iota
	// Indirect loads the effective address of the source operand from a
	// register.
	Indirect
	// Offset adds the specified value to the address of the source register to
	// determine the final memory address.
	Offset
)
*/

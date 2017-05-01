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
	Tok() token.Token
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
func (*BEStatement) stmt()      {}
func (*BNEStatement) stmt()     {}
func (*BNEGStatement) stmt()    {}
func (*BPOSStatement) stmt()    {}
func (*BAStatement) stmt()      {}

// Reference is implemented by types which can be referenced by a label. These
// are statements and identifiers.
type Reference interface {
	// ref is unexported to ensure implementations of Reference can only
	// originate in this package.
	ref()
	String() string
}

func (*Integer) ref()        {}
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

func (*Integer) op()  {}
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
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Text is the actual text of the comment.
	Text string
}

// Pos returns the statements position.
func (stmt CommentStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt CommentStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt CommentStatement) String() string {
	return "! " + strings.TrimSpace(stmt.Text[1:])
}

// BeginStatement marks the beginning of an ARC program.
type BeginStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos
}

// Tok returns the statements lexical token.
func (stmt BeginStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos
}

// Pos returns the statements position.
func (stmt EndStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt EndStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt EndStatement) String() string {
	return ".end"
}

// OrgStatement marks a new section of data in memory.
type OrgStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Value is the memory location.
	Value *Integer
}

// Pos returns the statements position.
func (stmt OrgStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt OrgStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt OrgStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString(".org ")
	buf.WriteString(stmt.Value.String())
	return buf.String()
}

// LabelStatement represents a label.
type LabelStatement struct {
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt LabelStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt LoadStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt StoreStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt AddStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt AddCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt SubStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt SubCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt AndStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt AndCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt OrStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt OrCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt OrnStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt OrnCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt XorStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt XorCCStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt SLLStatement) Tok() token.Token {
	return stmt.Token
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
	// Token is the statements lexical token.
	Token token.Token
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

// Tok returns the statements lexical token.
func (stmt SRAStatement) Tok() token.Token {
	return stmt.Token
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

// BEStatement represents a "branch on equal to zero" command (be).
type BEStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Target is the label the branch statement will go to.
	Target *Identifier
}

// Pos returns the statements position.
func (stmt BEStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt BEStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt BEStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("be ")
	buf.WriteString(stmt.Target.String())
	return buf.String()
}

// BNEStatement represents a "branch on not equal" command (bne).
type BNEStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Target is the label the branch statement will go to.
	Target *Identifier
}

// Pos returns the statements position.
func (stmt BNEStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt BNEStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt BNEStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("bne ")
	buf.WriteString(stmt.Target.String())
	return buf.String()
}

// BNEGStatement represents a "branch on negative" command (bneg).
type BNEGStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Target is the label the branch statement will go to.
	Target *Identifier
}

// Pos returns the statements position.
func (stmt BNEGStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt BNEGStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt BNEGStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("bneg ")
	buf.WriteString(stmt.Target.String())
	return buf.String()
}

// BPOSStatement represents a "branch on positive" command (bpos).
type BPOSStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Target is the label the branch statement will go to.
	Target *Identifier
}

// Pos returns the statements position.
func (stmt BPOSStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt BPOSStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt BPOSStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("bpos ")
	buf.WriteString(stmt.Target.String())
	return buf.String()
}

// BAStatement represents a "branch always" command (ba).
type BAStatement struct {
	// Token is the statements lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Target is the label the branch statement will go to.
	Target *Identifier
}

// Pos returns the statements position.
func (stmt BAStatement) Pos() token.Pos {
	return stmt.Position
}

// Tok returns the statements lexical token.
func (stmt BAStatement) Tok() token.Token {
	return stmt.Token
}

func (stmt BAStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("ba ")
	buf.WriteString(stmt.Target.String())
	return buf.String()
}

// Expression is an expression which bundles an identifier with an offset. In
// ARC an expression is delimited by an opening and a closing square bracket.
type Expression struct {
	// Position is the position in the source.
	Position token.Pos

	// Base is the register or identifer used as base in the expression.
	Base ExpressionBase
	// Operator is the operator which is used in the expression.
	Operator string
	// Offset is the second operand.
	Offset *Integer
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
		buf.WriteString(strconv.FormatInt(int64(e.Offset.Value), 10))
	}
	buf.WriteString("]")
	return buf.String()
}

// Identifier is a named identifier.
type Identifier struct {
	// Token is the identifiers lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Name is the name of the identifier.
	Name string
}

// Pos returns the statements position.
func (i Identifier) Pos() token.Pos {
	return i.Position
}

// Tok returns the identifiers lexical token.
func (i Identifier) Tok() token.Token {
	return i.Token
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

// Integer represents a 32 bit integer.
type Integer struct {
	// Token is the identifiers lexical token.
	Token token.Token
	// Position is the position in the source.
	Position token.Pos

	// Literal is the string representation of the value (hex, oct, dec).
	Literal string
	// Value is the actual 32 bit integer value.
	Value int32
}

func (i Integer) String() string {
	// We return the literal representation to preserve the format.
	return i.Literal
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

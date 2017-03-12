package ast

import (
	"bytes"
	"strconv"
	"strings"
)

// Statement is an ARC assembly statement.
type Statement interface {
	// stmt is unexported to ensure implementations of Statement can only
	// originate in this package.
	stmt()
	String() string
}

func (*LabelStatement) stmt() {}
func (*LoadStatement) stmt()  {}
func (*StoreStatement) stmt() {}

// Reference is implemented by types which can be referenced by a label. These
// are statements and identifiers.
type Reference interface {
	// ref is unexported to ensure implementations of Reference can only
	// originate in this package.
	ref()
	String() string
}

func (*LoadStatement) ref()  {}
func (*StoreStatement) ref() {}
func (*Identifier) ref()     {}
func (Integer) ref()         {}

// MemoryLocation is implemented by types which can be addressed as locations in
// memory. A single identifier can be addressed as well as expressions.
type MemoryLocation interface {
	memLoc()
	String() string
}

func (*Expression) memLoc() {}
func (*Identifier) memLoc() {}

// Statements is a list of statements.
type Statements []Statement

// String returns a string representation of the statements.
func (s Statements) String() string {
	var str []string
	for _, stmt := range s {
		str = append(str, stmt.String())
	}
	return strings.Join(str, "\n")
}

// Program represents a collection of statements.
type Program struct {
	Statements Statements
}

// String returns a string representation of the program.
func (p Program) String() string { return p.Statements.String() }

// LabelStatement represents a label.
type LabelStatement struct {
	// Ident is the labels identifier.
	Ident *Identifier
	// Reference is an Identifier, Integer or the Statement, the label addresses.
	Reference Reference
}

func (stmt LabelStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString(stmt.Ident.Value)
	buf.WriteString(": ")
	buf.WriteString(stmt.Reference.String())
	return buf.String()
}

// LoadStatement represents a load command (ld).
type LoadStatement struct {
	// Source is the memory location where the value is loaded from.
	Source MemoryLocation
	// Destination is the register where the value is loaded to.
	Destination *Identifier
}

// String returns a string representation of the statement.
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
	// Source is the register where the value is stored from.
	Source *Identifier
	// Destination is the memory location where the value is stored to.
	Destination MemoryLocation
}

// String returns a string representation of the statement.
func (stmt StoreStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("st ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(", ")
	buf.WriteString(stmt.Destination.String())
	return buf.String()
}

// Expression is an expression which bundles an identifier with an offset. In
// ARC an expression is delimited by an opening and a closing square bracket.
type Expression struct {
	// Ident is the identifier used in the expression.
	Ident *Identifier
	// Operator is the operator which is used in the expression.
	Operator string
	// Offset is the second operand.
	Offset Integer
}

func (e Expression) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.WriteString(e.Ident.String())
	buf.WriteString(e.Operator)
	buf.WriteString(strconv.FormatInt(int64(e.Offset), 10))
	buf.WriteString("]")
	return buf.String()
}

// Identifier is a named identifier.
type Identifier struct {
	// Value is the name of the identifier.
	Value string
}

// String returns a string representation of the identifier.
func (i Identifier) String() string {
	return i.Value
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

package ast

import (
	"bytes"
	"strconv"
	"strings"
)

// Node is a node in the ARC abstract syntax tree.
type Node interface {
	// node is unexported to ensure implementations of Node can only originate
	// in this package.
	node()
	String() string
}

func (*Program) node()   {}
func (Statements) node() {}

func (*LoadStatement) node() {}

// Statement is an ARC assembly statement.
type Statement interface {
	Node
	// stmt is unexported to ensure implementations of Statement can only
	// originate in this package.
	stmt()
}

func (*LoadStatement) stmt() {}

// Statements is a list of statements.
type Statements []Statement

// String returns a string representation of the statements.
func (s Statements) String() string {
	var str []string
	for _, stmt := range s {
		str = append(str, stmt.String())
	}
	return strings.Join(str, ";\n")
}

// Program represents a collection of statements.
type Program struct {
	Statements Statements
}

// String returns a string representation of the program.
func (p Program) String() string { return p.Statements.String() }

// Comment represents a comment in the ARC assembly language.
type Comment struct {
	Text string
}

// String returns a string representation of the comment.
func (c Comment) String() string {
	return c.Text
}

// LoadStatement represents a load command (ld).
type LoadStatement struct {
	// Source is where the value is loaded from.
	Source *MemoryLocation
	// Destination is where the value is loaded to.
	Destination *Identifier
}

// String returns a string representation of the statement.
func (stmt LoadStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("LOAD FROM ")
	buf.WriteString(stmt.Source.String())
	buf.WriteString(" TO ")
	buf.WriteString(stmt.Destination.String())
	buf.WriteString(" (" + stmt.Source.Mode.String() + ")")
	return buf.String()
}

// Identifier is an named identifier.
type Identifier struct {
	// Name is the name of the identifier, which is the tokens literal.
	Name string
}

// String returns a string representation of the identifier.
func (i Identifier) String() string {
	return i.Name
}

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

// MemoryLocation is the location of a value in the memory.
type MemoryLocation struct {
	// Base specifies a location in memory. It can also be a register containing
	// a memory address.
	Base *Identifier
	// Operator is the operator which is used to determine the final memory
	// location.
	Operator string
	// Offset is the offset from the base location used to determine the final
	// memory location.
	Offset uint64
	// Mode is the addressing mode used.
	Mode AddressingMode
}

// String returns a string representation of the MemoryLocation.
func (ml MemoryLocation) String() string {
	if ml.Mode == Direct {
		return "[" + ml.Base.String() + "]"
	} else if ml.Mode == Offset {
		return "[" + ml.Base.String() + ml.Operator + strconv.FormatUint(ml.Offset, 10) + "]"
	}
	return ml.Base.String()
}

package check

import (
	"fmt"

	"github.com/lukasmalkmus/arc/ast"
)

// Ineffassign checks if there are any useless "zero offsets" ([%r1 + 0]).
type Ineffassign struct {
	name string
}

func init() {
	Register(&Ineffassign{"ineffassign"})
}

// Desc returns a description of the Check.
func (c Ineffassign) Desc() string {
	return "searches unused declarations"
}

// Name returns the name of the Check.
func (c Ineffassign) Name() string {
	return c.name
}

// Run executes the Check. It implements the Check interface.
func (c *Ineffassign) Run(prog *ast.Program) ([]string, error) {
	var (
		res    []string
		idents []*ast.Identifier
		labels []*ast.LabelStatement
	)

	for _, stmt := range prog.Statements {
		ids, lbs := extractIdentLabel(stmt)
		idents = append(idents, ids...)
		labels = append(labels, lbs...)
	}

	// See if labels are declared but their identifiers never used.
	for _, label := range labels {
		has := false
		for _, ident := range idents {
			if label.Ident.Name == ident.Name {
				has = true
				break
			}
		}
		if !has {
			msg := buildMsg(c, label.Pos(), fmt.Sprintf("%q declared but not used", label.Ident))
			res = append(res, msg)
		}
	}

	return res, nil
}

func extractIdentLabel(stmt ast.Statement) ([]*ast.Identifier, []*ast.LabelStatement) {
	idents := []*ast.Identifier{}
	labels := []*ast.LabelStatement{}

	switch stmt.(type) {
	case *ast.LabelStatement:
		label := stmt.(*ast.LabelStatement)
		labels = append(labels, label)
		// Besides reading the label, we also need to examine the referenced
		// statement.
		if ref, valid := label.Reference.(ast.Statement); valid {
			ids, lbs := extractIdentLabel(ref)
			labels = append(labels, lbs...)
			for _, ident := range ids {
				if !has(idents, ident) {
					idents = append(idents, ident)
				}
			}
		}
	case *ast.LoadStatement:
		if ident, valid := stmt.(*ast.LoadStatement).Source.(*ast.Expression).Base.(*ast.Identifier); valid {
			if !has(idents, ident) {
				idents = append(idents, ident)
			}
		}
	case *ast.StoreStatement:
		if ident, valid := stmt.(*ast.StoreStatement).Destination.(*ast.Expression).Base.(*ast.Identifier); valid {
			if !has(idents, ident) {
				idents = append(idents, ident)
			}
		}
	case *ast.BEStatement:
		ident := stmt.(*ast.BEStatement).Target
		if !has(idents, ident) {
			idents = append(idents, ident)
		}
	case *ast.BNEStatement:
		ident := stmt.(*ast.BNEStatement).Target
		if !has(idents, ident) {
			idents = append(idents, ident)
		}
	case *ast.BNEGStatement:
		ident := stmt.(*ast.BNEGStatement).Target
		if !has(idents, ident) {
			idents = append(idents, ident)
		}
	case *ast.BPOSStatement:
		ident := stmt.(*ast.BPOSStatement).Target
		if !has(idents, ident) {
			idents = append(idents, ident)
		}
	case *ast.BAStatement:
		ident := stmt.(*ast.BAStatement).Target
		if !has(idents, ident) {
			idents = append(idents, ident)
		}
	}

	return idents, labels
}

func has(idents []*ast.Identifier, ident *ast.Identifier) bool {
	for _, val := range idents {
		if ident == val {
			return true
		}
	}
	return false
}

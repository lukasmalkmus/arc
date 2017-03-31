package check

import (
	"fmt"

	"github.com/LukasMa/arc/ast"
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
		switch stmt.(type) {
		case *ast.LabelStatement:
			labels = append(labels, stmt.(*ast.LabelStatement))
		case *ast.LoadStatement:
			if ident, valid := stmt.(*ast.LoadStatement).Source.(*ast.Expression).Base.(*ast.Identifier); valid {
				if !has(ident, idents) {
					idents = append(idents, ident)
				}
			}
		case *ast.StoreStatement:
			if ident, valid := stmt.(*ast.StoreStatement).Destination.(*ast.Expression).Base.(*ast.Identifier); valid {
				if !has(ident, idents) {
					idents = append(idents, ident)
				}
			}
		}
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

func has(ident *ast.Identifier, idents []*ast.Identifier) bool {
	for _, val := range idents {
		if ident == val {
			return true
		}
	}
	return false
}

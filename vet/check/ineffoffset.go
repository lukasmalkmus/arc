package check

import (
	"fmt"

	"github.com/lukasmalkmus/arc/ast"
)

// Ineffoffset checks if there are any useless "zero offsets" ([%r1 + 0]).
type Ineffoffset struct {
	name string
}

func init() {
	Register(&Ineffoffset{"ineffoffset"})
}

// Desc returns a description of the Check.
func (c Ineffoffset) Desc() string {
	return "checks for useless \"zero offsets\" ([%r1 + 0])"
}

// Name returns the name of the Check.
func (c Ineffoffset) Name() string {
	return c.name
}

// Run executes the Check. It implements the Check interface.
func (c *Ineffoffset) Run(prog *ast.Program) ([]string, error) {
	var (
		res  []string
		exps []*ast.Expression
	)

	for _, stmt := range prog.Statements {
		exs := extractExpression(stmt)
		exps = append(exps, exs...)
	}

	// See if expressions with a zero offset are defined.
	for _, exp := range exps {
		if exp.Operator != "" && exp.Offset == 0 {
			improvedExp := &ast.Expression{Base: exp.Base}
			msg := buildMsg(c, exp.Pos(), fmt.Sprintf("offset expression %q can be shortened to %q", exp, improvedExp))
			res = append(res, msg)
		}
	}

	return res, nil
}

func extractExpression(stmt ast.Statement) []*ast.Expression {
	exps := []*ast.Expression{}

	switch stmt.(type) {
	case *ast.LabelStatement:
		// We also need to examine the referenced statement.
		label := stmt.(*ast.LabelStatement)
		ref, valid := label.Reference.(ast.Statement)
		if valid {
			exp := extractExpression(ref)
			exps = append(exps, exp...)
		}
	case *ast.LoadStatement:
		if exp, valid := stmt.(*ast.LoadStatement).Source.(*ast.Expression); valid {
			exps = append(exps, exp)
		}
	case *ast.StoreStatement:
		if exp, valid := stmt.(*ast.StoreStatement).Destination.(*ast.Expression); valid {
			exps = append(exps, exp)
		}
	}

	return exps
}

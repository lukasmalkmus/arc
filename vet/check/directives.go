package check

import (
	"fmt"

	"github.com/lukasmalkmus/arc/ast"
)

// Directives checks if there are any statements outside the .begin and .end
// directives.
type Directives struct {
	name string
}

func init() {
	Register(&Directives{"directives"})
}

// Desc returns a description of the Check.
func (c Directives) Desc() string {
	return "checks if directives are set and used correctly"
}

// Name returns the name of the Check.
func (c Directives) Name() string {
	return c.name
}

// Run executes the Check. It implements the Check interface.
func (c *Directives) Run(prog *ast.Program) ([]string, error) {
	res := []string{}

	// Run different checks.
	res = append(res, c.checkOrder(prog)...)

	return res, nil
}

// checkBeginEndOrder ensures begin, end and org are not missing and in the
// correct order.
func (c *Directives) checkOrder(prog *ast.Program) []string {
	var (
		res       []string
		beginStmt *ast.BeginStatement
		endStmt   *ast.EndStatement
		orgStmts  []*ast.OrgStatement
	)

	for _, stmt := range prog.Statements {
		switch stmt.(type) {
		case *ast.BeginStatement:
			if beginStmt != nil {
				msg := buildMsg(c, stmt.Pos(), fmt.Sprintf("duplicate .begin: first one at %s", beginStmt.Pos().NoFile()))
				res = append(res, msg)
				continue
			}
			beginStmt = stmt.(*ast.BeginStatement)
		case *ast.EndStatement:
			if endStmt != nil {
				msg := buildMsg(c, stmt.Pos(), fmt.Sprintf("duplicate .end: first one at %s", endStmt.Pos().NoFile()))
				res = append(res, msg)
				continue
			}
			endStmt = stmt.(*ast.EndStatement)
		case *ast.OrgStatement:
			if endStmt != nil {
				msg := buildMsg(c, stmt.Pos(), ".org after .end")
				res = append(res, msg)
				continue
			}
			orgStmt := stmt.(*ast.OrgStatement)
			orgStmts = append(orgStmts, orgStmt)
			if len(orgStmts) > 1 {
				if prev := orgStmts[len(orgStmts)-2]; prev.Value.Value >= orgStmt.Value.Value {
					msg := buildMsg(c, stmt.Pos(), fmt.Sprintf(".org memory address %d must be greater than address %d of .org at %s", orgStmt.Value.Value, prev.Value.Value, prev.Pos().NoFile()))
					res = append(res, msg)
				}
			}
		case *ast.CommentStatement:
			// nop
		default:
			if beginStmt == nil {
				msg := buildMsg(c, stmt.Pos(), "statement before .begin")
				res = append(res, msg)
			}
			if endStmt != nil {
				msg := buildMsg(c, stmt.Pos(), "statement after .end")
				res = append(res, msg)
			}
		}

		// Catch wrong order.
		if beginStmt == nil && endStmt != nil {
			msg := buildMsg(c, stmt.Pos(), ".end before .begin")
			res = append(res, msg)
		}
		if beginStmt == nil && len(orgStmts) > 0 {
			msg := buildMsg(c, stmt.Pos(), ".org before .begin")
			res = append(res, msg)
		}
	}

	// Catch missing directives.
	if beginStmt == nil {
		msg := buildMsg(c, prog.Filename, "missing .begin")
		res = append(res, msg)
	}
	if endStmt == nil {
		msg := buildMsg(c, prog.Filename, "missing .end")
		res = append(res, msg)
	}

	// Check if .org directive is correct.
	if len(orgStmts) == 0 {
		msg := buildMsg(c, prog.Filename, "missing .org: program code should start at address 2048")
		res = append(res, msg)
	} else if org := orgStmts[0]; org.Value.Value != 2048 {
		msg := buildMsg(c, org.Pos(), fmt.Sprintf("program code should start at address 2048, not %d", org.Value.Value))
		res = append(res, msg)
	}

	return res
}

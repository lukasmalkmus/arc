package internal

import "github.com/lukasmalkmus/arc/ast"

// StatementName returns a human-friendly name of the given statement. An empty
// string is returned if the string can't be resolved.
func StatementName(stmt ast.Statement) string {
	switch stmt.(type) {
	case *ast.CommentStatement:
		return "COMMENT"
	case *ast.BeginStatement:
		return "BEGIN"
	case *ast.EndStatement:
		return "END"
	case *ast.OrgStatement:
		return "ORG"
	case *ast.LabelStatement:
		return "LABEL"
	case *ast.LoadStatement:
		return "LOAD"
	case *ast.StoreStatement:
		return "STORE"
	default:
		return ""
	}
}

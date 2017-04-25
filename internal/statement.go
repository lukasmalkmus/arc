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
	case *ast.AddStatement:
		return "ADD"
	case *ast.AddCCStatement:
		return "ADDCC"
	case *ast.SubStatement:
		return "SUB"
	case *ast.SubCCStatement:
		return "SUBCC"
	case *ast.AndStatement:
		return "AND"
	case *ast.AndCCStatement:
		return "ANDCC"
	case *ast.OrStatement:
		return "OR"
	case *ast.OrCCStatement:
		return "ORCC"
	case *ast.OrnStatement:
		return "ORN"
	case *ast.OrnCCStatement:
		return "ORNCC"
	case *ast.XorStatement:
		return "XOR"
	case *ast.XorCCStatement:
		return "XORCC"
	case *ast.SLLStatement:
		return "SLL"
	case *ast.SRAStatement:
		return "SRA"
	default:
		return ""
	}
}

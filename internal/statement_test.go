package internal

import (
	"testing"

	"github.com/lukasmalkmus/arc/ast"
)

func TestStatementName(t *testing.T) {
	tests := []struct {
		stmt ast.Statement
		str  string
	}{
		{stmt: &ast.CommentStatement{}, str: "COMMENT"},
		{stmt: &ast.BeginStatement{}, str: "BEGIN"},
		{stmt: &ast.EndStatement{}, str: "END"},
		{stmt: &ast.OrgStatement{}, str: "ORG"},
		{stmt: &ast.LabelStatement{}, str: "LABEL"},
		{stmt: &ast.LoadStatement{}, str: "LOAD"},
		{stmt: &ast.StoreStatement{}, str: "STORE"},
		{stmt: &ast.AddStatement{}, str: "ADD"},
		{stmt: &ast.AddCCStatement{}, str: "ADDCC"},
		{stmt: &ast.SubStatement{}, str: "SUB"},
		{stmt: &ast.SubCCStatement{}, str: "SUBCC"},
		{stmt: &ast.AndStatement{}, str: "AND"},
		{stmt: &ast.AndCCStatement{}, str: "ANDCC"},
		{stmt: &ast.OrStatement{}, str: "OR"},
		{stmt: &ast.OrCCStatement{}, str: "ORCC"},
		{stmt: &ast.OrnStatement{}, str: "ORN"},
		{stmt: &ast.OrnCCStatement{}, str: "ORNCC"},
		{stmt: &ast.XorStatement{}, str: "XOR"},
		{stmt: &ast.XorCCStatement{}, str: "XORCC"},
		{stmt: &ast.SLLStatement{}, str: "SLL"},
		{stmt: &ast.SRAStatement{}, str: "SRA"},
		{stmt: nil, str: ""},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			equals(t, tt.str, StatementName(tt.stmt))
		})
	}
}

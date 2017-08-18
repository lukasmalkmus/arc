package internal

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
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

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unttected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if tt is not equal to act.
func equals(tb testing.TB, tt, act interface{}) {
	if !reflect.DeepEqual(tt, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, tt, act)
		tb.FailNow()
	}
}

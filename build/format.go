package build

import "github.com/lukasmalkmus/arc/ast"

// InstructionFormats maps InstructionFormats to their respective operation code.
var InstructionFormats map[ast.Format][]byte

func init() {
	InstructionFormats = map[ast.Format][]byte{
		ast.Branch:     []byte("00"),
		ast.Sethi:      []byte("00"),
		ast.Call:       []byte("01"),
		ast.Arithmetic: []byte("10"),
		ast.Memory:     []byte("11"),
	}
}

// LookupInstructionFormat returns the instruction format for a given statement.
func LookupInstructionFormat(stmt ast.InstructionFormat) ([]byte, bool) {
	op, ok := InstructionFormats[stmt.InstructionFormat()]
	return op, ok
}

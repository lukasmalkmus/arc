package build

import (
	"github.com/lukasmalkmus/arc/ast"

	"github.com/lukasmalkmus/arc/token"
)

// OpCodes maps lexical tokens to their respective operation code.
var OpCodes map[token.Token][]byte

func init() {
	OpCodes = map[token.Token][]byte{
		token.LOAD: []byte("000000"),
	}
}

// LookupOpCode returns the operation code for a given statement.
func LookupOpCode(stmt ast.Statement) ([]byte, bool) {
	op, ok := OpCodes[stmt.Tok()]
	return op, ok
}

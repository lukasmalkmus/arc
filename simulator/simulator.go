package simulator

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/LukasMa/arc/ast"
)

// Simulator is simulating an ARC microprocessor. It executes one statement at a
// time.
type Simulator struct {
	registers map[string]Register
}

// New creates a new ARC Simulator.
func New() *Simulator {
	s := &Simulator{
		registers: make(map[string]Register),
	}

	for i := 0; i < 32; i++ {
		r := "%r" + strconv.Itoa(i)
		s.registers[r] = NewRegister()
	}

	return s
}

// Exec will parse and run the string on the Simulator.
func (s *Simulator) Exec(stmt ast.Statement) error {
	switch stmt.(type) {
	case *ast.LoadStatement:
		s.execLoadStatement(stmt.(*ast.LoadStatement))
	case *ast.StoreStatement:
		s.execStoreStatement(stmt.(*ast.StoreStatement))
	default:
		return fmt.Errorf("no logic implemented to run this type of statement")
	}

	return nil
}

// State returns a string representation of the Simulators state.
func (s Simulator) State() string {
	var buf bytes.Buffer

	for i := 0; i < 32; i++ {
		r := "%r" + strconv.Itoa(i)
		fmt.Fprintf(&buf, "%s:\t%s\n", r, s.registers[r].String())
	}

	return buf.String()
}

func (s *Simulator) execLoadStatement(stmt *ast.LoadStatement) error {
	return nil
}

func (s *Simulator) execStoreStatement(stmt *ast.StoreStatement) error {
	return nil
}

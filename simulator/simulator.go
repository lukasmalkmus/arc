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
		r := "r" + strconv.Itoa(i)
		s.registers[r] = NewRegister()
	}
	s.registers["pc"] = NewRegister()

	return s
}

// Exec will parse and run the string on the Simulator.
func (s *Simulator) Exec(stmt ast.Statement) error {
	var err error
	switch stmt.(type) {
	case *ast.LoadStatement:
		err = s.execLoadStatement(stmt.(*ast.LoadStatement))
	case *ast.StoreStatement:
		err = s.execStoreStatement(stmt.(*ast.StoreStatement))
	case *ast.LabelStatement:
		err = s.execLabelStatement(stmt.(*ast.LabelStatement))
	default:
		return fmt.Errorf("no logic implemented to run this type of statement")
	}

	return err
}

// State returns a string representation of the Simulators state.
func (s Simulator) State() string {
	var buf bytes.Buffer

	for i := 0; i < 32; i++ {
		r := "r" + strconv.Itoa(i)
		fmt.Fprintf(&buf, "%s:\t%s\n", r, s.registers[r].String())
	}
	fmt.Fprintf(&buf, "%s:\t%s\n", "pc", s.registers["pc"].String())

	return buf.String()
}

// execLoadStatement executes a ld command on the simulator.
func (s *Simulator) execLoadStatement(stmt *ast.LoadStatement) error {
	s.incPC()
	return nil
}

// execStoreStatement executes a st command on the simulator.
func (s *Simulator) execStoreStatement(stmt *ast.StoreStatement) error {
	s.incPC()
	return nil
}

// execLabelStatement executes a label command on the simulator.
func (s *Simulator) execLabelStatement(stmt *ast.LabelStatement) error {
	return nil
}

// incPC increments the simulators program counter.
func (s *Simulator) incPC() {
	s.registers["pc"] += Register(4)
}

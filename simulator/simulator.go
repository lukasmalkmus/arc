package simulator

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/lukasmalkmus/arc/ast"
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
	s.Reset()

	return s
}

// Exec will parse and run the statement on the simulator.
func (s *Simulator) Exec(stmt ast.Statement) error {
	var err error
	switch stmt.(type) {
	case *ast.LabelStatement:
		err = s.execLabelStatement(stmt.(*ast.LabelStatement))
	case *ast.LoadStatement:
		err = s.execLoadStatement(stmt.(*ast.LoadStatement))
	case *ast.StoreStatement:
		err = s.execStoreStatement(stmt.(*ast.StoreStatement))
	default:
		return fmt.Errorf("not implemented")
	}

	return err
}

// Reset resets the Simulator. This will clear all registers and memory
// allocations.
func (s *Simulator) Reset() {
	for i := 0; i < 32; i++ {
		r := "r" + strconv.Itoa(i)
		s.registers[r] = NewRegister()
	}
	s.registers["pc"] = NewRegister()
}

// State returns a string representation of the Simulators state.
func (s Simulator) State() string {
	var buf bytes.Buffer

	for i := 0; i < 32; i++ {
		r := "r" + strconv.Itoa(i)
		fmt.Fprintf(&buf, "%s:\t%s\n", r, s.registers[r].Hex())
	}
	fmt.Fprintf(&buf, "%s:\t%s\n", "pc", s.registers["pc"].Hex())

	return buf.String()
}

// Usage returns a usage string.
func (s Simulator) Usage() string {
	return "Usage"
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

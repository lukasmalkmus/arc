package simulator

import (
	"bytes"
	"fmt"
)

// A Register is 32bit wide register.
type Register [4]byte

// NewRegister creates a new Register.
func NewRegister() Register {
	return Register([4]byte{0x00, 0x00, 0x00, 0x00})
}

// String implements the Stringer interface and returns a string representation
// of the Register.
func (r Register) String() string {
	var buf bytes.Buffer
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&buf, "%.8b ", r[i])
	}
	buf.WriteString(" (")
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&buf, "0x%.2X ", r[i])
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(")")
	return buf.String()
}

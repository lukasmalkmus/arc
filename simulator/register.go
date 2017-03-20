package simulator

import (
	"bytes"
	"fmt"
)

// A Register is 32bit wide register.
type Register int32

// NewRegister creates a new Register.
func NewRegister() Register {
	return 0
}

// String implements the Stringer interface and returns a string representation
// of the Register.
func (r Register) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s (%s)", r.Bin(), r.Hex())
	return buf.String()
}

// Bin returns the binary representation of the registers content.
func (r Register) Bin() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%.32b", r)
	return buf.String()
}

// Hex returns the hexadecimal representation of the registers content.
func (r Register) Hex() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "0x%.8X", int32(r))
	return buf.String()
}

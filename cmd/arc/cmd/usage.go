package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const usage = `The following instructions are supported:

Directives:
".begin"
".end"
".org"

Memory:
"ld"
"st"

Arithmetic:
"add", "addcc"
"sub", "subcc"

Logic:
"and", "andcc"
"or", "orcc"
"orn", "orncc"
"xor", "xorcc"
"sll", "sra"

Control:
"be"
"bne"
"bneg"
"bpos"
"ba"

Subroutine:
"call"
"jmpl"


"ld": Load a register from main memory. The memory address
must be aligned on a word boundary.
Example usages: ld [x], %r1. Meaning: Copy the contents of
memory location x into register %r1.

"st": Store a register into main memory. The memory address
must be aligned on a word boundary.
Example usages: st %r1, [x]. Meaning: Store the contents of
%r1 into the memory location x.

"add", "addcc": Adds the source operands into the
destination register using two's complement arithmetic.
"addcc" sets the condition codes according to the result.
Example usage: add %r1, %r2, %r4. Meaning: %r4 = %r1 + %r2
Example usage: addcc %r1, 2, %r2. Meaning: %r2 = %r2 + 2

"sub", "subcc": Perform integer subtraction on the source
operands and put result into the destination register using
two's complement arithmetic. "subcc" sets the condition
codes according to the result.
Example usage: sub %r1, %r2, %r4. Meaning: %r4 = %r1 - %r2
Example usage: subcc %r1, 2, %r2. Meaning: %r2 = %r2 - 2

"and", "andcc": Bitwise AND the source operands into the
destination register. "andcc" sets the N and Z condition
codes according to the result.
Example usage: and %r1, %r2, %r4. Meaning: %r4 = %r1 AND %r2
Example usage: andcc %r1, 2, %r4. Meaning: %r4 = %r1 AND 2

"or", "orcc": Bitwise OR the source operands into the
destination register. "orcc" sets the N and Z condition
codes according to the result.
Example usage: or %r1, %r2, %r4. Meaning: %r4 = %r1 AND %r2
Example usage: orcc %r1, 1, %r2. Meaning: %r2 = %r1 OR 1

"orn", "orncc": Bitwise NOR the source operands into the
destination register. "orncc" will set the N and Z condition
codes according to the result.
Example usage: orncc %r1, %r0, %r1. Meaning: Complement %r1.

"xor", "xorcc": Bitwise XOR (exclusive OR) the source
operands into the destination register. "xorcc" will set
the N and Z condition codes according to the result.
Example usage: xorcc %r1, %r0, %r1. Meaning: %r1 = %r1 XOR %r0.

"sll": Shifts a register to the left by 0-31 bits. The vacant
bit positions in the right side of the shifted register are
filled with 0's.
Example usage: sll %r1, 3, %r4. Meaning: Shift %r1 left by
3 bits and store in %r4.
Example usage: sll %r1, %r4, %r5. Meaning: Shift %r1 left
by the value stored in %r4 and store in %r5.

"sra": Shifts a register to the right by 0-31 bits. The
sign bit is replicated as the value is shifted right.
Example usage: sra %r1, 3, %r4. Meaning:  Shift %r1 right
by 3 bits and store in %r4.
Example usage: sra %r1, %r4, %r5. Meaning: Shift %r1 right
by the value stored in %r4 and store in %r5.

"be": Branch on equal to zero. If the z condition code is 1,
then branch to the address represented by the label which is
the instruction operand.
Example usage: be label. Meaning: Branch to label if Z is 1.

"bne": Branch on not equal. Branch if not equal to zero to the
address represented by the label which is the instruction
operand.
Example usage: bne label. Meaning: Branch to label if not equal
to zero.

"bneg": Branch on negative. If the n condition code is 1, then
branch to the address represented by the label which is the
instruction operand.
Example usage: bneg label. Meaning: Branch to label if N is 1.

"bne": Branch on not equal. Branch if not equal to zero to the
address represented by the label which is the instruction
operand.
Example usage: bne label. Meaning: Branch to label if not equal
to zero.

"bpos": Branch on positive. If the condition codes signal a
positive result, branch to the address represented by the label
which is the instruction operand.
Example usage: bpos label. Meaning: Branch if positive.

"ba": Branch always. Always branch to the address represented
by the label which is the instruction operand.
Example usage: ba label. Meaning: Always branch to label.

"call": Call a subroutine and store the address of the current
instruction in %r15. The instruction operand is the address of
the subroutine and is stored as a 30 bit displacement in the
call instruction format.
Example usage: call sub_r. Meaning: Call the subroutine located
at sub_r.

"jmpl": Unconditional, register indirect control transfer. Jump
to a new address and store the address of the current
instruction in the destination register.
Example usage: jmpl %r15 + 4, %r2. Meaning: Set the program
counter to the contents of %r15 + 4. The current address is
stored into %r2.


The following pseudo-operations are supported:

.begin, .end: Start and stop assembly, respectively.

.org: Change location counter to the address specifed.
Example usage: .org 0x800. The next instruction will be
assembled at location 0x800 (2048).

NOTE: A pseudo-operation should NOT be followed by a colon
because it is not a program label.

Comments begin with the "!" symbol and continue to the end of
line.
Example usage: ba exit ! Go to program exit
`

// usageCmd represents the usage command
var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show advanced usage information and supported instructions",
	Long: `Show detailed information about the ARC assembly language and
supported ARC instructions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(usage)
	},
}

func init() {
	RootCmd.AddCommand(usageCmd)
}

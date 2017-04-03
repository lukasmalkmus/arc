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


"ld": load a register from main memory. The memory address
must be aligned on a word boundary.
Example usages: ld [x], %r1
Meaning: Copy the contents of memory location x into
register %r1.

"st": store a register into main memory. The memory address
must be aligned on a word boundary.
Example usages: st %r1, [x]
Meaning: Store the contents of %r1 into the memory location
x.

"add", "addcc": adds the source operands into the
destination register using two's complement arithmetic.
"addcc" sets the condition codes according to the result.
Example usage: add %r1, %r2, %r4. Meaning: %r4 = %r1 + %r2
Example usage: addcc %r1, 2, %r2. Meaning: %r2 = %r2 + 2

"sub", "subcc": perform integer subtraction on the source
operands and put result into the destination register using
two's complement arithmetic. "subcc" sets the condition
codes according to the result.
Example usage: sub %r1, %r2, %r4. Meaning: %r4 = %r1 - %r2
Example usage: subcc %r1, 2, %r2. Meaning: %r2 = %r2 - 2

"and", "andcc": bitwise AND the source operands into the
destination register. "andcc" sets the N and Z condition
codes according to the result.
Example usage: and %r1, %r2, %r4. Meaning: %r4 = %r1 AND %r2
Example usage: andcc %r1, 2, %r4. Meaning: %r4 = %r1 AND 2

"or", "orcc": bitwise OR the source operands into the
destination register. "orcc" sets the N and Z condition
codes according to the result.
Example usage: or %r1, %r2, %r4. Meaning: %r4 = %r1 AND %r2
Example usage: orcc %r1, 1, %r2. Meaning: %r2 = %r1 OR 1

"orn", "orncc": bitwise NOR the source operands into the
destination register. "orncc" will set the N and Z condition
codes according to the result.
Example usage: orncc %r1, %r0, %r1.  Meaning: Complement %r1.

"xor", "xorcc": bitwise XOR (exclusive OR) the source
operands into the destination register. "xorcc" will set
the N and Z condition codes according to the result.
Example usage: xorcc %r1, %r0, %r1.  Meaning: %r1 = %r1 XOR %r0.
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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// vetOptions are options for the vet command.
type vetOptions struct {
	fix bool
}

var vetOpts vetOptions

// vetCmd represents the vet command.
var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Examine ARC source code for suspicious constructs",
	Long: `Vet examines ARC source code and reports suspicious language
constructs, such as zero offset operations ([x+0]). It uses
heuristics that do not guarantee all reports are genuine
problems.

Every argument to this command is expected to be a valid
ARC source file. Passing no argument will vet every single
file having the .arc file extension in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here!
		fmt.Println("vet called")
	},
	SuggestFor: []string{"check"},
}

func init() {
	RootCmd.AddCommand(vetCmd)

	vetCmd.Flags().BoolVarP(&vetOpts.fix, "fix", "f", false, "Apply fixes to source code")
}

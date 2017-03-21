package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// vetCmd represents the vet command
var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Examine ARC source code and report suspicious constructs",
	Long: `Vet examines ARC source code and reports suspicious language
constructs, such as zero offset operations ([x+0]=>[x]).
Vet uses heuristics that do not guarantee all reports are
genuine problems.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here!
		fmt.Println("vet called")
	},
}

func init() {
	RootCmd.AddCommand(vetCmd)
}

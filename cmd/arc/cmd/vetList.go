package cmd

import (
	"fmt"

	"github.com/LukasMa/arc/vet/check"
	"github.com/spf13/cobra"
)

// vetListCmd represents the vetList command.
var vetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available checks",
	Long:  `Print a list of checks which are available for the vet command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Available checks:\n\n")
		for _, v := range check.List() {
			fmt.Printf("\t%s\n", v)
		}
		fmt.Println()
	},
}

func init() {
	vetCmd.AddCommand(vetListCmd)
}

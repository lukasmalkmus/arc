package cmd

import (
	"fmt"

	"github.com/LukasMa/arc/internal"
	"github.com/LukasMa/arc/vet"
	"github.com/LukasMa/arc/vet/check"
	"github.com/spf13/cobra"
)

var vetOpts vet.Options
var list bool

// vetCmd represents the vet command.
var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Examine ARC source code for suspicious constructs",
	Long: `Vet examines ARC source code and reports suspicious language
constructs. It uses heuristics that do not guarantee all
reports are genuine problems.

By default all checks are run. To disable this behaviour
individual checks can be enabled by using the "--enable" flag.

Every argument to this command is expected to be a valid
ARC source file. Passing no argument will vet every single
file having the .arc file extension in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// List all checks, if requested.
		if list {
			for _, v := range check.Desc() {
				fmt.Printf("%s\n", v)
			}
			return
		}

		// Vet every file given.
		if len(args) > 0 {
			for _, file := range args {
				res, err := vet.CheckFile(file, &vetOpts)
				if err != nil {
					fmt.Println(err)
				}
				printVetResult(res)
			}
			return
		}

		// Read all files in current directory and vet them.
		files, err := internal.ReadCurDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, file := range files {
			res, err := vet.CheckFile(file, &vetOpts)
			if err != nil {
				fmt.Println(err)
			}
			printVetResult(res)
		}
	},
	SuggestFor: []string{"check"},
}

func printVetResult(res []string) {
	if len(res) == 0 {
		return
	}

	for _, msg := range res {
		fmt.Printf("%s\n", msg)
	}
}

func init() {
	RootCmd.AddCommand(vetCmd)

	vetCmd.Flags().BoolVarP(&list, "list", "l", false, "List available checks")
	vetCmd.Flags().StringSliceVar(&vetOpts.Checks, "enable", []string{}, "Enable a specific check")
}

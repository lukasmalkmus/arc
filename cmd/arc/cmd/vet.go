package cmd

import (
	"fmt"

	"github.com/lukasmalkmus/arc/internal"
	"github.com/lukasmalkmus/arc/vet"
	"github.com/lukasmalkmus/arc/vet/check"
	"github.com/spf13/cobra"
)

var vetOpts vet.Options
var list bool

// vetCmd represents the vet command.
var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Examine ARC source code for suspicious language constructs",
	Long: `Vet examines ARC source code and reports suspicious language
constructs. It uses heuristics that do not guarantee all
reports are genuine problems.

By default all checks are run. To disable this behaviour
individual checks can be enabled by using the "--enable" flag.

The "--sort" ("-s") flag can be used to sort the results
according to the source code position they apply to. By
default, results are ordered after the execution order of
the different checks.

Every argument to this command is expected to be a valid
ARC source file. Passing no argument will vet every single
file in the current directory having the .arc file extension.`,
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
				// If an argument is a directory, ignore it.
				if is, _ := internal.IsDirectory(file); is {
					continue
				}

				res, err := vet.CheckFile(file, &vetOpts)
				if err != nil {
					printError(err)
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
				printError(err)
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

func printError(err error) {
	fmt.Printf("\033[31m%s\033[39m\n", err)
}

func init() {
	RootCmd.AddCommand(vetCmd)

	vetCmd.Flags().BoolVarP(&list, "list", "l", false, "list available checks")
	vetCmd.Flags().BoolVarP(&vetOpts.Sort, "sort", "s", false, "sort results according to the source code position they apply to")
	vetCmd.Flags().StringSliceVar(&vetOpts.Checks, "enable", []string{}, "enable a specific check")
}

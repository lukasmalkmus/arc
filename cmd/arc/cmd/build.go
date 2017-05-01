package cmd

import (
	"fmt"

	"github.com/lukasmalkmus/arc/build"
	"github.com/lukasmalkmus/arc/internal"
	"github.com/spf13/cobra"
)

var buildOpts build.Options

// buildCmd represents the build command.
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Assemble ARC source code",
	Long: `Build assembles ARC source code. A valid ARC program must
be enclosed by the .begin and .end directives. By
convenience the program code should start at memory
location 2048. Consider using the .org directive for this.

Every argument to this command is expected to be a valid
ARC source file. Passing no argument will assemble every
single file having the .arc file extension in the current
directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WIP! NOT WOKRING YET!")

		// Assemble every file given.
		if len(args) > 0 {
			for _, file := range args {
				// If an argument is a directory, ignore it.
				if is, _ := internal.IsDirectory(file); is {
					continue
				}

				if err := build.AssembleFile(file, &buildOpts); err != nil {
					fmt.Printf("\033[31m%s\033[39m\n", err)
				}
			}
			return
		}

		// Read all files in current directory and assemble them.
		files, err := internal.ReadCurDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, file := range files {
			if err := build.AssembleFile(file, &buildOpts); err != nil {
				fmt.Printf("\033[31m%s\033[39m\n", err)
			}
		}
	},
	SuggestFor: []string{"assemble", "compile"},
}

func init() {
	RootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVarP(&buildOpts.Verbose, "verbose", "v", false, "Log more build details")
}

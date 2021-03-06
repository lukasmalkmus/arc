package cmd

import (
	"fmt"

	arcfmt "github.com/lukasmalkmus/arc/fmt"
	"github.com/lukasmalkmus/arc/internal"
	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command.
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format ARC source code",
	Long: `Fmt formats ARC source code according to the ARC language
specification and best practices.

Every argument to this command is expected to be a valid
ARC source file. Passing no argument will format every
single file in the current directory having the .arc file
extension.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Format every file given.
		if len(args) > 0 {
			for _, file := range args {
				// If an argument is a directory, ignore it.
				if is, _ := internal.IsDirectory(file); is {
					continue
				}

				if err := arcfmt.FormatFile(file); err != nil {
					printError(err)
				}
			}
			return
		}

		// Read all files in current directory and format them.
		files, err := internal.ReadCurDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, file := range files {
			if err := arcfmt.FormatFile(file); err != nil {
				printError(err)
			}
		}
	},
	SuggestFor: []string{"format"},
}

func init() {
	RootCmd.AddCommand(fmtCmd)
}

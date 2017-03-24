package cmd

import (
	"fmt"

	arcfmt "github.com/LukasMa/arc/fmt"
	"github.com/LukasMa/arc/util"
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
single file having the .arc file extension in the current
directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Format every file given.
		if len(args) > 0 {
			for _, file := range args {
				if err := arcfmt.FormatFile(file); err != nil {
					fmt.Println(err)
				}
			}
			return
		}

		// Read all files in current directory and format them.
		files, err := util.ReadCurDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, file := range files {
			if err := arcfmt.FormatFile(file); err != nil {
				fmt.Println(err)
			}
		}
	},
	SuggestFor: []string{"format"},
}

func init() {
	RootCmd.AddCommand(fmtCmd)
}

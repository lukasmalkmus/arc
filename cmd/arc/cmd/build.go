package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an ARC program from an ARC source file",
	Long: `A valid ARC program must be enclosed by the .begin and .end
directives. By convenience the program code should start
at memory location 2048. Consider using the .org directive
for this.

A file containing ARC source code should be have the file
extension .arc.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("build called")
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)

	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

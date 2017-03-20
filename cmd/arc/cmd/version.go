package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the version of the arc tool.
var Version = "0.1"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the arc tool",
	Long:  `Print the version of the arc tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arc version %s\n", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

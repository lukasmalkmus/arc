package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the version of the arc application.
	Version = "0.1"

	// Author is the Author of the application.
	Author = "Lukas Malkmus <mail@lukasmalkmus.com>"

	// License  is the applications license.
	License = "MIT"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of the arc tool",
	Long:  `Prints the version of the arc tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Arc version %s\n\n", Version)
		fmt.Printf("© %s\n\n", Author)
		fmt.Printf("Distributed under %s license\n", License)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

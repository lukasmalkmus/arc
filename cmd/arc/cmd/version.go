package cmd

import (
	"fmt"

	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
)

var (
	// Author is the Author of the application.
	Author = "Lukas Malkmus <mail@lukasmalkmus.com>"

	// License  is the applications license.
	License = "MIT"
)

var (
	verbose bool
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of the arc tool",
	Long:  `Prints the version of the arc tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Arc version %s\n", version.Version)
		fmt.Printf("© %s\n", Author)
		fmt.Printf("Distributed under %s license\n", License)

		if verbose {
			fmt.Println()
			fmt.Print(version.Print("arc"))
		}
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print more detailed version information")
}

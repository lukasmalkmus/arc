package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "arc",
	Short: "Arc is a tool for managing ARC assembly source code",
	Long: `Arc is a tool for managing ARC assembly source code.

It offers features like assembling, formating and vetting
of ARC source code as well as an interactive mode (repl).`,
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(0)
	}
}

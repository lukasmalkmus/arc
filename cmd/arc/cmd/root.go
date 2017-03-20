package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "arc",
	Short: "Arc is a tool for managing ARC assembly source code",
	Long: `Arc is a tool for managing ARC assembly source code.

It offers features like building an ARC program from ARC
source code, running an ARC program on the simulator and
an interactive mode.`,
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

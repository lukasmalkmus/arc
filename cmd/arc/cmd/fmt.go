package cmd

import (
	"fmt"

	arcfmt "github.com/LukasMa/arc/fmt"
	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format ARC source code",
	Long: `The ARC source file must be syntactically correct to get
formatted.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: It is only possible to format a single file at the moment!
		if len(args) != 1 {
			fmt.Println("Expected exactly one source file as argument!")
		}

		// Format file.
		if err := arcfmt.FormatFile(args[0], verbose); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(fmtCmd)
}

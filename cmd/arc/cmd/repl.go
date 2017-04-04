package cmd

import (
	"fmt"
	"strings"

	"github.com/lukasmalkmus/arc/parser"
	"github.com/lukasmalkmus/interactive"
	"github.com/spf13/cobra"
)

var print bool

// replCmd represents the repl command.
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Read Evaluate Print Loop (Interactive mode)",
	Long: `The Read Evaluate Print Loop (REPL), also known as interactive
mode, takes an input string from Stdin and tries to parse
it into an ARC statement. Successful parsing will print a
check mark. Parser errors will be printed to Stdout.`,
	Run: func(cmd *cobra.Command, args []string) {
		session := interactive.New(">")
		session.Action = func(c *interactive.Context) error {
			// Scan input.
			text, err := c.Scan()
			if err != nil {
				return fmt.Errorf("Couldn't read user input: %s", err)
			}
			text = strings.TrimSpace(text)

			// Parse actual input.
			stmt, err := parser.ParseStatement(text)
			if err != nil {
				c.Printf("\033[31m%s\033[39m\n", err)
				return nil
			}
			if print {
				c.Println(stmt)
				return nil
			}
			c.Println("✓")

			return nil
		}
		session.Run()
	},
	SuggestFor: []string{"sim", "simulate"},
}

func init() {
	RootCmd.AddCommand(replCmd)

	replCmd.Flags().BoolVarP(&print, "print", "p", false, "Print the evaluated statement")
}

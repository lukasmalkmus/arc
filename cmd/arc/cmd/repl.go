package cmd

import (
	"fmt"
	"strings"

	"github.com/lukasmalkmus/arc/parser"
	"github.com/lukasmalkmus/interactive"
	"github.com/spf13/cobra"
)

var (
	confirm bool
	print   bool
)

// replCmd represents the repl command.
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Read Evaluate Print Loop (Interactive mode)",
	Long: `The Read Evaluate Print Loop (REPL), also known as interactive
mode, takes an input string from Stdin and tries to parse
it into an ARC statement. Parser errors will be printed to
Stdout. Pseudo operations "exit" and "quit" are supported
and will stop the interactive mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Init parser.
		p := parser.New(strings.NewReader(""))

		// Create new session.
		session := interactive.New(">")
		session.Before = func(c *interactive.Context) error {
			c.Println("Welcome to the ARC REPL!")
			return nil
		}
		session.Action = func(c *interactive.Context) error {

			// Scan user input.
			text, err := c.Scan()
			if err != nil {
				return fmt.Errorf("Couldn't read user input: %s", err)
			}
			text = strings.TrimSpace(text)

			// Check if the user wants to quit.
			if s := strings.ToLower(text); s == "exit" || s == "quit" {
				c.Close()
			}

			// Parse actual input. If evaluation fails print the error. Break
			// action if no statement was parsed (but the error is nil).
			p.Feed(text)
			prog, err := p.Parse()
			if err != nil {
				c.Printf("\033[31m%s\033[39m\n", err)
				return nil
			}
			if len(prog.Statements) == 0 {
				return nil
			}

			// Print confirmation if option is set and statement was parsed
			// correctly.
			if confirm {
				c.Println("✓")
				return nil
			}

			// Print statement if option is set and statement was parsed
			// correctly..
			if print {
				c.Println(prog.Statements[0])
				return nil
			}

			return nil
		}
		session.After = func(c *interactive.Context) error {
			c.Println("See you!")
			return nil
		}

		// Run session.
		session.Run()
	},
	SuggestFor: []string{"sim", "simulate"},
}

func init() {
	RootCmd.AddCommand(replCmd)

	replCmd.Flags().BoolVarP(&confirm, "confirm", "c", false, "Print a confirmation if the statement was evaluated correctly")
	replCmd.Flags().BoolVarP(&print, "print", "p", false, "Print the evaluated statement")
}

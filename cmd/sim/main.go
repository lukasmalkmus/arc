package main

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/LukasMa/arc/parser"
	"github.com/LukasMa/arc/simulator"
	"github.com/LukasMa/interactive"
)

func main() {
	session := interactive.New(">")
	sim := simulator.New()

	session.Before = func(c *interactive.Context) error {
		// Try to get username for personal welcome message.
		user, err := user.Current()
		if err != nil {
			c.Println("Hello! This is the ARC assembly language!")
		} else {
			c.Println(fmt.Sprintf("Hello %s! This is the ARC assembly language!", user.Username))
		}
		c.Println("Use 'exit', 'quit', Ctrl+C or Ctrl+D to quit.")
		c.Println("Use 'state' to print the simulators state.")
		c.Println("Use 'reset' to reset the simulators state.")
		c.Println()

		return nil
	}

	session.Action = func(c *interactive.Context) error {
		// Scan input.
		text, err := c.Scan()
		if err != nil {
			return fmt.Errorf("Couldn't read user input: %s", err)
		}

		// Trim spaces and evaluate what the user wants to do.
		text = strings.TrimSpace(text)
		switch text {
		case "exit", "quit":
			c.Close()
			return nil
		case "state":
			c.Printf("\n%s\n", sim.State())
			return nil
		}

		// Parse actual input.
		if stmt, err := parser.ParseStatement(text); err != nil {
			c.Printf("\033[31m[PARSER]:\033[39m %s\n", err)
		} else {
			if err := sim.Exec(stmt); err != nil {
				c.Printf("\033[31m[SIMULATOR]:\033[39m %s\n", err)
			}
		}

		return nil
	}

	session.After = func(c *interactive.Context) error {
		c.Println("See you next time!")
		return nil
	}

	session.Run()
}

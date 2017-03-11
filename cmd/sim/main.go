package main

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/LukasMa/arc/parser"
	"github.com/LukasMa/interactive"
)

func main() {
	s := interactive.New(">")

	s.Before = func(c *interactive.Context) error {
		// Try to get username for personal welcome message.
		user, err := user.Current()
		if err != nil {
			c.Println("Hello! This is the ARC assembly language!")
		} else {
			c.Println(fmt.Sprintf("Hello %s! This is the ARC assembly language!", user.Username))
		}
		c.Printf("Use 'exit', 'quit', Ctrl+C, Ctrl+D to quit.\n\n")

		return nil
	}

	s.Action = func(c *interactive.Context) error {
		// Scan input.
		text, err := c.Scan()
		if err != nil {
			return fmt.Errorf("Couldn't read user input: %s", err)
		}

		// Trim spaces. Check if user wants to close the application.
		text = strings.TrimSpace(text)
		switch text {
		case "exit", "quit":
			c.Close()
		}

		// Parse actual input.
		if stmt, err := parser.ParseStatement(text); err != nil {
			c.Println(err.Error())
		} else {
			c.Println(stmt.String())
		}

		return nil
	}

	s.After = func(c *interactive.Context) error {
		c.Println("Bye!")
		return nil
	}

	s.Run()
}

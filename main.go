package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	// In urfave/cli v3, the root is typically a Command.
	cmd := &cli.Command{
		Name:    "adr",
		Usage:   "Work with Architecture Decision Records (ADRs)",
		Version: "0.1.0",
		// Flags and Commands will be set by setFlags and setCommands
		// Action for the root command if no subcommand is called (optional)
		// Action: func(ctx context.Context, cmd *cli.Command) error {
		// 	 return cli.ShowAppHelp(cmd) // Example: show help by default
		// },
	}

	setFlags(cmd)    // Pass the *cli.Command
	setCommands(cmd) // Pass the *cli.Command

	// The Run method for a command takes a context and arguments.
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

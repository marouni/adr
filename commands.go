package main

import (
	"context" // Import context package
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

// NewCmd defines the 'new' command
var NewCmd = cli.Command{
	Name:    "new",
	Aliases: []string{"c"},
	Usage:   "Create a new ADR",
	Flags:   []cli.Flag{},
	Action: func(ctx context.Context, cmd *cli.Command) error { // Updated action signature
		currentConfig := getConfig()
		currentConfig.CurrentAdr++
		updateConfig(currentConfig)
		newAdr(currentConfig, cmd.Args().Slice()) // Use cmd.Args().Slice() for arguments
		return nil
	},
}

// InitCmd defines the 'init' command
var InitCmd = cli.Command{
	Name:        "init",
	Aliases:     []string{"i"},
	Usage:       "Initializes the ADR configurations",
	UsageText:   "adr init /home/user/adrs",
	Description: "Initializes the ADR configuration with an optional ADR base directory\n This is a a prerequisite to running any other adr sub-command",
	Action: func(ctx context.Context, cmd *cli.Command) error { // Updated action signature
		initDir := cmd.Args().Get(0) // Use cmd.Args().Get(0) for the first argument
		if initDir == "" {
			// Check if no arguments were provided, as Get(0) on empty Args might panic or return empty.
			// urfave/cli/v3 Args.Get(0) returns "" if not present, so this check is okay.
			initDir = adrDefaultBaseFolder
		}
		color.Green("Initializing ADR base at " + initDir)
		initBaseDir(initDir)
		initConfig(initDir)
		initTemplate()
		return nil
	},
}

func setCommands(rootCmd *cli.Command) { // Changed app *cli.App to rootCmd *cli.Command
	rootCmd.Commands = []*cli.Command{ // Correct field for subcommands in v3 is Commands
		&NewCmd, // Commands are now pointers
		&InitCmd,
	}
}

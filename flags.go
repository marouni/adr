package main

import (
	"github.com/urfave/cli/v3"
)

func setFlags(cmd *cli.Command) { // Changed app *cli.App to cmd *cli.Command
	cmd.Flags = []cli.Flag{}
}

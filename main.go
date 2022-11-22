package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "adr"
	app.Usage = "Work with Architecture Decision Records (ADRs)"
	app.Version = "0.1.1"

	setFlags(app)
	setCommands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

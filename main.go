package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {

	cli.VersionFlag = cli.BoolFlag{
		Name:  "print-version, V",
		Usage: "print only the version",
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Author = "toby"
	app.Email = "toby@deter.gent"

	app.Flags = globalFlags
	app.Commands = commands
	app.CommandNotFound = commandNotFound

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

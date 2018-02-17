package main

import (
	"fmt"
	"os"

	"git.playgrub.com/toby/det/command"
	"github.com/urfave/cli"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:    "listen",
		Usage:   "Build torrent database from network",
		Aliases: []string{"l"},
		Action:  command.CmdListen,
		Flags:   []cli.Flag{},
	},
	{
		Name:    "search",
		Usage:   "Search resolved torrents",
		Aliases: []string{"s"},
		Action:  command.CmdSearch,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "limit",
				Value: 25,
				Usage: "Number of popular torrents to show",
			},
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Verbose output",
			},
		},
	},
	{
		Name:    "resolve",
		Usage:   "Resolve a magnet hash",
		Aliases: []string{"r"},
		Action:  command.CmdResolve,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Verbose output",
			},
			cli.BoolFlag{
				Name:  "ping, p",
				Usage: "Ping nodes",
			},
		},
	},
	{
		Name:    "popular",
		Usage:   "List popular torrents",
		Aliases: []string{"p"},
		Action:  command.CmdPopular,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "limit",
				Value: 25,
				Usage: "Number of popular torrents to show",
			},
		},
	},
	{
		Name:    "info",
		Usage:   "Show Detergent info",
		Aliases: []string{"i"},
		Action:  command.CmdInfo,
		Flags:   []cli.Flag{},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.\n", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

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
		},
	},
	{
		Name:    "resolve",
		Usage:   "Resolve a magnet hash",
		Aliases: []string{"r"},
		Action:  command.CmdResolve,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "ping, p",
				Usage: "Ping nodes",
			},
		},
	},
	{
		Name:    "popular",
		Usage:   "List top torrents of all time",
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
		Name:    "timeline",
		Usage:   "List most popular torrents found each day",
		Aliases: []string{"t"},
		Action:  command.CmdTimeline,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "days",
				Value: 10,
				Usage: "Number of days to show",
			},
			cli.IntFlag{
				Name:  "limit",
				Value: 10,
				Usage: "Number torrents per day to show",
			},
		},
	},
	{
		Name:    "download",
		Usage:   "Download magnet URL",
		Aliases: []string{"d"},
		Action:  command.CmdDownload,
		Flags:   []cli.Flag{},
	},
	{
		Name:   "seed",
		Usage:  "Seed file on torrent network",
		Action: command.CmdSeed,
		Flags:  []cli.Flag{},
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

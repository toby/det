package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

// CmdListen builds a database of resolved announce hashes.
func CmdListen(c *cli.Context) error {
	cfg := server.Config{
		Listen: true,
		Seed:   true,
	}
	s := server.NewServer(&cfg)
	s.Run()
	return nil
}

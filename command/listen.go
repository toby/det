package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdListen(c *cli.Context) error {
	cfg := server.ServerConfig{
		Listen: true,
		Seed:   true,
	}
	s := server.NewServer(&cfg)
	s.Listen()
	return nil
}

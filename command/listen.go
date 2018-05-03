package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdListen(c *cli.Context) error {
	cfg := server.ServerConfig{
		StoreAnnounces: true,
		Seed:           false,
	}
	s := server.NewServer(&cfg)
	s.Listen()
	return nil
}

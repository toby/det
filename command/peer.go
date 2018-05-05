package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdPeer(c *cli.Context) error {
	cfg := server.ServerConfig{
		Listen: false,
		Seed:   true,
	}
	s := server.NewServer(&cfg)
	s.Run()
	return nil
}

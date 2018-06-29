package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

// CmdPeer puts det into peer discovery mode.
func CmdPeer(c *cli.Context) error {
	cfg := server.Config{
		Listen: false,
		Seed:   true,
	}
	s := server.NewServer(&cfg)
	s.Run()
	return nil
}

package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdListen(c *cli.Context) error {
	s := server.NewServer(true)
	s.Listen()
	return nil
}

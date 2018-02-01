package command

import (
	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdListen(c *cli.Context) error {
	s := server.NewServer()
	s.WaitForSig()
	return nil
}

package command

import (
	"github.com/toby/det"
	"github.com/urfave/cli"
)

// CmdPeer puts det into peer discovery mode.
func CmdPeer(c *cli.Context) error {
	cfg := det.Config{
		Listen: false,
		Seed:   true,
	}
	s, err := det.NewServer(&cfg)
	if err != nil {
		return err
	}
	s.Run()
	return nil
}

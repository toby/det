package command

import (
	"github.com/toby/det"
	"github.com/urfave/cli"
)

// CmdListen builds a database of resolved announce hashes.
func CmdListen(c *cli.Context) error {
	cfg := det.Config{
		Listen: true,
		Seed:   true,
	}
	s, err := det.NewServer(&cfg)
	if err != nil {
		return err
	}
	s.Run()
	return nil
}

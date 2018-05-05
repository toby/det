package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdResolve(c *cli.Context) error {
	cfg := server.ServerConfig{
		Listen: false,
		Seed:   false,
	}
	s := server.NewServer(&cfg)
	var hash string
	if c.NArg() > 0 {
		hash = c.Args().Get(0)
		log.Printf("Resolving: \"%s\"\n", hash)
	} else {
		log.Println("Missing hash")
		return nil
	}
	s.AddHash(hash)
	s.Run()
	return nil
}

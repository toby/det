package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdResolve(c *cli.Context) error {
	s := server.NewServer(false)
	var hash string
	if c.NArg() > 0 {
		hash = c.Args().Get(0)
		log.Printf("Resolving: \"%s\"\n", hash)
	} else {
		log.Println("Missing hash")
		return nil
	}
	s.AddHash(hash)
	s.Listen()
	return nil
}

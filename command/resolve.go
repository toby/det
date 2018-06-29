package command

import (
	"log"
	"time"

	"git.playgrub.com/toby/det/server"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/urfave/cli"
)

// CmdResolve resolves the metadata for a magnet uri supplied as a command line
// argument.
func CmdResolve(c *cli.Context) error {
	if c.NArg() > 0 {
		a := c.Args().Get(0)
		m, err := metainfo.ParseMagnetURI(a)
		if err != nil {
			return err
		}
		cfg := server.Config{
			Listen: false,
			Seed:   false,
		}
		s := server.NewServer(&cfg)
		i := s.ResolveHash(m.InfoHash, time.Second*120)
		if i != nil {
			log.Printf("Resolved: %s", i.Name)
		} else {
			log.Println("Resolve timeout")
		}
	} else {
		log.Println("Missing hash")
	}
	return nil
}

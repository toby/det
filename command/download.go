package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/urfave/cli"
)

// CmdDownload downloads a torrent from a magnet uri parsed from the command
// line arguments.
func CmdDownload(c *cli.Context) error {
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
		stor := storage.NewFile("./")
		t := <-s.DownloadInfoHash(m.InfoHash, 0, &stor)
		if t != nil {
			log.Printf("Downloaded: %s", t.Name())
		}
	} else {
		log.Println("Missing magnet URL")
	}
	return nil
}

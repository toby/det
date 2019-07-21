package command

import (
	"log"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/toby/det"
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
		cfg := det.Config{
			Listen: false,
			Seed:   false,
		}
		s, err := det.NewServer(&cfg)
		if err != nil {
			return err
		}
		// TODO: set the torrent download location from config
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

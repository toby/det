package command

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/toby/det"
	"github.com/urfave/cli"
)

// CmdSeed seeds the file provided as a command line argument.
func CmdSeed(c *cli.Context) error {
	if c.NArg() > 0 {
		p := c.Args().Get(0)
		mi := metainfo.MetaInfo{
			AnnounceList: det.BuiltinAnnounceList,
		}
		mi.SetDefaults()
		info := metainfo.Info{
			PieceLength: 256 * 1024,
		}
		err := info.BuildFromFilePath(p)
		if err != nil {
			return err
		}
		mi.InfoBytes, err = bencode.Marshal(info)
		if err != nil {
			return err
		}
		cfg := det.Config{
			Listen: false,
			Seed:   true,
		}
		s, err := det.NewServer(&cfg)
		if err != nil {
			return err
		}
		t, err := s.AddMetaInfo(&mi)
		if err != nil {
			return err
		}
		fmt.Printf("Seeding: %s - magnet:?xt=urn:btih:%s\n", p, t.InfoHash().HexString())
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			for {
				printTorrentStats(t)
				fmt.Println()
				<-time.After(time.Second * 10)
			}
		}()
		<-sigs
	} else {
		log.Println("Missing file path")
	}
	return nil
}

package command

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.playgrub.com/toby/det/server"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/urfave/cli"
)

func CmdSeed(c *cli.Context) error {
	if c.NArg() > 0 {
		p := c.Args().Get(0)
		mi := metainfo.MetaInfo{
			AnnounceList: server.BuiltinAnnounceList,
		}
		mi.SetDefaults()
		info := metainfo.Info{
			PieceLength: 256 * 1024,
		}
		err := info.BuildFromFilePath(p)
		if err != nil {
			panic(err)
		}
		mi.InfoBytes, err = bencode.Marshal(info)
		if err != nil {
			panic(err)
		}
		cfg := server.ServerConfig{
			Listen: false,
			Seed:   true,
		}
		s := server.NewServer(&cfg)
		t, err := s.AddMetaInfo(&mi)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Seeding: %s - magnet:?xt=urn:btih:%s\n", p, t.InfoHash().HexString())
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			for {
				PrintTorrentStats(t)
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

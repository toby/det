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

var (
	builtinAnnounceList = [][]string{
		{"udp://tracker.openbittorrent.com:80"},
		{"udp://tracker.publicbt.com:80"},
		{"udp://tracker.istole.it:6969"},
	}
)

func CmdSeed(c *cli.Context) error {
	if c.NArg() > 0 {
		p := c.Args().Get(0)
		mi := metainfo.MetaInfo{
			AnnounceList: builtinAnnounceList,
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
		defer s.Client.Close()
		t, err := s.Client.AddTorrent(&mi)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Seeding: %s - magnet:?xt=urn:btih:%s\n", p, t.InfoHash().HexString())
		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			for {
				select {
				case <-time.After(time.Second * 10):
					PrintTorrentStats(t)
					fmt.Println()
				case <-sigs:
					done <- true
				}
			}
		}()
		<-done
	} else {
		log.Println("Missing file path")
	}
	return nil
}

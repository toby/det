package command

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdDownload(c *cli.Context) error {
	cfg := server.ServerConfig{
		StoreAnnounces: false,
		Seed:           false,
	}
	s := server.NewServer(&cfg)
	defer s.Client.Close()
	if c.NArg() > 0 {
		mag := c.Args().Get(0)
		t, err := s.Client.AddMagnet(mag)
		if err != nil {
			panic(err)
		}
		log.Printf("Resolving: %s", mag)
		<-t.GotInfo()
		log.Printf("Downloading: %s", t)
		t.DownloadAll()
		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		downloaded := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			for {
				br := t.Stats().ConnStats.BytesReadUsefulData
				p := 100 * float64(br) / float64(t.Length())
				if p == 100 {
					downloaded <- true
				}
				select {
				case <-downloaded:
					for _, f := range t.Files() {
						log.Printf("Downloaded File: %s", f.Path())
					}
					t.Drop()
					done <- true
				case <-time.After(time.Second * 10):
					log.Printf("Downloaded: %.2f%% - %d of %d bytes", p, br, t.Length())
				case <-sigs:
					done <- true
				}
			}
		}()
		<-done

	} else {
		log.Println("Missing magnet URL")
	}
	return nil
}

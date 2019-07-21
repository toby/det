package command

import (
	"log"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:     "download",
	Short:   "Download magnet url from torrent network",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),
	RunE:    downloadCmdRun,
}

func downloadCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	cfg.Seed = false
	cfg.Listen = false
	s, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	m, err := metainfo.ParseMagnetURI(args[0])
	if err != nil {
		return err
	}
	stor := storage.NewFile(cfg.DownloadPath)
	t := <-s.DownloadInfoHash(m.InfoHash, 0, &stor)
	if t != nil {
		log.Printf("Downloaded: %s", t.Name())
	}
	return nil
}

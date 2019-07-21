package command

import (
	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

func init() {
	rootCmd.AddCommand(listenCmd)
}

var listenCmd = &cobra.Command{
	Use:     "listen",
	Short:   "Build torrent database from network",
	Aliases: []string{"l"},
	RunE:    listenCmdRun,
}

func listenCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	cfg.Listen = true
	cfg.Seed = true
	s, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	s.Run()
	return nil
}

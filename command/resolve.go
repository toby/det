package command

import (
	"log"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

var resolveTimeout time.Duration

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().DurationVarP(&resolveTimeout, "timeout", "t", time.Second*120, "Resolution timeout")
}

var resolveCmd = &cobra.Command{
	Use:     "resolve",
	Short:   "Resolve a magnet url and add it to the database",
	Aliases: []string{"r"},
	Args:    cobra.ExactArgs(1),
	RunE:    resolveCmdRun,
}

func resolveCmdRun(cmd *cobra.Command, args []string) error {
	m, err := metainfo.ParseMagnetURI(args[0])
	if err != nil {
		return err
	}
	cfg := serverConfigFromDefaults()
	s, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	i := s.ResolveHash(m.InfoHash, resolveTimeout)
	if i != nil {
		log.Printf("Resolved: %s", i.Name)
	} else {
		log.Println("Resolve timeout")
	}
	return nil
}

package command

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

var popularLimit int

func init() {
	rootCmd.AddCommand(popularCmd)
	popularCmd.Flags().IntVarP(&popularLimit, "limit", "l", 50, "Limit results")
}

var popularCmd = &cobra.Command{
	Use:     "popular",
	Short:   "List most announced torrents",
	Aliases: []string{"p"},
	Args:    cobra.ArbitraryArgs,
	RunE:    popularCmdRun,
}

func popularCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	db, err := server.NewSqliteDB(cfg.SqlitePath)
	if err != nil {
		return err
	}
	defer db.Close()
	ts, err := db.PopularTorrents(popularLimit)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	for _, t := range ts {
		printRankedTorrent(t)
	}
	return nil
}

package command

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

var searchLimit int

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 50, "Limit results")
}

var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search resolved torrents",
	Aliases: []string{"s"},
	Args:    cobra.ArbitraryArgs,
	RunE:    searchCmdRun,
}

func searchCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	db, err := server.NewSqliteDB(cfg.SqlitePath)
	if err != nil {
		return err
	}
	defer db.Close()
	term := strings.Join(args, " ")
	log.Printf("Searching: \"%s\"\n", term)
	rows, err := db.SearchTorrents(term, searchLimit)
	if err != nil {
		return err
	}
	for _, t := range rows {
		printRankedTorrent(t)
	}
	return nil
}

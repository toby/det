package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toby/det/server"
)

var timelineDays int
var timelineLimit int

func init() {
	rootCmd.AddCommand(timelineCmd)
	timelineCmd.Flags().IntVarP(&timelineDays, "days", "d", 10, "Limit number of days")
	timelineCmd.Flags().IntVarP(&timelineLimit, "limit", "l", 10, "Limit results per day")
}

var timelineCmd = &cobra.Command{
	Use:     "timeline",
	Short:   "Timeline resolved torrents",
	Aliases: []string{"s"},
	Args:    cobra.ArbitraryArgs,
	RunE:    timelineCmdRun,
}

func timelineCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	db, err := server.NewSqliteDB(cfg.SqlitePath)
	if err != nil {
		return err
	}
	defer db.Close()
	tl, err := db.TimelineTorrents(timelineDays, timelineLimit)
	if err != nil {
		return err
	}
	for _, entry := range tl {
		if len(entry.Torrents) > 0 {
			fmt.Printf("%s\n", underline(entry.Day.Format("Mon Jan _2")))
			for _, t := range entry.Torrents {
				printRankedTorrent(t)
			}
			println()
		}
	}
	return nil
}

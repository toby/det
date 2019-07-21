package command

import (
	"fmt"
	"log"

	"github.com/toby/det"
	"github.com/urfave/cli"
)

// CmdTimeline shows the top torrents discovered by day.
func CmdTimeline(c *cli.Context) error {
	db, err := det.NewSqliteDB("./")
	if err != nil {
		return err
	}
	defer db.Close()
	days := c.Int("days")
	limit := c.Int("limit")
	tl, err := db.TimelineTorrents(days, limit)
	if err != nil {
		log.Printf("ERROR: %s", err)
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

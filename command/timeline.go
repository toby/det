package command

import (
	"fmt"
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdTimeline(c *cli.Context) error {
	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./")
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
			fmt.Printf("%s\n", Underline(entry.Day.Format("Mon Jan _2")))
			for _, t := range entry.Torrents {
				PrintRankedTorrent(t)
			}
			println()
		}
	}
	return nil
}

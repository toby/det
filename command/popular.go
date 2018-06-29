package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

// CmdPopular displays the most popular torrents in the det db based on number
// of announces.
func CmdPopular(c *cli.Context) error {
	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./")
	defer db.Close()
	limit := c.Int("limit")
	ts, err := db.PopularTorrents(limit)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	for _, t := range ts {
		printRankedTorrent(t)
	}
	return nil
}

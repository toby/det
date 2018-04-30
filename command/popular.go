package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

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
		PrintRankedTorrent(t)
	}
	return nil
}

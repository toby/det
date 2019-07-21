package command

import (
	"log"

	"github.com/toby/det"
	"github.com/urfave/cli"
)

// CmdPopular displays the most popular torrents in the det db based on number
// of announces.
func CmdPopular(c *cli.Context) error {
	db, err := det.NewSqliteDB("./")
	if err != nil {
		return err
	}
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

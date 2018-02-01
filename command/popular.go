package command

import (
	"fmt"
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdPopular(c *cli.Context) error {
	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./").(*server.SqliteDBClient)
	defer db.Close()
	limit := c.Int("limit")
	rows, err := db.PopularTorrents(limit)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	log.Printf("rows: %d", len(rows))
	for _, row := range rows {
		peers, _ := row.AnnounceCount.Value()
		name, _ := row.Name.Value()
		if name == nil {
			name = "UNRESOLVED"
		}
		infoHash, _ := row.InfoHash.Value()
		fmt.Printf("%d\t%s\tmagnet:?xt=urn:btih:%s\n", peers, name, infoHash)
	}
	return nil
}

package command

import (
	"fmt"
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func CmdSearch(c *cli.Context) error {
	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./")
	defer db.Close()
	var term string

	if c.NArg() > 0 {
		term = c.Args().Get(0)
		fmt.Printf("Searching: \"%s\"\n", term)
	} else {
		fmt.Println("Missing search term")
		return nil
	}
	limit := c.Int("limit")
	rows, err := db.SearchTorrents(term, limit)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	for _, t := range rows {
		PrintRankedTorrent(t)
	}
	return nil
}

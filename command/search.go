package command

import (
	"fmt"
	"io/ioutil"
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
)

func PrintRankedTorrent(t server.RankedTorrent) {
	peers, _ := t.AnnounceCount.Value()
	infoHash, _ := t.InfoHash.Value()
	name, _ := t.Name.Value()
	if name == nil {
		name = "UNRESOLVED"
	}
	fmt.Printf("%-9d %-80s magnet:?xt=urn:btih:%-40s\n", peers, name, infoHash)
}

func CmdSearch(c *cli.Context) error {
	if !c.Bool("verbose") {
		log.SetOutput(ioutil.Discard)
	}

	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./").(*server.SqliteDBClient)
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

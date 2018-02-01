package command

import (
	"log"

	"git.playgrub.com/toby/det/server"
	"github.com/urfave/cli"
	"golang.org/x/text/message"
)

func CmdInfo(c *cli.Context) error {
	var db *server.SqliteDBClient
	db = server.NewSqliteDB("./").(*server.SqliteDBClient)
	defer db.Close()
	stats, err := db.Stats()
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	p := message.NewPrinter(message.MatchLanguage("en"))
	p.Printf("Torrents:\t%v\n", stats.Torrents)
	p.Printf("Resolved:\t%v\n", stats.Resolved)
	p.Printf("Announces:\t%v\n", stats.Announces)
	return nil
}

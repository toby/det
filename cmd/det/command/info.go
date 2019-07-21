package command

import (
	"github.com/toby/det"
	"github.com/urfave/cli"
	"golang.org/x/text/message"
)

// CmdInfo prints statistics about the det database.
func CmdInfo(c *cli.Context) error {
	db, err := det.NewSqliteDB("./")
	if err != nil {
		return err
	}
	defer db.Close()
	stats, err := db.Stats()
	if err != nil {
		return err
	}
	p := message.NewPrinter(message.MatchLanguage("en"))
	p.Printf("Torrents:\t%v\n", stats.Torrents)
	p.Printf("Resolved:\t%v\n", stats.Resolved)
	p.Printf("Announces:\t%v\n", stats.Announces)
	return nil
}

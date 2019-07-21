package command

import (
	"github.com/spf13/cobra"
	"github.com/toby/det/server"
	"golang.org/x/text/message"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show Detergent info",
	Aliases: []string{"i"},
	RunE:    infoCmdRun,
}

func infoCmdRun(cmd *cobra.Command, args []string) error {
	cfg := serverConfigFromDefaults()
	db, err := server.NewSqliteDB(cfg.SqlitePath)
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

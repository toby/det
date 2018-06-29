package command

import (
	"fmt"
	"regexp"

	"git.playgrub.com/toby/det/server"
	"github.com/anacrolix/torrent"
)

func printRankedTorrent(t server.Torrent) {
	name := t.Name
	if name == "" {
		name = "-- unresolved --"
	}
	fmt.Printf("%-9d %-80s magnet:?xt=urn:btih:%-40s\n", t.AnnounceCount, name, t.InfoHash)
}

func printTorrentStats(t *torrent.Torrent) {
	fmt.Printf("Seeding:           %t\n", t.Seeding())
	fmt.Printf("Total Peers:       %d\n", t.Stats().TotalPeers)
	fmt.Printf("Pending Peers:     %d\n", t.Stats().PendingPeers)
	fmt.Printf("Active Peers:      %d\n", t.Stats().ActivePeers)
	fmt.Printf("Connected Seeders: %d\n", t.Stats().ConnectedSeeders)
	fmt.Printf("Half Open Peers:   %d\n", t.Stats().HalfOpenPeers)
}

func underline(s string) string {
	r := regexp.MustCompile(".")
	u := r.ReplaceAllString(s, "-")
	return fmt.Sprintf("%s\n%s", s, u)
}

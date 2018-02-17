package command

import (
	"fmt"

	"git.playgrub.com/toby/det/server"
)

func PrintRankedTorrent(t server.Torrent) {
	name := t.Name
	if name == "" {
		name = "-- unresolved --"
	}
	fmt.Printf("%-9d %-80s magnet:?xt=urn:btih:%-40s\n", t.AnnounceCount, name, t.InfoHash)
}

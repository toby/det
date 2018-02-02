package command

import (
	"fmt"

	"git.playgrub.com/toby/det/server"
)

func PrintRankedTorrent(t server.RankedTorrent) {
	peers, _ := t.AnnounceCount.Value()
	infoHash, _ := t.InfoHash.Value()
	name, _ := t.Name.Value()
	if name == nil {
		name = "-- unresolved --"
	}
	fmt.Printf("%-9d %-80s magnet:?xt=urn:btih:%-40s\n", peers, name, infoHash)
}

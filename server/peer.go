package server

import (
	"encoding/json"

	"github.com/anacrolix/torrent"
)

const DetVersion = "0.1"

type DetAnnounce struct {
	Name    string
	Version string
}

func DetAnnounceTorrentSpec() *torrent.TorrentSpec {
	d := DetAnnounce{
		Name:    "detergent",
		Version: DetVersion,
	}
	b, _ := json.Marshal(d)
	return TorrentSpecForBytes("detergent.json", b)
}

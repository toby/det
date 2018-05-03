package server

import (
	"encoding/json"

	"github.com/anacrolix/torrent/metainfo"
)

const DetVersion = "0.1"

type DetAnnounce struct {
	Name    string
	Version string
}

func DetAnnounceInfo() *metainfo.Info {
	d := DetAnnounce{
		Name:    "detergent",
		Version: DetVersion,
	}
	b, _ := json.Marshal(d)
	info, _ := InfoForBytes("detergent.json", b)
	return info
}

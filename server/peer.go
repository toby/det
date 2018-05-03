package server

import (
	"encoding/json"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

const DetVersion = "0.1"

var (
	BuiltinAnnounceList = [][]string{
		{"udp://tracker.openbittorrent.com:80"},
		{"udp://tracker.publicbt.com:80"},
		{"udp://tracker.istole.it:6969"},
	}
)

type DetAnnounce struct {
	Name    string
	Version string
}

func DetAnnounceMetaInfo() *metainfo.MetaInfo {
	d := DetAnnounce{
		Name:    "detergent",
		Version: DetVersion,
	}
	b, _ := json.Marshal(d)
	info, err := InfoForBytes("detergent.json", b)
	mi := &metainfo.MetaInfo{
		AnnounceList: BuiltinAnnounceList,
	}
	mi.SetDefaults()
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		panic(err)
	}
	return mi
}

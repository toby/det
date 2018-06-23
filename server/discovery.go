package server

import (
	"encoding/json"
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// Used to satisfy json.Marshaler when seeding
type DiscoverMessage interface{}

const version = "0.1"

type Version struct {
	Name    string
	Version string
}

type Peer struct {
	Version Version
	Hash    metainfo.Hash
}

func torrentSpecForMessage(name string, m DiscoverMessage) *torrent.TorrentSpec {
	var b TorrentBytes
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b.TorrentSpec(name)
}

func CurrentVersion() Version {
	return Version{
		Version: version,
		Name:    "detergent",
	}
}

func CreatePeer(h metainfo.Hash) Peer {
	return Peer{
		Version: CurrentVersion(),
		Hash:    h,
	}
}

func VersionTorrentSpec() *torrent.TorrentSpec {
	return torrentSpecForMessage("detergent.json", CurrentVersion())
}

func PeerTorrentSpec(h metainfo.Hash) *torrent.TorrentSpec {
	return torrentSpecForMessage(fmt.Sprintf("%s.json", h.HexString()), CreatePeer(h))
}

package server

import (
	"encoding/json"
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// Peer discovery protocol version
const ver = "0.1"

type versionMessage struct {
	Name    string
	Version string
}

type peerMessage struct {
	Version versionMessage
	Hash    metainfo.Hash
}

func torrentSpecForMessage(name string, m interface{}) *torrent.TorrentSpec {
	var b TorrentBytes
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b.TorrentSpec(name)
}

func currentVersion() versionMessage {
	return versionMessage{
		Version: ver,
		Name:    "detergent",
	}
}

func createPeerMessage(h metainfo.Hash) peerMessage {
	return peerMessage{
		Version: currentVersion(),
		Hash:    h,
	}
}

func VersionTorrentSpec() *torrent.TorrentSpec {
	return torrentSpecForMessage("detergent.json", currentVersion())
}

func PeerTorrentSpec(h metainfo.Hash) *torrent.TorrentSpec {
	return torrentSpecForMessage(fmt.Sprintf("%s.json", h.HexString()), createPeerMessage(h))
}

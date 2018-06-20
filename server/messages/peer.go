package messages

import (
	"github.com/anacrolix/torrent/metainfo"
)

const version = "0.1"

// Used to satisfy json.Marshaler when seeding
type Message interface{}

type Version struct {
	Name    string
	Version string
}

type Peer struct {
	Version Version
	Hash    metainfo.Hash
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

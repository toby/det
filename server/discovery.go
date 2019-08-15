package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

const discoverVersion = "0.2"

// Discoverable is an interface peers must implement to work with the discover
// protocol.
type Discoverable interface {
	// Namespace is a unique string for this application. When seeded, the
	// discoverVersion will be appended.
	Namespace() string

	// PeerId returns a metainfo.Hash unique to this peer.
	PeerID() metainfo.Hash

	// AddPeer will be called with a torrent.Peer as they are discovered.
	AddPeer(torrent.Peer)
}

// TorrentPeer is an interface to functionality in the `anacrolix/torrent` library.
type TorrentPeer interface {
	// TorrentClient returns the client for lower level torrent operations.
	TorrentClient() *torrent.Client

	// DownloadInfoHash will return a channel that closes after the
	// specified timeout or returns a *torrent.Torrent if it was able to
	// fully download. A timeout of zero will block until the torrent is
	// downloaded.
	DownloadInfoHash(metainfo.Hash, time.Duration, *storage.ClientImpl) <-chan *torrent.Torrent
}

// DiscoveryPeer is a composite type to combine Discoverable and TorrentPeer.
type DiscoveryPeer interface {
	Discoverable
	TorrentPeer
}

type discoverMessage interface {
	name() string
	data() []byte
}

type messageData struct{}

// data marshals the receiver into json and returns the []byte representation.
func (m messageData) data() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

// namepsaceMessage is the semaphore message used to signify participation in
// the discovery protocol.
type namespaceMessage struct {
	messageData
	namespace string
}

func (m namespaceMessage) name() string {
	return fmt.Sprintf("%s.json", m.namespace)
}

// peerMessage represents unique peers in the namespace. Seeding this message
// allows the validation of peers when this message is deterministically
// constructed by a remote peer using the peer id discovered in the swarm of
// the namespaceMessage.
type peerMessage struct {
	messageData
	namespace string
	hash      metainfo.Hash
}

func (m peerMessage) name() string {
	return fmt.Sprintf("%s.json", m.hash.HexString())
}

// StartDiscovery begins and coordinates the discovery protocol. It returns a
// channel of verified peers. AddPeer will also be called on d as they are
// discovered.
func StartDiscovery(d DiscoveryPeer) <-chan torrent.Peer {
	n := namespace(d.Namespace())
	ps := make(chan torrent.Peer)
	nm := namespaceMessage{
		namespace: n,
	}
	t := seedMessage(d.TorrentClient(), nm)
	pm := peerMessage{
		namespace: n,
		hash:      d.PeerID(),
	}
	seedMessage(d.TorrentClient(), pm)
	go func() {
		peers := extractPeers(t)
		for p := range verifyPeers(d, peers) {
			d.AddPeer(p)
			ps <- p
		}
	}()
	return ps
}

func verifyPeer(d DiscoveryPeer, p torrent.Peer, out chan<- torrent.Peer) {
	dht := d.TorrentClient().DhtServers()[0]
	n := namespace(d.Namespace())
	ip := fmt.Sprintf("%s:%d", p.IP, p.Port)
	a, err := net.ResolveUDPAddr("udp", ip)
	if err != nil {
		return
	}
	log.Printf("Ping: %s", ip)
	_ = dht.Ping(a, func(m krpc.Msg, err error) {
		if err == nil {
			h := metainfo.HashBytes(m.R.ID[:])
			log.Printf("Pong: %s\t%s", h.HexString(), ip)
			ts := torrentSpecForMessage(peerMessage{namespace: n, hash: h})
			t := <-d.DownloadInfoHash(ts.InfoHash, time.Second*120, nil)
			if t != nil {
				log.Printf("Peer Verified: %s\t%s", h.HexString(), ip)
				out <- p
				t.Drop()
			} else {
				log.Printf("Peer Not Verified: %s\t%s", h.HexString(), ip)
			}
		}
	})
}

func verifyPeers(d DiscoveryPeer, in <-chan torrent.Peer) <-chan torrent.Peer {
	out := make(chan torrent.Peer)
	go func() {
		for p := range in {
			verifyPeer(d, p, out)
		}
		close(out)
	}()
	return out
}

func extractPeers(t *torrent.Torrent) <-chan torrent.Peer {
	out := make(chan torrent.Peer)
	go func() {
		seen := make(map[string]torrent.Peer)
		for {
			for _, p := range t.KnownSwarm() {
				h := metainfo.HashBytes(p.Id[:]).HexString()
				_, ok := seen[h]
				if !ok {
					seen[h] = p
					out <- p
				}
			}
			<-time.After(time.Second * 5)
		}
	}()
	return out
}

func torrentSpecForMessage(m discoverMessage) *torrent.TorrentSpec {
	var b TorrentBytes
	b = m.data()
	return b.TorrentSpec(m.name())
}

func seedMessage(cl *torrent.Client, m discoverMessage) *torrent.Torrent {
	ts := torrentSpecForMessage(m)
	t, err := seedTorrentSpec(cl, ts)
	if err != nil {
		panic(err)
	}
	log.Printf("Seeding %s: magnet:?xt=urn:btih:%s\n", m.name(), t.InfoHash().HexString())
	return t
}

func namespace(n string) string {
	return fmt.Sprintf("%s-%s", n, discoverVersion)
}

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

const discoverVersion = "0.2"

type Discoverable interface {
	Namespace() string
	PeerId() metainfo.Hash
	AddPeer(torrent.Peer)
}

type TorrentPeer interface {
	TorrentClient() *torrent.Client
	DownloadInfoHash(metainfo.Hash, time.Duration, *storage.ClientImpl) <-chan *torrent.Torrent
}

type DiscoveryPeer interface {
	Discoverable
	TorrentPeer
}

type discoverMessage interface {
	name() string
	data() []byte
}

type messageData struct{}

func (m messageData) data() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

type namespaceMessage struct {
	messageData
	namespace string
}

func (m namespaceMessage) name() string {
	return fmt.Sprintf("%s.json", m.namespace)
}

type peerMessage struct {
	messageData
	namespace string
	hash      metainfo.Hash
}

func (m peerMessage) name() string {
	return fmt.Sprintf("%s.json", m.hash.HexString())
}

func StartDiscovery(d DiscoveryPeer, c *torrent.Client) <-chan torrent.Peer {
	n := namespace(d.Namespace())
	ps := make(chan torrent.Peer)
	nm := namespaceMessage{
		namespace: n,
	}
	t := seedNamespace(d.TorrentClient(), nm)
	pm := peerMessage{
		namespace: n,
		hash:      d.PeerId(),
	}
	seedPeer(d.TorrentClient(), pm)
	go func() {
		peers := extractPeers(t)
		for p := range verifyPeers(d, peers) {
			d.AddPeer(p)
			ps <- p
		}
	}()
	return ps
}

func verifyPeers(d DiscoveryPeer, in <-chan torrent.Peer) <-chan torrent.Peer {
	out := make(chan torrent.Peer)
	go func() {
		dht := d.TorrentClient().DHT()
		n := namespace(d.Namespace())
		for p := range in {
			ip := fmt.Sprintf("%s:%d", p.IP, p.Port)
			a, err := net.ResolveUDPAddr("udp", ip)
			if err == nil {
				log.Printf("Ping: %s", ip)
				dht.Ping(a, func(m krpc.Msg, err error) {
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

func seedTorrentSpec(cl *torrent.Client, ts *torrent.TorrentSpec) (*torrent.Torrent, error) {
	t, _, err := cl.AddTorrentSpec(ts)
	return t, err
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

func seedNamespace(cl *torrent.Client, m namespaceMessage) *torrent.Torrent {
	return seedMessage(cl, m)
}

func seedPeer(cl *torrent.Client, m peerMessage) *torrent.Torrent {
	return seedMessage(cl, m)
}

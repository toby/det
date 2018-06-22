package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"git.playgrub.com/toby/det/server/messages"
	"github.com/anacrolix/dht"
	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/mattn/go-sqlite3"
	"github.com/muesli/cache2go"
)

const ResolveTimeout = time.Second * 30
const ResolveWindow = time.Minute * 10

type Server struct {
	Client         *torrent.Client
	hashes         []string
	db             *SqliteDBClient
	hashLock       sync.Mutex
	resolveCache   *cache2go.CacheTable
	listen         bool
	seed           bool
	versionTorrent *torrent.Torrent
	peerTorrent    *torrent.Torrent
}

type ServerConfig struct {
	Listen bool
	Seed   bool
}

type TorrentResolver interface {
	AddHash(h string)
}

func torrentSpecForMessage(name string, m messages.Message) *torrent.TorrentSpec {
	var b TorrentBytes
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b.TorrentSpec(name)
}

func (s *Server) SeedMessage(name string, m messages.Message) (*torrent.Torrent, error) {
	ts := torrentSpecForMessage(name, m)
	t, _, err := s.Client.AddTorrentSpec(ts)
	return t, err
}

func (s *Server) SeedVersion() {
	if s.versionTorrent == nil {
		t, err := s.SeedMessage("detergent.json", messages.CurrentVersion())
		if err != nil {
			panic(err)
		}
		s.versionTorrent = t
		log.Printf("Seeding detergent.json: magnet:?xt=urn:btih:%s\n", s.versionTorrent.InfoHash().HexString())
	}
}

func (s *Server) SeedPeer() {
	if s.peerTorrent == nil {
		id := s.Client.DHT().ID()
		h := metainfo.HashBytes(id[:])
		n := fmt.Sprintf("%s.json", h.HexString())
		t, err := s.SeedMessage(n, messages.CreatePeer(h))
		if err != nil {
			panic(err)
		}
		s.peerTorrent = t
		log.Printf("Seeding %s: magnet:?xt=urn:btih:%s\n", n, t.InfoHash().HexString())
	}
}

func (s *Server) AddHash(h string) error {
	if len(h) != 40 {
		return errors.New("Invalid hash length")
	}

	err := s.db.CreateTorrent(h)
	if err != nil {
		// duplicate hash
		if err.(sqlite3.Error).Code != 19 {
			log.Printf("AddHash Error:\t%s", err)
			return err
		}

		// have we tried to resolve this in the last 10 mins?
		if s.resolveCache.Exists(h) {
			return nil
		}
	}

	s.resolveCache.Add(h, ResolveWindow, true)
	s.hashes = append(s.hashes, h)
	return nil
}

func (s *Server) onQuery(query *krpc.Msg, source net.Addr) bool {
	if query.Q == "get_peers" && bytes.Equal(query.A.InfoHash[:], s.versionTorrent.InfoHash().Bytes()) {
		//log.Printf("Detergent API GetPeers: %s", query.IP)
	}
	return true
}

func (s *Server) onAnnouncePeer(h metainfo.Hash, peer dht.Peer) {
	s.hashLock.Lock()
	defer s.hashLock.Unlock()
	hx := h.HexString()

	if err := s.AddHash(hx); err != nil {
		log.Printf("Error adding hash: %s", err)
		return
	}

	if err := s.db.CreateAnnounce(hx, peer.String()); err != nil {
		log.Printf("CreateAnnounce Error:\t%s", err)
	}
}

func (s *Server) resolveHash(hx string) error {
	st, err := s.db.GetTorrent(hx)
	if err == sql.ErrNoRows || st.ResolvedAt.IsZero() {
		h := metainfo.NewHashFromHex(hx)
		t, new := s.Client.AddTorrentInfoHashWithStorage(h, make(TorrentBytes, 0))
		if new {
			select {
			case <-t.GotInfo():
				s.db.CreateTorrentSearch(t.InfoHash().HexString(), t.Name())
				s.db.SetTorrentMeta(t.InfoHash().HexString(), t.Name())
				log.Printf("Resolved:\t%s\t%s", t.InfoHash().HexString(), t.Name())
			case <-time.After(ResolveTimeout):
				log.Printf("Timeout:\t%s", h)
			}
		} else {
			log.Printf("Resolved Found:\t%s", t)
		}
		t.Drop()
	} else if err != nil {
		return fmt.Errorf("GetTorrent err:\t%s", err)
	} else {
		log.Printf("Found:\t%s\t%s", st.InfoHash, st.Name)
	}

	return nil
}

func (s *Server) resolvePeer(h metainfo.Hash) {
	go func() {
		n := fmt.Sprintf("%s.json", h.HexString())
		ts := torrentSpecForMessage(n, messages.CreatePeer(h))
		if t, _ := s.Client.Torrent(ts.InfoHash); t == nil {
			log.Printf("Resolving Peer: %s", h.HexString())
			t, _ := s.Client.AddTorrentInfoHash(ts.InfoHash)
			select {
			case <-t.GotInfo():
				t.DownloadAll()
				p := float64(0)
				for p != 100 {
					br := t.Stats().ConnStats.BytesReadUsefulData
					p = 100 * float64(br) / float64(t.Length())
					<-time.After(time.Second * 1)
					log.Printf("P: %s, br: %s, t.len: %s", p, br, t.Length())
				}
				log.Printf("Peer Resolved:\t%s", h.HexString())
			case <-time.After(time.Second * 10):
				log.Printf("Peer Timeout:\t%s", h)
			}
			t.Drop()
		} else {
			log.Printf("Already Resolving: %s", h.HexString())
		}
	}()
}

func (s *Server) verifyPeer(p torrent.Peer) {
	dht := s.Client.DHT()
	ip := fmt.Sprintf("%s:%d", p.IP, p.Port)
	a, err := net.ResolveUDPAddr("udp", ip)
	if err == nil {
		log.Printf("Ping: %s", ip)
		dht.Ping(a, func(m krpc.Msg, err error) {
			if err == nil {
				h := metainfo.HashBytes(m.R.ID[:])
				log.Printf("Pong: %s\t%s", h.HexString(), ip)
				s.resolvePeer(h)
			}
		})
	}
}

func (s *Server) extractPeers(t *torrent.Torrent) chan torrent.Peer {
	peers := make(chan torrent.Peer)
	go func() {
		l := &sync.Mutex{}
		seen := make(map[string]torrent.Peer)
		for {
			for _, p := range t.KnownSwarm() {
				h := metainfo.HashBytes(p.Id[:]).HexString()
				l.Lock()
				_, ok := seen[h]
				l.Unlock()
				if !ok {
					l.Lock()
					seen[h] = p
					l.Unlock()
					peers <- p
				}
			}
			<-time.After(time.Second * 5)
		}
	}()
	return peers
}

func (s *Server) Run() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			if len(s.hashes) > 0 {
				err := s.resolveHash(s.hashes[0])
				if err != nil {
					log.Println(err)
				}
				s.hashes = s.hashes[1:]
			} else {
				<-time.After(time.Second)
			}
		}
	}()
	<-sigs
	s.db.Close()
	log.Printf("Exiting Detergent, here are some stats:")
	s.Client.WriteStatus(os.Stderr)
}

func NewServer(cfg *ServerConfig) *Server {
	if cfg == nil {
		cfg = &ServerConfig{
			Listen: true,
			Seed:   false,
		}
	}
	db := NewSqliteDB("./")
	s := &Server{
		Client:         nil,
		hashes:         make([]string, 0),
		resolveCache:   cache2go.Cache("resolveCache"),
		listen:         cfg.Listen,
		seed:           cfg.Seed,
		db:             db,
		versionTorrent: nil,
		peerTorrent:    nil,
	}

	dcfg := dht.ServerConfig{StartingNodes: dht.GlobalBootstrapAddrs}
	if s.listen {
		dcfg.OnAnnouncePeer = s.onAnnouncePeer
	}
	if s.seed {
		dcfg.OnQuery = s.onQuery
	}
	torrentCfg := torrent.Config{
		DHTConfig:      dcfg,
		Seed:           cfg.Seed,
		DefaultStorage: storage.NewBoltDB("./"),
	}
	cl, err := torrent.NewClient(&torrentCfg)
	id := cl.PeerID()
	log.Printf("Torrent Peer ID: %s", metainfo.HashBytes(id[:]).HexString())
	id = cl.DHT().ID()
	log.Printf("DHT Node ID: %s", metainfo.HashBytes(id[:]).HexString())
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	s.Client = cl
	if s.seed {
		s.SeedVersion()
		s.SeedPeer()
		go func() {
			peers := s.extractPeers(s.versionTorrent)
			for p := range peers {
				s.verifyPeer(p)
			}
		}()
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			cl.WriteStatus(w)
		})
		http.HandleFunc("/torrent", func(w http.ResponseWriter, r *http.Request) {
			cl.WriteStatus(w)
		})
		http.HandleFunc("/dht", func(w http.ResponseWriter, r *http.Request) {
			cl.DHT().WriteStatus(w)
		})
		if s.listen {
			log.Println("Web stats listening on: http://0.0.0.0:8888")
		}
		log.Fatal(http.ListenAndServe(":8888", nil))
	}()
	return s
}

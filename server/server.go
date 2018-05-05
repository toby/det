package server

import (
	"database/sql"
	"encoding/hex"
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

	"github.com/anacrolix/dht"
	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/mattn/go-sqlite3"
	"github.com/muesli/cache2go"
)

const ResolveWindow = time.Minute * 10

type Server struct {
	Client       *torrent.Client
	hashes       []string
	db           *SqliteDBClient
	hashLock     sync.Mutex
	resolveCache *cache2go.CacheTable
	listen       bool
	seed         bool
	apiTorrent   *torrent.Torrent
}

type ServerConfig struct {
	Listen bool
	Seed   bool
}

type TorrentResolver interface {
	AddHash(h string)
}

func (s *Server) SeedBytes(name string, b []byte) (*torrent.Torrent, error) {
	ts := TorrentSpecForBytes(name, b)
	t, _, err := s.Client.AddTorrentSpec(ts)
	return t, err
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
		t, new := s.Client.AddTorrentInfoHash(h)
		if new {
			select {
			case <-t.GotInfo():
				s.db.CreateTorrentSearch(t.InfoHash().HexString(), t.Name())
				s.db.SetTorrentMeta(t.InfoHash().HexString(), t.Name())
				log.Printf("Resolved:\t%s\t%s", t.InfoHash().HexString(), t.Name())
			case <-time.After(time.Second * 2):
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

func (s *Server) Run() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-sigs:
				done <- true
			default:
			}

			if s.listen && len(s.hashes) > 0 {
				err := s.resolveHash(s.hashes[0])
				if err != nil {
					log.Println(err)
				}
				s.hashes = s.hashes[1:]
				if !s.listen {
					done <- true
				}
			} else {
				select {
				case <-time.After(time.Second):
				case <-sigs:
					done <- true
				}
			}
		}
	}()
	<-done
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
		Client:       nil,
		hashes:       make([]string, 0),
		resolveCache: cache2go.Cache("resolveCache"),
		listen:       cfg.Listen,
		seed:         cfg.Seed,
		db:           db,
		apiTorrent:   nil,
	}

	dcfg := dht.ServerConfig{StartingNodes: dht.GlobalBootstrapAddrs}
	if s.listen {
		dcfg.OnAnnouncePeer = s.onAnnouncePeer
	}
	if s.seed {
		dcfg.OnQuery = s.onQuery
	}
	torrentCfg := torrent.Config{
		DHTConfig: dcfg,
		Seed:      cfg.Seed,
	}
	cl, err := torrent.NewClient(&torrentCfg)
	id := cl.PeerID()
	log.Printf("Starting Detergent With Peer ID: %s", hex.EncodeToString(id[:]))
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	s.Client = cl
	if s.seed {
		t, err := s.SeedBytes("detergent.json", DetSemaphoreBytes())
		if err != nil {
			panic(err)
		}
		s.apiTorrent = t
		log.Printf("Seeding detergent.json: magnet:?xt=urn:btih:%s\n", s.apiTorrent.InfoHash().HexString())
		go func() {
			for {
				for _, p := range s.apiTorrent.KnownSwarm() {
					if s.seed && !s.listen {
						log.Printf("Maybe Detergent Peer: %s:%d\n", p.IP, p.Port)
					}
				}
				<-time.After(time.Second * 5)
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

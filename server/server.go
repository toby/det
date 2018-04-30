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
	Client         *torrent.Client
	hashes         []string
	db             *SqliteDBClient
	hashLock       sync.Mutex
	resolveCache   *cache2go.CacheTable
	listenAnnounce bool
}

type TorrentResolver interface {
	AddHash(h string)
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

func (s *Server) OnQuery(query *krpc.Msg, source net.Addr) bool {
	s.hashLock.Lock()
	defer s.hashLock.Unlock()
	if query.Q == "get_peers" {
		if err := s.AddHash(hex.EncodeToString(query.A.InfoHash[:])); err != nil {
			log.Printf("Error adding hash: %s", err)
		}
	}
	return true
}

func (s *Server) OnAnnouncePeer(h metainfo.Hash, peer dht.Peer) {
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
				s.db.SetTorrentMeta(t.InfoHash().HexString(), t.Name())
				s.db.CreateTorrentSearch(t.InfoHash().HexString(), t.Name())
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

func (s *Server) Listen() {
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

			if len(s.hashes) > 0 {
				err := s.resolveHash(s.hashes[0])
				if err != nil {
					log.Println(err)
				}
				s.hashes = s.hashes[1:]
				if !s.listenAnnounce {
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
	if s.listenAnnounce {
		log.Printf("Exiting Detergent, here are some stats:")
		s.Client.WriteStatus(os.Stderr)
	}
}

func NewServer(listenAnnounce bool) *Server {
	var dhtCfg dht.ServerConfig
	db := NewSqliteDB("./")
	s := &Server{
		Client:         nil,
		hashes:         make([]string, 0),
		resolveCache:   cache2go.Cache("resolveCache"),
		listenAnnounce: listenAnnounce,
		db:             db,
	}
	if listenAnnounce {
		dhtCfg = dht.ServerConfig{
			StartingNodes: dht.GlobalBootstrapAddrs,
			//OnQuery:       s.OnQuery,
			OnAnnouncePeer: s.OnAnnouncePeer,
		}
	} else {
		dhtCfg = dht.ServerConfig{
			StartingNodes:  dht.GlobalBootstrapAddrs,
			OnAnnouncePeer: func(h metainfo.Hash, p dht.Peer) {},
		}
	}
	cfg := torrent.Config{
		DHTConfig: dhtCfg,
	}
	cl, err := torrent.NewClient(&cfg)
	id := cl.PeerID()
	log.Printf("Starting Detergent With Peer ID: %s", hex.EncodeToString(id[:]))
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	s.Client = cl

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
		if listenAnnounce {
			log.Println("Web stats listening on: http://0.0.0.0:8888")
		}
		log.Fatal(http.ListenAndServe(":8888", nil))
	}()
	return s
}

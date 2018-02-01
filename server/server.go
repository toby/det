package server

import (
	"database/sql"
	"encoding/hex"
	"errors"
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

const ResolveWindow = time.Second * 600

type Server struct {
	Client       *torrent.Client
	hashes       []string
	db           *SqliteDBClient
	hashLock     sync.Mutex
	resolveCache *cache2go.CacheTable
}

type TorrentResolver interface {
	AddHash(h string)
}

func (s *Server) AddHash(h string) error {
	if len(h) == 40 {
		err := s.db.CreateTorrent(h)
		if err != nil {
			// duplicate hash
			if err.(sqlite3.Error).Code == 19 {
				// have we tried to resolve this in the last 10 mins?
				_, err := s.resolveCache.Value(h)
				if err == nil {
					// yes, do nothing
					return nil
				} else {
					// no, try to resolve
					s.resolveCache.Add(h, ResolveWindow, true)
					s.hashes = append(s.hashes, h)
					return nil
				}

			}
			log.Printf("AddHash Error:\t%s", err)
			return err
		} else {
			s.resolveCache.Add(h, ResolveWindow, true)
			s.hashes = append(s.hashes, h)
			return nil
		}
	}
	return errors.New("Invalid hash length")
}

func (s *Server) OnQuery(query *krpc.Msg, source net.Addr) (propagate bool) {
	s.hashLock.Lock()
	defer s.hashLock.Unlock()
	if query.Q == "get_peers" {
		s.AddHash(hex.EncodeToString(query.A.InfoHash[:]))
	}
	return true
}

func (s *Server) OnAnnouncePeer(h metainfo.Hash, peer dht.Peer) {
	s.hashLock.Lock()
	defer s.hashLock.Unlock()
	hx := h.HexString()
	err := s.AddHash(hx)
	if err == nil {
		err := s.db.CreateAnnounce(hx, peer.String())
		if err != nil {
			log.Printf("CreateAnnounce Error:\t%s", err)
		}
	}
}

func (s *Server) WaitForSig() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			if len(s.hashes) > 0 {
				hx := s.hashes[0]
				s.hashes = s.hashes[1:]
				h := metainfo.NewHashFromHex(hx)
				st, err := s.db.GetTorrent(h)
				if err == sql.ErrNoRows || st.ResolvedAt == nil {
					t, new := s.Client.AddTorrentInfoHash(h)
					if new {
						select {
						case <-t.GotInfo():
						case <-time.After(time.Second * 2):
							log.Printf("Timeout:\t%s", h)
						case <-sigs:
							done <- true
						}
					} else {
						log.Printf("Resolved Found:\t%s", t)
					}
					t.Drop()
				} else if err != nil {
					log.Printf("GetTorrent err:\t%s", err)
				} else {
					log.Printf("Found:\t%s\t%s", st.InfoHash.HexString(), st.Name)
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

func NewServer() *Server {
	stor := NewSqliteDB("./")
	s := &Server{
		Client:       nil,
		hashes:       make([]string, 0),
		db:           stor.(*SqliteDBClient),
		resolveCache: cache2go.Cache("resolveCache"),
	}
	cfg := torrent.Config{
		DefaultStorage: stor,
		DHTConfig: dht.ServerConfig{
			StartingNodes: dht.GlobalBootstrapAddrs,
			//OnQuery:       s.OnQuery,
			OnAnnouncePeer: s.OnAnnouncePeer,
		},
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
		log.Println("Web stats listening on: http://0.0.0.0:8888")
		log.Fatal(http.ListenAndServe(":8888", nil))
	}()
	return s
}

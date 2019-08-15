package server

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/muesli/cache2go"
)

// Server is a det peer that contains torrent, dht and det specifc
// functionality.
type Server struct {
	config       *Config
	client       *torrent.Client
	hashes       chan string
	db           *SqliteDBClient
	hashLock     sync.Mutex
	resolveCache *cache2go.CacheTable
	listen       bool
	seed         bool
	peers        []torrent.Peer
	peerEvents   <-chan torrent.Peer
}

// Config tells the Server if it should listen (build a db of resolved announce
// hashes) and seed (share files on the torrent network after they are
// complete). Seeding will also result in participation in the det peer
// discovery protocol.
type Config struct {
	ListenHost      string
	ListenPort      int
	PublicHost      string
	DisableUpnp     bool
	HashQueueLength int
	SqlitePath      string
	BoltDBPath      string
	DownloadPath    string
	Listen          bool
	Seed            bool
	NumResolvers    int
	ResolverTimeout time.Duration
	ResolverWindow  time.Duration
	TorrentDebug    bool
}

// NewServer returns a Server configured with cfg.
func NewServer(cfg *Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Missing server config")
	}
	db, err := NewSqliteDB(cfg.SqlitePath)
	if err != nil {
		return nil, err
	}
	s := &Server{
		config:       cfg,
		client:       nil,
		hashes:       make(chan string, cfg.HashQueueLength),
		resolveCache: cache2go.Cache("resolveCache"),
		listen:       cfg.Listen,
		seed:         cfg.Seed,
		db:           db,
		peers:        make([]torrent.Peer, 0),
		peerEvents:   nil,
	}

	torrentCfg := torrent.NewDefaultClientConfig()
	torrentCfg.ListenHost = func(network string) string { return cfg.ListenHost }
	torrentCfg.ListenPort = cfg.ListenPort
	torrentCfg.NoDefaultPortForwarding = cfg.DisableUpnp
	torrentCfg.Seed = cfg.Seed
	torrentCfg.Debug = cfg.TorrentDebug
	torrentCfg.DataDir = cfg.DownloadPath
	if cfg.PublicHost != "" {
		torrentCfg.PublicIp4 = net.ParseIP(cfg.PublicHost)
		torrentCfg.DisableIPv6 = true
	}
	torrentCfg.DefaultStorage = storage.NewBoltDB(cfg.BoltDBPath)
	if s.listen {
		torrentCfg.DHTOnQuery = s.onQuery
	}
	cl, err := torrent.NewClient(torrentCfg)
	id := cl.PeerID()
	if err != nil {
		return nil, err
	}
	s.client = cl
	log.Printf("Torrent Peer ID: %s", metainfo.HashBytes(id[:]).HexString())
	log.Printf("Listen Address: %s", cfg.ListenHost)
	log.Printf("Listen Port: %d", cfg.ListenPort)
	log.Printf("Public IP: %s", cfg.PublicHost)
	log.Printf("Upnp Enabled: %t", !cfg.DisableUpnp)

	// TODO: reneable peer discovery
	// if s.seed {
	// 	s.peerEvents = StartDiscovery(s)
	// }

	return s, nil
}

// DB returns the servers underlying database.
func (s *Server) DB() *SqliteDBClient {
	return s.db
}

// Run starts the server and waits for SIGINT or SIGTERM.
func (s *Server) Run() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Number of resolvers: %d", s.config.NumResolvers)
	for i := 0; i <= s.config.NumResolvers; i++ {
		go func() {
			for {
				h := <-s.hashes
				err := s.resolveAndStoreHash(h)
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}
	<-sigs
	_ = s.db.Close()
	// log.Printf("Exiting Detergent, here are some stats:")
	// s.client.WriteStatus(os.Stderr)
	s.client.Close()
}

// AddMetaInfo seeds the MetaInfo on the torrent network.
func (s *Server) AddMetaInfo(m *metainfo.MetaInfo) (*torrent.Torrent, error) {
	return s.client.AddTorrent(m)
}

// ResolveHash is a blocking call to resolve metadata for the given hash. Returns nil after
// timeout if no info is found.
func (s *Server) ResolveHash(h metainfo.Hash, timeout time.Duration) *metainfo.Info {
	t, _ := s.client.Torrent(h)
	if t == nil {
		t, _ = s.client.AddTorrentInfoHash(h)
	}
	select {
	case <-t.GotInfo():
		i := t.Info()
		t.Drop()
		return i
	case <-time.After(timeout):
		return nil
	}
}

// DownloadInfoHash will return a channel that closes after the
// specified timeout or returns a *torrent.Torrent if it was able to
// fully download. A timeout of zero will block until the torrent is
// downloaded.
func (s *Server) DownloadInfoHash(h metainfo.Hash, timeout time.Duration, stor *storage.ClientImpl) <-chan *torrent.Torrent {
	out := make(chan *torrent.Torrent)
	go func() {
		var t *torrent.Torrent
		if t, _ = s.client.Torrent(h); t != nil {
			close(out)
			return
		}
		if stor != nil {
			t, _ = s.client.AddTorrentInfoHashWithStorage(h, *stor)
		} else {
			t, _ = s.client.AddTorrentInfoHash(h)
		}
		log.Printf("Downloading: %s", h)
		downloaded := make(chan bool)
		stop := make(chan bool)
		go func() {
			<-t.GotInfo()
			log.Printf("Resolved: %s", t.Name())
			t.DownloadAll()
			for {
				if t.BytesMissing() == 0 {
					downloaded <- true
					return
				}
				br := t.Stats().ConnStats.BytesReadUsefulData
				p := 100 * float64(br.Int64()) / float64(t.Length())
				select {
				case <-time.After(time.Second * 10):
					log.Printf("Downloaded: %.2f%% - %d of %d bytes", p, br, t.Length())
				case <-stop:
					return
				}
			}
		}()
		if timeout == 0 {
			<-downloaded
			out <- t
		} else {
			select {
			case <-downloaded:
				out <- t
			case <-time.After(timeout):
				log.Printf("Download timeout: %s", t.Name())
			}
		}
		stop <- true
		close(stop)
		close(downloaded)
		close(out)
	}()
	return out
}

// PeerID returns the unique id for this peer. It satisfies the Discoverable interface.
func (s *Server) PeerID() metainfo.Hash {
	id := s.client.DhtServers()[0].ID()
	return metainfo.HashBytes(id[:])
}

// Namespace returns the unique namespace for det. It satisfies the Discoverable interface.
func (s *Server) Namespace() string {
	return "detergent"
}

// AddPeer stores the discovered peer in the Server's peer list. It satisfies
// the Discoverable interface.
func (s *Server) AddPeer(p torrent.Peer) {
	s.peers = append(s.peers, p)
}

// TorrentClient returns the underlying torrent client. It satisfies the
// TorrentPeer interface.
func (s *Server) TorrentClient() *torrent.Client {
	return s.client
}

func (s *Server) onQuery(query *krpc.Msg, source net.Addr) bool {
	if query.Q == "announce_peer" {
		s.hashLock.Lock()
		defer s.hashLock.Unlock()
		hx := hex.EncodeToString(query.A.InfoHash[:])
		p := hex.EncodeToString(query.A.ID[:])
		if err := s.addHash(hx); err != nil {
			log.Printf("Error adding hash: %s", err)
			return true
		}
		if err := s.db.CreateAnnounce(hx, p); err != nil {
			log.Printf("CreateAnnounce Error:\t%s", err)
		}
		// have we tried to resolve this in the last 10 mins?
		if !s.resolveCache.Exists(hx) {
			s.resolveCache.Add(hx, s.config.ResolverWindow, true)
			s.hashes <- hx
		}
	}
	return true
}

func (s *Server) addHash(hx string) error {
	if len(hx) != 40 {
		return errors.New("Invalid hash length")
	}
	return s.db.CreateTorrent(hx)
}

func (s *Server) resolveAndStoreHash(hx string) error {
	st, err := s.db.GetTorrent(hx)
	if err == sql.ErrNoRows || st.ResolvedAt.IsZero() {
		h := metainfo.NewHashFromHex(hx)
		t, new := s.client.AddTorrentInfoHashWithStorage(h, make(TorrentBytes, 0))
		if new {
			select {
			case <-t.GotInfo():
				log.Printf("Resolved:\t%s\t%s", hx, t.Name())
				err = s.db.CreateTorrentSearch(hx, t.Name())
				if err != nil {
					return err
				}
				info := t.Info()
				for i, fi := range info.Files {
					for _, p := range fi.Path {
						err = s.db.CreateFileInfo(hx, p, fi.Length, i)
						err = s.db.CreateTorrentSearch(hx, p)
					}
				}
				err = s.db.SetTorrentMeta(hx, t.Name(), t.Length())
				if err != nil {
					return err
				}
			case <-time.After(s.config.ResolverTimeout):
				// log.Printf("Timeout:\t%s", h)
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

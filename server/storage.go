package server

import (
	"database/sql"
	"encoding/binary"
	"io"
	"log"
	"path/filepath"
	"time"

	"github.com/anacrolix/missinggo/x"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/gchaincl/dotsql"
	_ "github.com/mattn/go-sqlite3"
)

var dot, _ = dotsql.LoadFromFile("server/sql/db.sql")

type SqliteDBClient struct {
	db *sql.DB
}

type sqliteDBTorrent struct {
	cl         *SqliteDBClient
	InfoHash   *metainfo.Hash
	Name       string
	CreatedAt  *time.Time
	ResolvedAt *time.Time
}

type RankedTorrent struct {
	AnnounceCount sql.NullInt64
	Name          sql.NullString
	InfoHash      sql.NullString
	CreatedAt     *time.Time
	ResolvedAt    *time.Time
}

type sqliteDBPiece struct {
	io.WriterAt
	io.ReaderAt
	db  *sql.DB
	p   metainfo.Piece
	ih  metainfo.Hash
	key [24]byte
}

type sqlitePieceCompletion struct {
	db *sql.DB
}

type Stats struct {
	Torrents  int64
	Announces int64
	Resolved  int64
}

func NewSqliteDB(filePath string) storage.ClientImpl {
	log.Printf("Using SQLite DB: %ssqlite.db", filePath)
	ret := &SqliteDBClient{}
	var err error
	ret.db, err = sql.Open("sqlite3", filepath.Join(filePath, "sqlite.db"))
	if err != nil {
		panic(err)
	}
	_, err = dot.Exec(ret.db, "create-completed-table")
	if err != nil {
		ret.db.Close()
		panic(err)
	}
	_, err = dot.Exec(ret.db, "create-torrent-table")
	if err != nil {
		ret.db.Close()
		panic(err)
	}
	_, err = dot.Exec(ret.db, "create-search-table")
	if err != nil {
		ret.db.Close()
		panic(err)
	}
	_, err = dot.Exec(ret.db, "create-announce-table")
	if err != nil {
		ret.db.Close()
		panic(err)
	}
	return ret
}

func (me *SqliteDBClient) Close() error {
	return me.db.Close()
}

func (me *SqliteDBClient) Stats() (stats Stats, err error) {
	var torrents, announces, resolved int64
	stats = Stats{torrents, announces, resolved}
	row, err := dot.QueryRow(me.db, "total-torrents")
	if err != nil {
		return
	}
	err = row.Scan(&stats.Torrents)
	if err != nil {
		return
	}
	row, err = dot.QueryRow(me.db, "total-announces")
	if err != nil {
		return
	}
	err = row.Scan(&stats.Announces)
	if err != nil {
		return
	}
	row, err = dot.QueryRow(me.db, "total-resolved")
	if err != nil {
		return
	}
	err = row.Scan(&stats.Resolved)
	return
}

func (me *SqliteDBClient) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	_, err := dot.Exec(me.db, "set-torrent-meta", info.Name, infoHash.HexString())
	if err != nil {
		return nil, err
	}
	_, err = dot.Exec(me.db, "create-torrent-search", infoHash.HexString(), info.Name)
	if err != nil {
		return nil, err
	}
	log.Printf("Resolved:\t%s\t%s", infoHash, info.Name)
	return &sqliteDBTorrent{me, &infoHash, info.Name, nil, nil}, nil
}

func (me *SqliteDBClient) CreateTorrent(hash string) error {
	_, err := dot.Exec(me.db, "create-torrent", hash)
	return err
}

func (me *SqliteDBClient) GetTorrent(infoHash metainfo.Hash) (ret sqliteDBTorrent, err error) {
	ret = sqliteDBTorrent{}
	row, err := dot.QueryRow(me.db, "get-torrent", infoHash.HexString())
	if err != nil {
		return sqliteDBTorrent{me, nil, "", nil, nil}, err
	}
	var h string
	err = row.Scan(&h, &ret.Name, &ret.CreatedAt, &ret.ResolvedAt)
	if err != nil {
		return sqliteDBTorrent{me, nil, "", nil, nil}, err
	} else {
		hash := metainfo.NewHashFromHex(h)
		ret.InfoHash = &hash
	}
	ret.cl = me
	return
}

func (me *SqliteDBClient) PopularTorrents(limit int) (ret []RankedTorrent, err error) {
	ret = make([]RankedTorrent, 0)
	rows, err := dot.Query(me.db, "popular-torrents", limit)
	defer rows.Close()
	if err != nil {
		return ret, err
	}
	for rows.Next() {
		var row RankedTorrent
		err = rows.Scan(&row.AnnounceCount, &row.InfoHash, &row.Name, &row.CreatedAt, &row.ResolvedAt)
		ret = append(ret, row)
	}
	return
}

func (me *SqliteDBClient) SearchTorrents(term string, limit int) (ret []RankedTorrent, err error) {
	ret = make([]RankedTorrent, 0)
	rows, err := dot.Query(me.db, "search-torrents", term, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var row RankedTorrent
		err = rows.Scan(&row.AnnounceCount, &row.InfoHash, &row.Name, &row.CreatedAt, &row.ResolvedAt)
		ret = append(ret, row)
	}
	return
}

func (me *SqliteDBClient) CreateAnnounce(hash string, peerId string) error {
	_, err := dot.Exec(me.db, "create-announce", hash, peerId)
	if err != nil {
		return err
	}
	_, err = dot.Exec(me.db, "update-announce-count", hash)
	return err
}

func (me *sqliteDBTorrent) Piece(p metainfo.Piece) storage.PieceImpl {
	ret := &sqliteDBPiece{
		p:  p,
		db: me.cl.db,
		ih: *me.InfoHash,
	}
	copy(ret.key[:], me.InfoHash[:])
	binary.BigEndian.PutUint32(ret.key[20:], uint32(p.Index()))
	return ret
}

func (me *sqliteDBTorrent) Close() error { return nil }

func (p *sqliteDBPiece) pc() storage.PieceCompletionGetSetter {
	return &sqlitePieceCompletion{p.db}
}

func (p *sqliteDBPiece) pk() metainfo.PieceKey {
	return metainfo.PieceKey{p.ih, p.p.Index()}
}

func (p *sqliteDBPiece) Completion() storage.Completion {
	c, err := p.pc().Get(p.pk())
	x.Pie(err)
	return c
}

func (p *sqliteDBPiece) MarkComplete() error {
	return p.pc().Set(p.pk(), true)
}

func (p *sqliteDBPiece) MarkNotComplete() error {
	return p.pc().Set(p.pk(), false)
}

func (me *sqlitePieceCompletion) Get(pk metainfo.PieceKey) (ret storage.Completion, err error) {
	row, err := dot.QueryRow(me.db, "completed-exists", pk.InfoHash.HexString(), pk.Index)
	if err != nil {
		return
	}
	var c bool
	err = row.Scan(&c)
	ret = storage.Completion{
		Complete: c,
		Ok:       true,
	}
	return
}

func (me *sqlitePieceCompletion) Set(pk metainfo.PieceKey, b bool) (err error) {
	if b {
		_, err = dot.Exec(me.db, "insert-completed", pk.InfoHash.HexString(), pk.Index)
	} else {
		_, err = dot.Exec(me.db, "delete-completed", pk.InfoHash.HexString(), pk.Index)
	}
	return
}

func (me *sqlitePieceCompletion) Close() {
	me.db.Close()
}

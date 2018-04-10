package server

import (
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent/storage"
	"github.com/gchaincl/dotsql"
	_ "github.com/mattn/go-sqlite3"
)

var dot, _ = dotsql.LoadFromFile("server/sql/db.sql")

type SqliteDBClient struct {
	db *sql.DB
}

type Torrent struct {
	AnnounceCount int
	Name          string
	InfoHash      string
	CreatedAt     *time.Time
	ResolvedAt    *time.Time
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

func (me *SqliteDBClient) CreateTorrent(hash string) error {
	_, err := dot.Exec(me.db, "create-torrent", hash)
	return err
}

func (me *SqliteDBClient) GetTorrent(hash string) (ret Torrent, err error) {
	ret = Torrent{}
	row, err := dot.QueryRow(me.db, "get-torrent", hash)
	if err != nil {
		return Torrent{0, "", "", nil, nil}, err
	}
	var h string
	err = row.Scan(&h, &ret.Name, &ret.CreatedAt, &ret.ResolvedAt)
	if err != nil {
		return Torrent{0, "", "", nil, nil}, err
	}

	ret.InfoHash = h
	return
}

func (me *SqliteDBClient) PopularTorrents(limit int) (ret []Torrent, err error) {
	ret = make([]Torrent, 0)
	rows, err := dot.Query(me.db, "popular-torrents", limit)
	defer rows.Close()
	if err != nil {
		return ret, err
	}
	for rows.Next() {
		var row Torrent
		err = rows.Scan(&row.AnnounceCount, &row.InfoHash, &row.Name, &row.CreatedAt, &row.ResolvedAt)
		ret = append(ret, row)
	}
	return
}

func (me *SqliteDBClient) SearchTorrents(term string, limit int) (ret []Torrent, err error) {
	ret = make([]Torrent, 0)
	rows, err := dot.Query(me.db, "search-torrents", term, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var row Torrent
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

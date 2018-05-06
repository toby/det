package server

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

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
	CreatedAt     time.Time
	ResolvedAt    time.Time
}

type Stats struct {
	Torrents  int64
	Announces int64
	Resolved  int64
}

type TimelineEntry struct {
	Day      time.Time
	Torrents []Torrent
}

func scanTorrent(scan func(...interface{}) error) (Torrent, error) {
	st := struct {
		AnnounceCount int
		Name          *string
		InfoHash      string
		CreatedAt     *time.Time
		ResolvedAt    *time.Time
	}{}

	err := scan(&st.AnnounceCount, &st.InfoHash, &st.Name, &st.CreatedAt, &st.ResolvedAt)
	if err != nil {
		return Torrent{}, err
	}

	t := Torrent{
		AnnounceCount: st.AnnounceCount,
		InfoHash:      st.InfoHash,
	}
	if st.Name != nil {
		t.Name = *st.Name
	}
	if st.CreatedAt != nil {
		t.CreatedAt = *st.CreatedAt
	}
	if st.ResolvedAt != nil {
		t.ResolvedAt = *st.ResolvedAt
	}

	return t, nil
}

func NewSqliteDB(filePath string) *SqliteDBClient {
	log.Printf("Using SQLite DB: %ssqlite.db", filePath)
	ret := &SqliteDBClient{}
	var err error
	ret.db, err = sql.Open("sqlite3", filepath.Join(filePath, "sqlite.db"))
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

func (me *SqliteDBClient) GetTorrent(hash string) (Torrent, error) {
	row, err := dot.QueryRow(me.db, "get-torrent", hash)
	if err != nil {
		return Torrent{}, err
	}

	t, err := scanTorrent(row.Scan)
	if err != nil {
		return Torrent{}, err
	}

	return t, nil
}

func (me *SqliteDBClient) PopularTorrents(limit int) ([]Torrent, error) {
	ret := make([]Torrent, 0)
	rows, err := dot.Query(me.db, "popular-torrents", limit)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		t, err := scanTorrent(rows.Scan)
		if err != nil {
			return ret, err
		}
		ret = append(ret, t)
	}

	return ret, nil
}

func (me *SqliteDBClient) TimelineTorrents(days int, limit int) ([]TimelineEntry, error) {
	ret := make([]TimelineEntry, 0)
	d := time.Now()
	df := "-%d days"
	for i := 0; i <= days; i++ {
		ts := make([]Torrent, 0)
		rows, err := dot.Query(me.db, "popular-torrents-day", fmt.Sprintf(df, i), fmt.Sprintf(df, i+1), limit)
		if err != nil {
			return ret, err
		}

		for rows.Next() {
			t, err := scanTorrent(rows.Scan)
			if err != nil {
				return ret, err
			}
			ts = append(ts, t)
		}
		rows.Close()
		ret = append(ret, TimelineEntry{d, ts})
		d = d.Add(time.Hour * -24)
	}

	return ret, nil
}

func (me *SqliteDBClient) SearchTorrents(term string, limit int) ([]Torrent, error) {
	ret := make([]Torrent, 0)
	rows, err := dot.Query(me.db, "search-torrents", term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t, err := scanTorrent(rows.Scan)
		if err != nil {
			return ret, err
		}
		ret = append(ret, t)
	}

	return ret, nil
}

func (me *SqliteDBClient) CreateAnnounce(hash string, peerId string) error {
	_, err := dot.Exec(me.db, "create-announce", hash, peerId)
	if err != nil {
		return err
	}
	_, err = dot.Exec(me.db, "update-announce-count", hash)
	return err
}

func (me *SqliteDBClient) SetTorrentMeta(hash string, name string) error {
	_, err := dot.Exec(me.db, "set-torrent-meta", name, hash)
	return err
}

func (me *SqliteDBClient) CreateTorrentSearch(hash string, name string) error {
	t, err := me.GetTorrent(hash)
	if err != nil {
		return err
	}
	if t.Name == "" {
		_, err = dot.Exec(me.db, "create-torrent-search", hash, name)
		return err
	}
	return nil
}

package server

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

type SqliteDBClient struct {
	db *sql.DB
}

type FileInfo struct {
	Path     string
	Length   int64
	Index    int8
	InfoHash string
}

type Torrent struct {
	AnnounceCount int
	Name          string
	InfoHash      string
	Length        int64
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
		Length        int64
		CreatedAt     *time.Time
		ResolvedAt    *time.Time
	}{}

	err := scan(&st.AnnounceCount, &st.InfoHash, &st.Name, &st.Length, &st.CreatedAt, &st.ResolvedAt)
	if err != nil {
		return Torrent{}, err
	}

	t := Torrent{
		AnnounceCount: st.AnnounceCount,
		InfoHash:      st.InfoHash,
		Length:        st.Length,
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

func NewSqliteDB(filePath string) (*SqliteDBClient, error) {
	log.Printf("Using SQLite DB: %ssqlite.db", filePath)
	ret := &SqliteDBClient{}
	var err error
	ret.db, err = sql.Open("sqlite3", filepath.Join(filePath, "sqlite.db"))
	if err != nil {
		return nil, err
	}
	_, err = ret.db.Exec(sqlCreateTorrentTable)
	if err != nil {
		ret.db.Close()
		return nil, err
	}
	_, err = ret.db.Exec(sqlCreateFileInfoTable)
	if err != nil {
		ret.db.Close()
		return nil, err
	}
	_, err = ret.db.Exec(sqlCreateSearchTable)
	if err != nil {
		ret.db.Close()
		return nil, err
	}
	_, err = ret.db.Exec(sqlCreateAnnounceTable)
	if err != nil {
		ret.db.Close()
		return nil, err
	}
	return ret, nil
}

func (me *SqliteDBClient) Close() error {
	return me.db.Close()
}

func (me *SqliteDBClient) Stats() (*Stats, error) {
	var torrents, announces, resolved int64
	stats := &Stats{torrents, announces, resolved}
	row := me.db.QueryRow(sqlTotalTorrents)
	err := row.Scan(&stats.Torrents)
	if err != nil {
		return nil, err
	}
	row = me.db.QueryRow(sqlTotalAnnounces)
	err = row.Scan(&stats.Announces)
	if err != nil {
		return nil, err
	}
	row = me.db.QueryRow(sqlTotalResolved)
	err = row.Scan(&stats.Resolved)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (me *SqliteDBClient) CreateTorrent(hash string) error {
	_, err := me.db.Exec(sqlCreateTorrent, hash)
	return err
}

func (me *SqliteDBClient) GetTorrent(hash string) (Torrent, error) {
	row := me.db.QueryRow(sqlGetTorrent, hash)
	t, err := scanTorrent(row.Scan)
	if err != nil {
		return Torrent{}, err
	}
	return t, nil
}

func (me *SqliteDBClient) GetFileInfo(hash string) ([]*FileInfo, error) {
	ret := make([]*FileInfo, 0)
	rows, err := me.db.Query(sqlGetFileInfo, hash)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		f := FileInfo{}
		err = rows.Scan(&f.Path, &f.InfoHash, &f.Length, &f.Index)
		if err != nil {
			return ret, err
		}
		ret = append(ret, &f)
	}
	return ret, nil
}

func (me *SqliteDBClient) PopularTorrents(limit int) ([]Torrent, error) {
	ret := make([]Torrent, 0)
	rows, err := me.db.Query(sqlPopularTorrents, limit)
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
		rows, err := me.db.Query(sqlPopularTorrentsDay, fmt.Sprintf(df, i), fmt.Sprintf(df, i+1), limit)
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
	rows, err := me.db.Query(sqlSearchTorrents, strings.ToLower(term), limit)
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

func (me *SqliteDBClient) CreateFileInfo(hash string, path string, length int64, index int) error {
	_, err := me.db.Exec(sqlCreateFileInfo, hash, path, length, index)
	return err
}

func (me *SqliteDBClient) CreateAnnounce(hash string, peerId string) error {
	_, err := me.db.Exec(sqlCreateAnnounce, hash, peerId)
	if err != nil {
		return err
	}
	_, err = me.db.Exec(sqlUpdateAnnounceCount, hash)
	return err
}

func (me *SqliteDBClient) SetTorrentMeta(hash string, name string, length int64) error {
	_, err := me.db.Exec(sqlSetTorrentMeta, name, length, hash)
	return err
}

func (me *SqliteDBClient) CreateTorrentSearch(hash string, name string) error {
	_, err := me.db.Exec(sqlCreateTorrentSearch, hash, strings.ToLower(name))
	return err
}

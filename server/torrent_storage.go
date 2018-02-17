package server

import (
	"database/sql"
	"encoding/binary"
	"io"
	"log"
	"time"

	"github.com/anacrolix/missinggo/x"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteDBTorrent struct {
	cl         *SqliteDBClient
	InfoHash   *metainfo.Hash
	Name       string
	CreatedAt  *time.Time
	ResolvedAt *time.Time
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

func (me *sqliteDBTorrent) Close() error { return nil }

func (me *sqlitePieceCompletion) Close() {
	me.db.Close()
}

package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

var (
	BuiltinAnnounceList = [][]string{
		{"udp://tracker.openbittorrent.com:80"},
		{"udp://tracker.publicbt.com:80"},
		{"udp://tracker.istole.it:6969"},
	}
)

// Read-Only torrent/storage implementation for []byte
type TorrentBytes []byte

// ClientImpl
func (b TorrentBytes) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	fmt.Println("ClientImpl.OpenTorrent()")
	return b, nil
}

// TorrentImpl
func (b TorrentBytes) Piece(p metainfo.Piece) storage.PieceImpl {
	fmt.Println("TorrentImpl.Piece()")
	fmt.Printf("Piece Index: %d, Piece Length %d, Piece Offset %d\n", p.Index(), p.Length(), p.Offset())
	off := p.Offset()
	l := off + p.Length()
	return b[off:l]
}

func (b TorrentBytes) Close() error {
	fmt.Println("TorrentImpl.Close()")
	return nil
}

// PieceImpl
func (b TorrentBytes) MarkComplete() error {
	fmt.Println("PieceImpl.MarkComplete()")
	return nil
}

func (b TorrentBytes) MarkNotComplete() error {
	fmt.Println("PieceImpl.MarkNotComplete()")
	return nil
}

func (b TorrentBytes) Completion() storage.Completion {
	fmt.Println("PieceImpl.Completion()")
	return storage.Completion{true, true}
}

// io.ReaderAt
func (b TorrentBytes) ReadAt(p []byte, off int64) (n int, err error) {
	fmt.Println("io.ReaderAt.ReadAt()")
	if off >= int64(len(b)) {
		return 0, errors.New("Offset too large")
	}
	n = copy(p, b[off:])
	return n, nil
}

// io.WriterAt
func (b TorrentBytes) WriteAt(p []byte, off int64) (n int, err error) {
	fmt.Println("io.WriterAt.WriteAt()")
	return 0, nil
}

func InfoForBytes(name string, b TorrentBytes) (*metainfo.Info, storage.ClientImpl) {
	info := metainfo.Info{
		Name:        name,
		Length:      int64(len(b)),
		Files:       nil,
		PieceLength: 256 * 1024,
	}
	info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(b)), nil
	})
	return &info, b
}

func TorrentSpecForBytes(name string, b TorrentBytes) *torrent.TorrentSpec {
	i, s := InfoForBytes(name, b)
	mi := &metainfo.MetaInfo{AnnounceList: BuiltinAnnounceList}
	mi.SetDefaults()
	mi.InfoBytes, _ = bencode.Marshal(i)
	ts := torrent.TorrentSpecFromMetaInfo(mi)
	ts.Storage = s
	return ts
}

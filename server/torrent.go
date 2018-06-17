package server

import (
	"bytes"
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
	return b, nil
}

// TorrentImpl
func (b TorrentBytes) Piece(p metainfo.Piece) storage.PieceImpl {
	off := p.Offset()
	l := off + p.Length()
	if len(b) >= int(off) && len(b) >= int(l) {
		return b[off:l]
	}
	return make(TorrentBytes, 0)
}

func (b TorrentBytes) Close() error {
	return nil
}

// PieceImpl
func (b TorrentBytes) MarkComplete() error {
	return nil
}

func (b TorrentBytes) MarkNotComplete() error {
	return nil
}

func (b TorrentBytes) Completion() storage.Completion {
	if len(b) == 0 {
		return storage.Completion{false, false}
	} else {
		return storage.Completion{true, true}
	}
}

// io.ReaderAt
func (b TorrentBytes) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b)) {
		return 0, nil
	}
	n = copy(p, b[off:])
	return n, nil
}

// io.WriterAt
func (b TorrentBytes) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func torrentSpecForInfo(i *metainfo.Info, s storage.ClientImpl) *torrent.TorrentSpec {
	mi := &metainfo.MetaInfo{AnnounceList: BuiltinAnnounceList}
	mi.SetDefaults()
	ib, err := bencode.Marshal(i)
	if err != nil {
		panic(err)
	}
	mi.InfoBytes = ib
	ts := torrent.TorrentSpecFromMetaInfo(mi)
	ts.Storage = s
	return ts
}

func (b TorrentBytes) TorrentInfo(name string) *metainfo.Info {
	info := metainfo.Info{
		Name:        name,
		Length:      int64(len(b)),
		Files:       nil,
		PieceLength: 256 * 1024,
	}
	info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(b)), nil
	})
	return &info
}

func (b TorrentBytes) TorrentSpec(name string) *torrent.TorrentSpec {
	i := b.TorrentInfo(name)
	ts := torrentSpecForInfo(i, b)
	return ts
}

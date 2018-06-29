package server

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

// TorrentBytes is a read-only `torrent/storage` implementation for []byte. It
// can be used to seed data from memory.
type TorrentBytes []byte

// TorrentInfo creates a Torrent info dictionary from the receiver.
func (b TorrentBytes) TorrentInfo(name string) *metainfo.Info {
	info := metainfo.Info{
		Name:        name,
		Length:      int64(len(b)),
		Files:       nil,
		PieceLength: 256 * 1024,
	}
	err := info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(b)), nil
	})
	if err != nil {
		panic(err)
	}
	return &info
}

// TorrentSpec creates a TorrentSpec ready for seeding from the receiver.
func (b TorrentBytes) TorrentSpec(name string) *torrent.TorrentSpec {
	i := b.TorrentInfo(name)
	ts := torrentSpecForInfo(i, b)
	return ts
}

// OpenTorrent is a noop method that returns the receiving bytes. It satisfies
// ClientImpl in the `torrent/storage` package.
func (b TorrentBytes) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	return b, nil
}

// Piece returns a slice of the receiver from the metainfo.Piece offset. Since
// the returned slice is also a TorrentBytes implementation, it can satisfy
// PieceImpl.  Piece satisfies TorrentImpl in the `torrent/storage` package.
func (b TorrentBytes) Piece(p metainfo.Piece) storage.PieceImpl {
	off := p.Offset()
	l := off + p.Length()
	if len(b) >= int(off) && len(b) >= int(l) {
		return b[off:l]
	}
	return make(TorrentBytes, 0)
}

// Close is a noop for TorrentBytes. It satisfies TorrentImpl in the
// `torrent/storage` package.
func (b TorrentBytes) Close() error {
	return nil
}

// MarkComplete is a noop for TorrentBytes. It satisfies PieceImpl in the
// `torrent/storage` package.
func (b TorrentBytes) MarkComplete() error {
	return nil
}

// MarkNotComplete is anoop for TorrentBytes. It satisfies PieceImpl in the
// `torrent/storage` package.
func (b TorrentBytes) MarkNotComplete() error {
	return nil
}

// Completion lets the Torrent client know if this file is complete and ok. If
// the length of the receiver is zero, this returns not ok, and not complete.
// Otherwise it reutrns ok and complete. You may want to use a zero length
// TorrentBytes when you only want to resolve the torrent metadata and not
// download the content. Completion satisfies PieceImpl in the
// `torrent/storage` package.
func (b TorrentBytes) Completion() storage.Completion {
	if len(b) == 0 {
		return storage.Completion{
			Complete: false,
			Ok:       false,
		}
	}
	return storage.Completion{
		Complete: true,
		Ok:       true,
	}
}

// ReadAt reads into p from the supplied offset. It returns zero bytes read and
// a nil error in the case or the offset being larger than the length of the
// receiver. This is so that metadata can be resolved for TorrentBytes without
// causing errors in the torrent client. It's pretty hacky. ReadAt satisfies
// io.ReaderAt.
func (b TorrentBytes) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b)) {
		return 0, nil
	}
	n = copy(p, b[off:])
	return n, nil
}

// WriteAt is a noop for TorrentBytes. It satisfies io.WriterAt.
func (b TorrentBytes) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

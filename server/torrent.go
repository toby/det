package server

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

var (
	// BuiltinAnnounceList are the default trackers when creating a torrent
	// MetaInfo.
	BuiltinAnnounceList = [][]string{
		{"udp://tracker.openbittorrent.com:80"},
		{"udp://tracker.publicbt.com:80"},
		{"udp://tracker.istole.it:6969"},
	}
)

func seedTorrentSpec(cl *torrent.Client, ts *torrent.TorrentSpec) (*torrent.Torrent, error) {
	t, _, err := cl.AddTorrentSpec(ts)
	return t, err
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

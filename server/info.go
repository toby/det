package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/anacrolix/torrent/metainfo"
)

func InfoForBytes(name string, b []byte) (*metainfo.Info, error) {
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
		return nil, fmt.Errorf("error generating pieces: %s", err)
	}
	return &info, nil
}

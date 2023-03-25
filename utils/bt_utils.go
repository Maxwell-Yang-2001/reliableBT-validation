package utils

import (
	"testing"

	rbt "github.com/anacrolix/torrent"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

// Create a magnet link for an existing file, and add it to the seeder client, then return the magnet link.
func MakeMagnet(t *testing.T, seeder *rbt.Client, filePath string, trackers [][]string) string {
	mi := metainfo.MetaInfo{AnnounceList: trackers}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: PieceLength}
	err := info.BuildFromFilePath(filePath)
	require.NoError(t, err)
	mi.InfoBytes, err = bencode.Marshal(info)
	require.NoError(t, err)
	magnet := mi.Magnet(nil, &info).String()
	tr, err := seeder.AddTorrent(&mi)
	require.NoError(t, err)
	require.True(t, tr.Seeding())
	tr.VerifyData()
	return magnet
}

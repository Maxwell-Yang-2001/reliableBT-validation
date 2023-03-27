package utils

import (
	"path/filepath"
	"testing"

	rbt "github.com/anacrolix/torrent"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

// Create a magnet link for a newly created file within seeder's directory and specified size,
// and add it to the seeder client, then return the magnet link.
func CreateFileAndMagnet(t *testing.T, seeder *rbt.Client, dir, name string, size int64, trackers [][]string) string {
	// Create the file first
	filePath := filepath.Join(dir, name)
	CreateFile(t, filePath, size)

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

// Create a magnet link for an existing file, and add it to the seeder client, then return the magnet link.
func MakeMagnet(t *testing.T, seeder *rbt.Client, dir, name string, trackers [][]string) string {
	mi := metainfo.MetaInfo{AnnounceList: trackers}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: PieceLength}
	err := info.BuildFromFilePath(filepath.Join(dir, name))
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

package utils

import (
	"path/filepath"
	"testing"

	rbt "github.com/anacrolix/torrent"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

// Create a file within each seeder's directory and specified size, return the metaInfo to be added to any client.
func CreateFileAndMetaInfo(t *testing.T, seederDirs []string, name string, size int64, trackers [][]string) metainfo.MetaInfo {
	// Create the files first - one for each seeder
	require.NotZero(t, len(seederDirs))
	filePath := ""
	for _, seederDir := range seederDirs {
		filePath = filepath.Join(seederDir, name)
		CreateFile(t, filePath, size)
	}

	mi := metainfo.MetaInfo{AnnounceList: trackers}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: PieceLength}
	err := info.BuildFromFilePath(filePath)
	require.NoError(t, err)
	mi.InfoBytes, err = bencode.Marshal(info)
	require.NoError(t, err)
	return mi
}

// Test the initial setup of a seeder torrent.
func TestSeederInitial(t *testing.T, tr rbt.Torrent, err error) {
	require.NoError(t, err)
	require.True(t, tr.Seeding())
	tr.VerifyData()
}

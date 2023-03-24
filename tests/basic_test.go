package tests

import (
	"path/filepath"
	"rbtValidation/tests/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

// Create a magnet link for an existing file, and add it to the seeder client, then return the magnet link.
func makeMagnet(t *testing.T, seeder *rbt.Client, dir, name string, trackers [][]string) string {
	mi := metainfo.MetaInfo{AnnounceList: trackers}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: 256 * 1024}
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

// Create the configuration for a seeder.
func seederConfig() (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.Seed = true
	config.DataDir = "./seeder"
	config.NoUpload = false
	config.NoDHT = true
	config.ListenPort = 0
	return
}

// Create the configuration for a leecher
func leecherConfig() (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.ListenPort = 0
	config.DataDir = "./leecher"
	config.NoDHT = true
	return
}

// Test whether a seeder can transfer file to a leecher successfully by directly feeding the seeder as a peer for the leecher.
func TestSeederLeecher(t *testing.T) {
	// Create a seeder
	seederConfig := seederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a magnet link and add it to the seeder (note that there is no tracker info in the magnet)
	magnetLink := makeMagnet(t, seeder, seederConfig.DataDir, "hello.txt", [][]string{})

	// Create a leecher
	leecher, _ := rbt.NewClient(leecherConfig())
	defer leecher.Close()

	// Also attach the magnet link to the leecher (and directly given the seeder as peer)
	leecher_torrent, _ := leecher.AddMagnet(magnetLink)
	leecher_torrent.AddClientPeer(seeder)
	<-leecher_torrent.GotInfo()

	// Wait until transfer is complete
	leecher_torrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, "seeder/hello.txt", []string{"leecher/hello.txt"})
}

// Test whether a seeder can transfer file to a leecher successfully by tracker letting them discover each other.
func TestSeederLeecherTracker(t *testing.T) {
	// Create a seeder
	seederConfig := seederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a magnet link and add it to the seeder (tracker on localhost is attached in the magnet)
	magnetLink := makeMagnet(t, seeder, seederConfig.DataDir, "C#.pdf", [][]string{{"http://127.0.0.1:1337/announce"}})

	// Create a leecher
	leecher, _ := rbt.NewClient(leecherConfig())
	defer leecher.Close()

	// Also attach the magnet link to the leecher
	leecher_torrent, _ := leecher.AddMagnet(magnetLink)
	<-leecher_torrent.GotInfo()

	// Wait until transfer is complete
	leecher_torrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, "seeder/C#.pdf", []string{"leecher/C#.pdf"})
}

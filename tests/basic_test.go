package tests

import (
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

// Create the configuration for a seeder.
func SeederConfig() (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.Seed = true
	config.DataDir = "./seeder"
	config.NoUpload = false
	config.NoDHT = true
	config.DisableTCP = false
	config.ListenPort = 0
	return
}

// Create the configuration for a leecher
func LeecherConfig() (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.ListenPort = 0
	config.DataDir = "./leecher"
	config.NoDHT = true
	config.DisableTCP = false
	return
}

// Test whether a seeder can transfer file to a leecher successfully by directly feeding the seeder as a peer for the leecher.
func TestSeederLeecher(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a magnet link and add it to the seeder (note that there is no tracker info in the magnet)
	magnetLink := utils.MakeMagnet(t, seeder, utils.TestFile1kb, [][]string{})

	// Create a leecher
	leecher, _ := rbt.NewClient(LeecherConfig())
	defer leecher.Close()

	// Also attach the magnet link to the leecher (and directly given the seeder as peer)
	leecher_torrent, _ := leecher.AddMagnet(magnetLink)
	leecher_torrent.AddClientPeer(seeder)
	<-leecher_torrent.GotInfo()

	// Wait until transfer is complete
	leecher_torrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFile1kb, []string{"leecher/test_1kb.txt"})
}

// Test whether a seeder can transfer file to a leecher successfully by tracker letting them discover each other.
func TestSeederLeecherTracker(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a magnet link and add it to the seeder (tracker on localhost is attached in the magnet)
	magnetLink := utils.MakeMagnet(t, seeder, utils.TestFile1mb, [][]string{{"http://127.0.0.1:1337/announce"}})

	// Create a leecher
	leecher, _ := rbt.NewClient(LeecherConfig())
	defer leecher.Close()

	// Also attach the magnet link to the leecher
	leecher_torrent, _ := leecher.AddMagnet(magnetLink)
	<-leecher_torrent.GotInfo()

	// Wait until transfer is complete
	leecher_torrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFile1mb, []string{"leecher/test_1mb.txt"})
}

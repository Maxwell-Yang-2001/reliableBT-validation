package tests

import (
	"os"
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
	defer os.RemoveAll(seederConfig.DataDir)

	// Create the test file, a magnet link to it and add to the seeder (note that there is no tracker info in the magnet)
	magnetLink := utils.CreateFileAndMagnet(t, seeder, seederConfig.DataDir, utils.TestFileName, 1e3, [][]string{})

	// Create a leecher
	leecherConfig := LeecherConfig()
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the magnet link to the leecher (and directly given the seeder as peer)
	leecherTorrent, _ := leecher.AddMagnet(magnetLink)
	leecherTorrent.AddClientPeer(seeder)
	<-leecherTorrent.GotInfo()

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

// Test whether a seeder can transfer file to a leecher successfully by tracker letting them discover each other.
func TestSeederLeecherTracker(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file, a magnet link and add it to the seeder (tracker on localhost is attached in the magnet)
	magnetLink := utils.CreateFileAndMagnet(t, seeder, seederConfig.DataDir, utils.TestFileName, 1e6, [][]string{{utils.TestTrackerAnnounceUrl}})

	// Create a leecher
	leecherConfig := LeecherConfig()
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the magnet link to the leecher
	leecherTorrent, _ := leecher.AddMagnet(magnetLink)
	<-leecherTorrent.GotInfo()

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

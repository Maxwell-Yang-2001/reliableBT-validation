package tests

import (
	"fmt"
	"os"
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

// Create the configuration for a seeder.
func SeederConfig(id int, listenPort int) (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.Seed = true
	config.DataDir = fmt.Sprintf("./seeder%d", id)
	config.NoUpload = false
	config.NoDHT = true
	config.DisableTCP = false
	config.ListenPort = listenPort
	return
}

// Create the configuration for a baseline provider.
func BaselineProviderConfig(id int, listenPort int) (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.Seed = true
	config.DataDir = fmt.Sprintf("./baselineProvider%d", id)
	config.NoUpload = false
	config.NoDHT = true
	config.DisableTCP = false
	config.ListenPort = listenPort
	config.Reliable = true
	return
}

// Create the configuration for a leecher
func LeecherConfig(id int, listenPort int) (config *rbt.ClientConfig) {
	config = rbt.NewDefaultClientConfig()
	config.DataDir = fmt.Sprintf("./leecher%d", id)
	config.NoDHT = true
	config.DisableTCP = false
	config.ListenPort = listenPort
	return
}

// Test whether a seeder can transfer file to a leecher successfully by directly feeding the seeder as a peer for the leecher.
func TestSeederLeecher(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 1e3, [][]string{})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the metaInfo to the leecher (and directly given the seeder as peer)
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent.AddClientPeer(seeder)
	<-leecherTorrent.GotInfo()

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify baseline provider
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent}, nil)

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

// Test whether a seeder can transfer file to a leecher successfully by tracker letting them discover each other.
func TestSeederLeecherTracker(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 1e6, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the metaInfo to the leecher
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	<-leecherTorrent.GotInfo()

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify baseline provider
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent}, nil)

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

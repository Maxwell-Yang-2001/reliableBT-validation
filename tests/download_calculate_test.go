package tests

import (
	"os"
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

// This function is extremely similar to TestSeederLeecherTracker in basic_test.go with most notable difference being smaller announce interval
// When testing this function, please configure the tracker with base: 1s and fuzz: 0s to reduce the announce interval
func TestDownloadSpeed(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(3000)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, *seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(3001)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the metaInfo to the leecher
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent.SmallIntervalAllowed = true
	<-leecherTorrent.GotInfo()

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

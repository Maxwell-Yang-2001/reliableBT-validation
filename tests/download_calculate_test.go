package tests

import (
	"os"
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

func TestDownloadSpeed(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(3000)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file, a magnet link and add it to the seeder (tracker on localhost is attached in the magnet)
	magnetLink := utils.CreateFileAndMagnet(t, seeder, seederConfig.DataDir, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})

	// Create a leecher
	leecherConfig := LeecherConfig(3001)
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

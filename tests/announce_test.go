package tests

import (
	"os"
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

// IMPORTANT:
// 1. When testing this file, please configure the tracker with base: 1s and fuzz: 0s to reduce the announce interval
// 2. Please run each test individually, with the tracker reset between every two tests - we want a clean tracker state before each test.
// To do this, you can run `bash restart.sh` in the root of the tracker repo.

// Test whether periodic announcement is made from each peer to the tracker.
// This test requires Wireshark to be capturing on loopback and observe on the periodic requests. Each peer should have an announce per 1s.
func TestBasicAnnounce(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 0)
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

	// Verify baseline provider
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent}, nil)

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

func TestBaselineProviderAnnounce(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a baseline provider (PORT 4000 is a known trusted source by the tracker)
	baselineProviderPort := 4000
	baselineProviderConfig := BaselineProviderConfig(0, baselineProviderPort)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a test file within the seeder and baseline provider dir and add it to both clients
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)

	baselineProviderTorrent, err := baselineProvider.AddTorrent(&metaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 0)
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

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent}, []int{baselineProviderPort})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

func TestFakeBaselineProviderAnnounce(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a "fraud" baseline provider (PORT 4500 is a NOT a known trusted source by the tracker)
	baselineProviderConfig := BaselineProviderConfig(0, 4500)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a test file within the seeder and baseline provider dir and add it to both clients
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)

	baselineProviderTorrent, err := baselineProvider.AddTorrent(&metaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 0)
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

	// Verify baseline provider (as bad actors are ignored, no one should have this info)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent, baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

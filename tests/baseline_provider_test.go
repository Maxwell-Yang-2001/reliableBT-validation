package tests

import (
	"os"
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

func TestCompleteBaselineProviderDisallowOutgoingConn(t *testing.T) {
	// Create a baseline provider (PORT 4000 is a known trusted source by the tracker)
	baselineProviderPort := 4000
	baselineProviderConfig := BaselineProviderConfig(0, baselineProviderPort)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a test file within the baseline provider dir and add it to the baseline provider
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{baselineProviderConfig.DataDir}, utils.TestFileName, 1e9, [][]string{{}})

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

	// Directly promote leecher to the baseline provider
	// In expectation this should not initiate any connection
	baselineProviderTorrent.AddClientPeer(leecher)

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent, leecherTorrent}, []int{})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})
}

func TestCompleteBaselineProviderAllowIngoingConn(t *testing.T) {
	// Create a baseline provider (PORT 4000 is a known trusted source by the tracker)
	baselineProviderPort := 4000
	baselineProviderConfig := BaselineProviderConfig(0, baselineProviderPort)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a test file within the baseline provider dir and add it to the baseline provider
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{baselineProviderConfig.DataDir}, utils.TestFileName, 1e9, [][]string{{}})

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

	// Directly promote leecher to the baseline provider
	// In expectation this should not initiate any connection
	leecherTorrent.AddClientPeer(baselineProvider)

	// Wait until transfer is complete
	leecherTorrent.DownloadAll()
	leecher.WaitAll()

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent, leecherTorrent}, []int{})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})
}

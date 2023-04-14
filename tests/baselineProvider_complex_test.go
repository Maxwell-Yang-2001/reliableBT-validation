package tests

import (
	"fmt"
	"os"
	"rbtValidation/utils"
	"testing"
	"time"

	rbt "github.com/anacrolix/torrent"
	"github.com/stretchr/testify/require"
)

// Starts with a seeder and an empty leecher, runs for 3s,
// kills the seeder, starts the baseline provider (which starts with the complete file).
// Expectation: the leecher should be able to finish the rest of the download with the baseline provider.
func TestSeederWaitAndDieHandOverToBaselineProvider(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seederConfig.UploadRateLimiter = utils.SmallRateLimiter
	seeder, _ := rbt.NewClient(seederConfig)
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a baseline provider (PORT 4000 is a known trusted source by the tracker)
	baselineProviderPort := 4000
	baselineProviderConfig := BaselineProviderConfig(0, baselineProviderPort)
	utils.CreateDir(t, baselineProviderConfig.DataDir)

	// Create a test file within the seeder and baseline provider dir and add it to both clients
	utils.CreateFilesInDirs(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 1e7)
	metaInfo := utils.CreateMetaInfo(t, seederConfig.DataDir, utils.TestFileName, [][]string{{utils.TestTrackerAnnounceUrl}})
	trackerlessMetaInfo := utils.CreateMetaInfo(t, seederConfig.DataDir, utils.TestFileName, [][]string{})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

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

	leecherTorrent.DownloadAll()

	// Sleep for 3 seconds and close seeder
	time.Sleep(3 * time.Second)
	seederUploadedBytes := seederTorrent.UploadedBytes()
	fmt.Println("Seeder Uploaded Bytes: ", seederUploadedBytes)
	fmt.Println("Leecher Downloaded Bytes: ", seederTorrent.DownloadedBytes())

	seeder.Close()

	// Start baseline provider
	baselineProviderTorrent, err := baselineProvider.AddTorrent(&trackerlessMetaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	// Let it process that it has the complete file,
	// So it will promote itself to the tracker as a complete baseline provider right away
	time.Sleep(3 * time.Second)
	baselineProviderTorrent.AddTrackers([][]string{{utils.TestTrackerAnnounceUrl}})

	// Wait until transfer is complete
	leecher.WaitAll()
	baselineProviderUploadedBytes := baselineProviderTorrent.UploadedBytes()
	fmt.Println("Baseline Provider Uploaded Bytes: ", baselineProviderUploadedBytes)
	fmt.Println("Baseline Provider Downloaded Bytes: ", baselineProviderTorrent.DownloadedBytes())

	// Verify that baseline provider does contribute to the leecher
	if seederUploadedBytes < 2e9 {
		require.NotZero(t, baselineProviderUploadedBytes)
	}

	// Verify baseline provider (baseline provider should not get itself as baseline provider, but everyone else should)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{leecherTorrent}, []int{baselineProviderPort})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

// Starts with a seeder and an empty leecher, runs for 3s,
// starts the baseline provider (which starts with the complete file).
// Expectation: the leecher should be able to finish the rest of the download with both the seeder and the baseline provider.
func TestSeederWaitAndBaselineProviderJoin(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a baseline provider (PORT 4000 is a known trusted source by the tracker)
	baselineProviderPort := 4000
	baselineProviderConfig := BaselineProviderConfig(0, baselineProviderPort)
	utils.CreateDir(t, baselineProviderConfig.DataDir)

	// Create a test file within the seeder and baseline provider dir and add it to both clients
	utils.CreateFilesInDirs(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 2e9)
	metaInfo := utils.CreateMetaInfo(t, seederConfig.DataDir, utils.TestFileName, [][]string{{utils.TestTrackerAnnounceUrl}})
	trackerlessMetaInfo := utils.CreateMetaInfo(t, seederConfig.DataDir, utils.TestFileName, [][]string{})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

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

	leecherTorrent.DownloadAll()

	// Sleep for 3 seconds and close seeder
	time.Sleep(3 * time.Second)
	seederUploadedBytes := seederTorrent.UploadedBytes()
	fmt.Println("Seeder Uploaded Bytes after 3 seconds: ", seederUploadedBytes)
	fmt.Println("Seeder Downloaded Bytes after 3 seconds: ", seederTorrent.DownloadedBytes())

	// Start baseline provider
	baselineProviderTorrent, err := baselineProvider.AddTorrent(&trackerlessMetaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	// Let it process that it has the complete file,
	// So it will promote itself to the tracker as a complete baseline provider right away
	time.Sleep(3 * time.Second)
	baselineProviderTorrent.AddTrackers([][]string{{utils.TestTrackerAnnounceUrl}})

	// Wait until transfer is complete
	leecher.WaitAll()
	baselineProviderUploadedBytes := baselineProviderTorrent.UploadedBytes()
	fmt.Println("Baseline Provider Uploaded Bytes: ", baselineProviderUploadedBytes)
	fmt.Println("Baseline Provider Downloaded Bytes: ", baselineProviderTorrent.DownloadedBytes())

	// Verify that baseline provider does contribute to the leecher
	if seederUploadedBytes < 2e9 {
		require.NotZero(t, baselineProviderUploadedBytes)
	}

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{leecherTorrent}, []int{baselineProviderPort})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

package tests

import (
	"fmt"
	"os"
	"rbtValidation/utils"
	"testing"
	"time"

	rbt "github.com/anacrolix/torrent"
)

// start with a seeder and an empty leecher, runs for 3s,
// kills the seeder, starts the bp (it has the complete file as well).
func TestSeederWaitAndDiesHandOverToBaseLineProvider(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 3000)
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
	// metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 4030)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the metaInfo to the leecher
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent.SmallIntervalAllowed = true
	<-leecherTorrent.GotInfo()

	leecherTorrent.DownloadAll()

	// sleep for 3 seconds and close seeder
	time.Sleep(3 * time.Second)
	fmt.Println("Uploaded Bytes ", seederTorrent.UploadedBytes())
	fmt.Println("Downloaded Bytes: ", seederTorrent.DownloadedBytes())

	seeder.Close()

	// start bp
	baselineProviderTorrent, err := baselineProvider.AddTorrent(&trackerlessMetaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	time.Sleep(3 * time.Second)
	baselineProviderTorrent.AddTrackers([][]string{{utils.TestTrackerAnnounceUrl}})

	// Wait until transfer is complete
	leecher.WaitAll()
	fmt.Println("Uploaded Bytes ", baselineProviderTorrent.UploadedBytes())
	fmt.Println("Downloaded Bytes: ", baselineProviderTorrent.DownloadedBytes())

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{leecherTorrent}, []int{baselineProviderPort})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

func TestSeederWaitAndBaseLineProviderJoin(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig(0, 3000)
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
	// metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir, baselineProviderConfig.DataDir}, utils.TestFileName, 2e9, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	seederTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, seederTorrent, err)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	// Create a leecher
	leecherConfig := LeecherConfig(0, 4030)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	// Also attach the metaInfo to the leecher
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent.SmallIntervalAllowed = true
	<-leecherTorrent.GotInfo()

	leecherTorrent.DownloadAll()

	// sleep for 3 seconds and close seeder
	time.Sleep(3 * time.Second)
	fmt.Println("Uploaded Bytes ", seederTorrent.UploadedBytes())
	fmt.Println("Downloaded Bytes: ", seederTorrent.DownloadedBytes())

	// seeder.Close()

	// start bp
	baselineProviderTorrent, err := baselineProvider.AddTorrent(&trackerlessMetaInfo)
	baselineProviderTorrent.SmallIntervalAllowed = true
	utils.TestSeederInitial(t, baselineProviderTorrent, err)

	time.Sleep(3 * time.Second)
	baselineProviderTorrent.AddTrackers([][]string{{utils.TestTrackerAnnounceUrl}})

	// Wait until transfer is complete
	leecher.WaitAll()
	fmt.Println("Uploaded Bytes ", baselineProviderTorrent.UploadedBytes())
	fmt.Println("Downloaded Bytes: ", baselineProviderTorrent.DownloadedBytes())

	// Verify baseline provider (baseline provider should not get itself as baseline provider)
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{leecherTorrent}, []int{baselineProviderPort})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{baselineProviderTorrent}, []int{})

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

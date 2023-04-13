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

func TestMultipleSeedersOneLeecher(t *testing.T) {
	seederConfig1 := SeederConfig(0, 0)
	utils.CreateDir(t, seederConfig1.DataDir)
	seeder1, _ := rbt.NewClient(seederConfig1)
	defer seeder1.Close()
	defer os.RemoveAll(seederConfig1.DataDir)

	seederConfig2 := SeederConfig(1, 0)
	utils.CreateDir(t, seederConfig2.DataDir)
	seeder2, _ := rbt.NewClient(seederConfig2)
	defer seeder2.Close()
	defer os.RemoveAll(seederConfig2.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig1.DataDir, seederConfig2.DataDir}, utils.TestFileName, 1e6, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent1, err := seeder1.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, seederTorrent1, err)

	seederTorrent2, err := seeder2.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, seederTorrent2, err)

	leecherConfig1 := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig1.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig1)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig1.DataDir)

	leecherTorrent1, _ := leecher.AddTorrent(&metaInfo)
	<-leecherTorrent1.GotInfo()

	leecherTorrent1.DownloadAll()
	leecher.WaitAll()

	utils.VerifyFileContent(t, utils.TestFileName, seederConfig1.DataDir, []string{leecherConfig1.DataDir})
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig2.DataDir, []string{leecherConfig1.DataDir})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent1, seederTorrent2, leecherTorrent1}, []int{})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{}, []int{4000})
}


func TestOneSeederMultipleLeechers(t *testing.T) {
	// Create a seeder 1
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
	leecherConfig1 := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig1.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig1)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig1.DataDir)


	leecherConfig2 := LeecherConfig(1, 0)
	utils.CreateDir(t, leecherConfig2.DataDir)
	leecher2, _ := rbt.NewClient(leecherConfig2)
	defer leecher2.Close()
	defer os.RemoveAll(leecherConfig2.DataDir)

	leecherConfig3 := LeecherConfig(2, 0)
	utils.CreateDir(t, leecherConfig3.DataDir)
	leecher3, _ := rbt.NewClient(leecherConfig3)
	defer leecher3.Close()
	defer os.RemoveAll(leecherConfig3.DataDir)

	// Also attach the metaInfo to the leecher
	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent2, _  := leecher2.AddTorrent(&metaInfo)
	leecherTorrent3, _  := leecher3.AddTorrent(&metaInfo)
	<-leecherTorrent.GotInfo()
	<-leecherTorrent2.GotInfo()
	<-leecherTorrent3.GotInfo()



	// Wait until transfer is complete
	go func() {
		leecherTorrent.DownloadAll()
		
	}()
	
	go func() {
		leecherTorrent2.DownloadAll()
		
	}()

	go func() {
		leecherTorrent3.DownloadAll()
	}()
	

	leecher.WaitAll()
	leecher2.WaitAll()
	leecher3.WaitAll()

	utils.VerifyBaselineProvider(t, []*rbt.Torrent{seederTorrent, leecherTorrent, leecherTorrent2, leecherTorrent3}, []int{})
	utils.VerifyBaselineProvider(t, []*rbt.Torrent{}, []int{4000})
	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig1.DataDir, leecherConfig2.DataDir, leecherConfig3.DataDir})
}
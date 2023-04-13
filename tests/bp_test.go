package tests

import (
	"os"
	"rbtValidation/utils"
	"testing"
	"time"

	rbt "github.com/anacrolix/torrent"
)


func TestBPKnowsLeecherButNoConnection(t *testing.T) {
	baselineProviderConfig := BaselineProviderConfig(0, 4000)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	metaInfo := utils.CreateFileAndMetaInfo(t, []string{baselineProviderConfig.DataDir}, utils.TestFileName, 1e3, [][]string{})
	baselineTorrent, err := baselineProvider.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, baselineTorrent, err)

	leecherConfig := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	baselineTorrent.AddClientPeer(leecher)
	<-leecherTorrent.GotInfo()

	// you want it to timeout 
	leecherTorrent.DownloadAll()
	timeout := time.After(10 * time.Second)
    done := make(chan bool)
    go func() {
		leecher.WaitAll()
        done <- true
    }()

    select {
    case <-timeout:
		t.Logf("Expected")
    case <-done:
		t.Fatal("Download shouldn't have finished")
    }
}


func TestBPNotKnowingLeecher(t *testing.T) {
	baselineProviderConfig := BaselineProviderConfig(0, 4000)
	utils.CreateDir(t, baselineProviderConfig.DataDir)
	baselineProvider, _ := rbt.NewClient(baselineProviderConfig)
	defer baselineProvider.Close()
	defer os.RemoveAll(baselineProviderConfig.DataDir)

	metaInfo := utils.CreateFileAndMetaInfo(t, []string{baselineProviderConfig.DataDir}, utils.TestFileName, 1e3, [][]string{})
	baselineTorrent, err := baselineProvider.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, baselineTorrent, err)

	leecherConfig := LeecherConfig(0, 0)
	utils.CreateDir(t, leecherConfig.DataDir)
	leecher, _ := rbt.NewClient(leecherConfig)
	defer leecher.Close()
	defer os.RemoveAll(leecherConfig.DataDir)

	leecherTorrent, _ := leecher.AddTorrent(&metaInfo)
	leecherTorrent.AddClientPeer(baselineProvider)
	<-leecherTorrent.GotInfo()

	// you want it to timeout 
	leecherTorrent.DownloadAll()
	timeout := time.After(10 * time.Second)
    done := make(chan bool)
    go func() {
		leecher.WaitAll()
        done <- true
    }()

    select {
    case <-timeout:
        t.Fatal("Test should have finished in time")
    case <-done:
		t.Logf("Expected")
    }

	utils.VerifyFileContent(t, utils.TestFileName, baselineProviderConfig.DataDir, []string{leecherConfig.DataDir})
}


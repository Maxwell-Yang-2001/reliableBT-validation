package tests

import (
	"fmt"
	"io"
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
	config.IsFTP = false
	config.DefaultFTPport = 2121
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
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 1e3, [][]string{})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, *seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig()
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

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

// Test the row connection:
// FTP server can transfer file to the leecher successfully by directly build up a client-server connection.
// Note: To run this test make sure your reliableBT-FTP server is running on port 2121.
func TestSeederLeecherFTP(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig()
	seederConfig.IsFTP = true

	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a leecher
	leecher, _ := rbt.NewClient(LeecherConfig())
	defer leecher.Close()

	c, err := leecher.DialFTP()
	if err != nil {
		t.Log("Dial to FTP failed")
	}

	fileName := "/kitty.jpg"
	// read the hello.txt from
	ftpResp, err := c.Retr(fileName)
	if err != nil {
		t.Log("File not Found")
	}

	defer ftpResp.Close()

	newFile, err := os.OpenFile(LeecherConfig().DataDir+fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Log(err)
	}

	defer newFile.Close()

	// Read from remote file in chunks of 1024 bytes and write to leetcher local db
	buffer := make([]byte, 1024)
	for {
		n, err := ftpResp.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		_, err = newFile.Write(buffer[:n])
		if err != nil {
			panic(err)
		}

		fmt.Printf("Wrote %d bytes to file\n", n)
	}

	// check if the file actually exists
	_, err = os.Stat(LeecherConfig().DataDir + fileName)
	if err != nil {
		t.Log("File doesn't exist in the Leecher data dir")
		t.FailNow()
	}
}

// Test whether a seeder can transfer file to a leecher successfully by tracker letting them discover each other.
func TestSeederLeecherTracker(t *testing.T) {
	// Create a seeder
	seederConfig := SeederConfig()
	utils.CreateDir(t, seederConfig.DataDir)
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()
	defer os.RemoveAll(seederConfig.DataDir)

	// Create a test file within the seeder dir and add it to the seeder client
	metaInfo := utils.CreateFileAndMetaInfo(t, []string{seederConfig.DataDir}, utils.TestFileName, 1e6, [][]string{{utils.TestTrackerAnnounceUrl}})
	seederTorrent, err := seeder.AddTorrent(&metaInfo)
	utils.TestSeederInitial(t, *seederTorrent, err)

	// Create a leecher
	leecherConfig := LeecherConfig()
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

	// Verify file content equality
	utils.VerifyFileContent(t, utils.TestFileName, seederConfig.DataDir, []string{leecherConfig.DataDir})
}

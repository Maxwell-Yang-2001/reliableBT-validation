package tests

import (
	"testing"

	rbt "github.com/anacrolix/torrent"
)

func TestDownloadSpeed(t *testing.T) {
	// Create a seeder
	seederConfig := seederConfig()
	seeder, _ := rbt.NewClient(seederConfig)
	defer seeder.Close()

	// Create a magnet link and add it to the seeder (tracker on localhost is attached in the magnet)
	magnetLink := MakeMagnet(t, seeder, seederConfig.DataDir, "ubuntu-14.04.2-desktop-amd64.iso", [][]string{{"http://127.0.0.1:1337/announce"}})

	// Create a leecher
	leecher, _ := rbt.NewClient(leecherConfig())
	defer leecher.Close()

	// Also attach the magnet link to the leecher
	leecher_torrent, _ := leecher.AddMagnet(magnetLink)
	<-leecher_torrent.GotInfo()

	// Wait until transfer is complete
	leecher_torrent.DownloadAll()
    
	// downloadSpeed := leecher_torrent.DoHttpSend(leecher_torrent.Stats().BytesReadData)

	
	// fmt.Println(downloadSpeed)
	// time.Sleep(5 * time.Second)
	// speed := leecher_torrent.DoHttpSend(leecher_torrent.Stats().BytesReadData)
	// fmt.Println(speed)
	leecher.WaitAll()
}

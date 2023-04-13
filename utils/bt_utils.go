package utils

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"

	rbt "github.com/anacrolix/torrent"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

// Create a file within each directory with specified size, return the metaInfo to be added to any client.
func CreateFileAndMetaInfo(t *testing.T, dirs []string, name string, size int64, trackers [][]string) (metaInfo metainfo.MetaInfo) {
	// Create the files first - one for each dir
	require.NotZero(t, len(dirs))
	filePaths := make([]string, len(dirs))
	for i, dir := range dirs {
		filePaths[i] = filepath.Join(dir, name)
	}
	CreateFiles(t, filePaths, size)

	metaInfo = metainfo.MetaInfo{AnnounceList: trackers}
	metaInfo.SetDefaults()
	info := metainfo.Info{PieceLength: PieceLength}
	err := info.BuildFromFilePath(filePaths[0])
	require.NoError(t, err)
	metaInfo.InfoBytes, err = bencode.Marshal(info)
	require.NoError(t, err)
	return
}

// Create a file within each directory with specified size, return the metaInfo to be added to any client.
func CreateFilesInDirs(t *testing.T, dirs []string, name string, size int64) {
	// Create the files first - one for each dir
	require.NotZero(t, len(dirs))
	filePaths := make([]string, len(dirs))
	for i, dir := range dirs {
		filePaths[i] = filepath.Join(dir, name)
	}
	CreateFiles(t, filePaths, size)
}

// Create a file within each directory with specified size, return the metaInfo to be added to any client.
func CreateMetaInfo(t *testing.T, dir string, name string, trackers [][]string) (metaInfo metainfo.MetaInfo) {
	filePath := filepath.Join(dir, name)

	metaInfo = metainfo.MetaInfo{AnnounceList: trackers}
	metaInfo.SetDefaults()
	info := metainfo.Info{PieceLength: PieceLength}
	err := info.BuildFromFilePath(filePath)
	require.NoError(t, err)
	metaInfo.InfoBytes, err = bencode.Marshal(info)
	require.NoError(t, err)
	return
}

// Test the initial setup of a seeder torrent.
func TestSeederInitial(t *testing.T, tr *rbt.Torrent, err error) {
	require.NoError(t, err)
	require.True(t, tr.Seeding())
	tr.VerifyData()
}

// Test whether a baseline provider for given torrent instances matches with one of the expected values
func VerifyBaselineProvider(t *testing.T, trs []*rbt.Torrent, ports []int) {
	fmt.Println("Verifying baseline provider against expectation for each torrent instance")
	for _, tr := range trs {
		bpIP, bpPort := tr.GetBaselineProvider()
		// If no ports are given, expect no baseline provider - IP should be nil and port should be 0
		if ports == nil || len(ports) == 0 {
			require.Nil(t, bpIP)
			require.Zero(t, bpPort)
			continue
		}

		// Otherwise, the IP should be the localhost, while the port should be one from the expectation
		localhostIP := net.ParseIP(Localhost)
		require.True(t, bpIP.Equal(localhostIP))
		foundPort := false
		for _, port := range ports {
			if bpPort == port {
				foundPort = true
				break
			}
		}
		if !foundPort {
			t.FailNow()
		}
	}
	fmt.Println("SUCCESS: Baseline provider matches expectation from each torrent instance")
}

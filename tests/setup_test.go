package tests

import (
	"rbtValidation/utils"
	"testing"

	rbt "github.com/anacrolix/torrent"
)

// Verifies whether local copy of the reliableBT client repo is in use.
func TestClientSetup(t *testing.T) {
	ret := rbt.SetupCheck()
	const exp = "Setup Successfully"
	if ret == exp {
		t.Logf(ret)
	} else {
		t.Errorf("got %s, wanted %s", ret, exp)
	}
}

// Sets up the necessary files under tests/seeder for testing purpose.
// The total size is about 31 GB - please make sure your device has enough storage for them.
// WARNING: this might take a bit of time (up to 1-2 mins).
func TestFileSetup(t *testing.T) {
	// Power of 10 for K, M and G - so actual size is slightly smaller than specified
	utils.CreateFile(t, "../resources/test_1kb.txt", 1e3)
	utils.CreateFile(t, "../resources/test_1mb.txt", 1e6)
	utils.CreateFile(t, "../resources/test_1gb.txt", 1e9)
	utils.CreateFile(t, "../resources/test_10gb.txt", 1e10)
	utils.CreateFile(t, "../resources/test_20gb.txt", 2e10)
}

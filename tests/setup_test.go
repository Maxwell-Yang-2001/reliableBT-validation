package tests

import (
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

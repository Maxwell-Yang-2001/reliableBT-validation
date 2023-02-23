package tests

import (
	"testing"

	rbt "github.com/anacrolix/torrent"
)

func TestSetup(t *testing.T) {
	ret := rbt.SetupCheck()
	const exp = "Setup Successfully"
	if ret == exp {
		t.Logf(ret)
	} else {
		t.Errorf("got %s, wanted %s", ret, exp)
	}
}

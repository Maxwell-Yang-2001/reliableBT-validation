package references

import (
	"log"

	rbt "github.com/anacrolix/torrent"
)

// Running this function will download a ubuntu .iso file (~ 1GB) in the directory
func main() {
	c, _ := rbt.NewClient(nil)
	defer c.Close()
	t, _ := c.AddMagnet("magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU")
	<-t.GotInfo()
	t.DownloadAll()
	c.WaitAll()
	log.Print("ermahgerd, torrent downloaded")
}

package utils

import "golang.org/x/time/rate"

const (
	PieceLength            = 256 * 1024 // 256K
	TestFileName           = "test.txt"
	TestTrackerAnnounceUrl = "http://127.0.0.1:1337/announce"
	Localhost              = "127.0.0.1"
	DefaultChunkSize       = 16 * 1024 // 16k
)

var SmallRateLimiter = rate.NewLimiter(1, DefaultChunkSize)

package config

import (
	"time"

	"github.com/anyweez/matchgrab/structs"
)

type config struct {
	HTTPTimeout        time.Duration // timeout on requests to Riot API
	MatchStoreLocation string
	SeedAccount        structs.RiotID

	MaxSimultaneousRequests int
	RequestsPerMinute       int
	MaxTimeAgo              time.Duration
}

var Config config

func Setup() {
	Config = config{
		HTTPTimeout:             20 * time.Second,
		MatchStoreLocation:      "/Volumes/LOLMatches/matches.db",
		SeedAccount:             50669460,
		MaxSimultaneousRequests: 6,
		RequestsPerMinute:       480,
		MaxTimeAgo:              time.Duration(60 * 24 * time.Hour), // 60 days
	}
}

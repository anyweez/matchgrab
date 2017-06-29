package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/anyweez/matchgrab/structs"
)

type config struct {
	HTTPTimeout        time.Duration  `json:"http_timeout"` // timeout on requests to Riot API
	MatchStoreLocation string         `json:"match_store_location"`
	SeedAccount        structs.RiotID `json:"seed_account"`

	MaxSimultaneousRequests int           `json:"max_sim_requests"`
	RequestsPerMinute       int           `json:"requests_per_min"`
	MaxTimeAgo              time.Duration `json:"max_time_ago"`
}

var Config config

func Setup() {
	// Check if there's a config file specified. If so, we should use those settings.
	raw, err := ioutil.ReadFile("config.json")

	// Fallbacks in case users don't specify these.
	defaults := config{
		HTTPTimeout:             20 * time.Second,
		MatchStoreLocation:      "/Volumes/LOLMatches/matches.db",
		SeedAccount:             50669460,
		MaxSimultaneousRequests: 6,
		RequestsPerMinute:       480,
		MaxTimeAgo:              time.Duration(60 * 24 * time.Hour), // 60 days
	}

	// TODO: probably a cleaner way to do this; need to find golang pattern
	if err == nil {
		specified := struct {
			HTTPTimeout        string         `json:"http_timeout"` // timeout on requests to Riot API
			MatchStoreLocation string         `json:"match_store_location"`
			SeedAccount        structs.RiotID `json:"seed_account"`

			MaxSimultaneousRequests int    `json:"max_sim_requests"`
			RequestsPerMinute       int    `json:"requests_per_min"`
			MaxTimeAgo              string `json:"max_time_ago"`
		}{}

		json.Unmarshal(raw, &specified)

		fmt.Println(specified)

		// Replace timeout if its present (parse first!)
		if specified.HTTPTimeout != "" {
			timeout, err := time.ParseDuration(specified.HTTPTimeout)

			if err != nil {
				panic(err)
			}

			defaults.HTTPTimeout = timeout
		}

		if specified.MatchStoreLocation != "" {
			defaults.MatchStoreLocation = specified.MatchStoreLocation
		}

		if specified.MaxSimultaneousRequests != 0 {
			defaults.MaxSimultaneousRequests = specified.MaxSimultaneousRequests
		}

		if specified.RequestsPerMinute != 0 {
			defaults.RequestsPerMinute = specified.RequestsPerMinute
		}

		// Replace MaxTimeAgo if its present (parse first!)
		if specified.MaxTimeAgo != "" {
			timeago, err := time.ParseDuration(specified.MaxTimeAgo)

			if err != nil {
				panic(err)
			}

			defaults.MaxTimeAgo = timeago
		}
	}

	Config = defaults
}

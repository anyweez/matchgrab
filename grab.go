package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/anyweez/matchgrab/api"
	"github.com/anyweez/matchgrab/config"
	"github.com/anyweez/matchgrab/display"
	"github.com/anyweez/matchgrab/structs"
)

var done chan bool
var requestRoutines chan bool // respawn routines if they die
var matches *structs.IDList
var summoners *structs.IDList
var store *structs.MatchStore
var ui *display.Display
var rateLimit chan bool

var knownSummoners map[structs.RiotID]bool
var ksLock sync.Mutex

func main() {
	// Initialize application configuration
	config.Setup()

	knownSummoners = make(map[structs.RiotID]bool, 0)
	done = make(chan bool)
	requestRoutines = make(chan bool, config.Config.MaxSimultaneousRequests)
	rateLimit = make(chan bool, 100)
	matches = structs.NewIDList()
	summoners = structs.NewIDList()
	store = structs.NewMatchStore(config.Config.MatchStoreLocation)
	ui = display.NewDisplay(Shutdown)

	summoners.Add(config.Config.SeedAccount)

	// Rate limit channel
	go func() {
		for {
			rateLimit <- true
			time.Sleep((1 * time.Minute) / time.Duration(config.Config.RequestsPerMinute))
		}
	}()

	// Load all existing matches and summoners in parallel
	store.Each(func(m *structs.Match) {
		ui.AddEvent(fmt.Sprintf("[ Match  ] Loading %d...", m.GameID))
		matches.Blacklist(m.GameID) // don't need to re-run matches

		ksLock.Lock()
		for _, p := range m.Participants {
			// Some duplicates (fine), many new folks as well
			summoners.Add(p.AccountID)
			knownSummoners[p.AccountID] = true
		}
		ksLock.Unlock()

		ui.UpdateQueuedSummoners(summoners.Filled() * 100)
		ui.UpdateTotalSummoners(len(knownSummoners))
	})
	// Shuffle so we don't start with the same group every time.
	summoners.Shuffle()

	ui.AddEvent("Loaded existing match database!")

	// Start requesting and never stop.
	requestLoop()

	<-done
}

func requestLoop() {
	lastRps := time.Now()
	rpsInterval := 5 * time.Second // recompute @ this interval
	rpsWindow := 30                // compute rps using records within this window (seconds)
	// Start running the requests
	pace := structs.NewPacer(
		config.Config.RequestsPerMinute,
		config.Config.MaxSimultaneousRequests,
	)
	// Requests must ALWAYS be queued earliest first. This order is assumed for rps counting.
	requestLog := make(chan time.Time, 100000)

	// RPS calculation
	go func() {
		// Wait until we have a full window before displaying stats.
		time.Sleep(time.Duration(rpsWindow) * time.Second)

		for {
			if time.Since(lastRps) > rpsInterval {
				// Pull all out-of-range records off the queue. Note that we'll pull one
				// more than we actually intend to, so we'll need to +1 below.
				for t := range requestLog {
					if time.Since(t) < time.Duration(rpsWindow)*time.Second {
						break
					}
				}

				rps := float32(len(requestLog)+1) / float32(rpsWindow)
				ui.UpdateRequestsPerSecond(rps)

				lastRps = time.Now()
			}

			time.Sleep(1 * time.Second)
		}
	}()

	// Seed the RNG
	rand.Seed(time.Now().Unix())
	pace.Run(func() {
		if rand.Float32() < matches.Filled() { // Request match
			getMatch()
			requestLog <- time.Now()
		} else if summoners.Available() { // Request summoner games
			getSummoner()
			requestLog <- time.Now()
		}
	}, 0)
}

func getMatch() {
	match, available := matches.Next()
	if !available {
		ui.AddEvent(fmt.Sprintf("[ Match  ] Queue empty, skipping..."))
		return
	}

	ui.AddEvent(fmt.Sprintf("[ Match  ] Fetching %d...", match))

	url := fmt.Sprintf(
		"https://na1.api.riotgames.com/lol/match/v3/matches/%d",
		match,
	)

	api.Get(url, func(body []byte) {
		var full structs.APIMatch
		json.Unmarshal(body, &full)

		// Store the match
		match := structs.ToMatch(full)
		store.Add(match)

		// Add all account ID's to the summoner queue.
		ksLock.Lock()
		for i := 0; i < len(match.Participants); i++ {
			summoners.Add(match.Participants[i].AccountID)
			knownSummoners[match.Participants[i].AccountID] = true
		}
		ksLock.Unlock()

		ui.UpdateQueuedSummoners(summoners.Filled() * 100)
		ui.UpdateTotalSummoners(len(knownSummoners))
	})
}

// Fetch a new set of match ID's for a new summoner. All returned match ID's are queued
// up for future requests.
func getSummoner() {
	summoner, available := summoners.Next()
	if !available {
		ui.AddEvent(fmt.Sprintf("[Summoner] Queue empty, skipping..."))
		return
	}

	ui.AddEvent(fmt.Sprintf("[Summoner] Fetching %d...", summoner))

	url := fmt.Sprintf(
		"https://na1.api.riotgames.com/lol/match/v3/matchlists/by-account/%d",
		summoner,
	)

	err := api.Get(url, func(body []byte) {
		summaries := struct {
			Matches []structs.MatchSummary `json:"matches"`
		}{
			Matches: make([]structs.MatchSummary, 0),
		}

		json.Unmarshal(body, &summaries)

		// Add all summoners to the
		for _, match := range summaries.Matches {
			matchTime := time.Unix(match.Timestamp/1000, 0)

			// Only look for matches that occurred recently.
			if time.Since(matchTime) < config.Config.MaxTimeAgo {
				matches.Add(match.GameID)
			}
		}

		ui.UpdateQueuedMatches(matches.Filled() * 100)
		ui.UpdateTotalMatches(store.Count())
	})

	if err != nil {
		ui.AddEvent(err.Error())
	}
}

// Shutdown : Called by termui when the user indicates they want to quit
func Shutdown() {
	store.Close()

	done <- true

	os.Exit(0)
}

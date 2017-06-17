package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/anyweez/kickoff/utils"
	"github.com/anyweez/matchgrab/api"
	"github.com/anyweez/matchgrab/display"
	"github.com/anyweez/matchgrab/structs"
)

var done chan bool
var matches *structs.IDList
var summoners *structs.IDList
var store structs.MatchStore
var ui *display.Display
var rateLimit chan bool

const (
	MaxSimultaneousRequests = 1
	RequestsPerMinute       = 120
	RequestTimeout          = 20
)

func main() {
	/* If we don't have an API key we can't do anything */
	if len(os.Getenv("RIOT_API_KEY")) == 0 {
		utils.Log("No RIOT_API_KEY specified; cannot continue.\n")
		return
	}

	done = make(chan bool)
	rateLimit = make(chan bool, 10)
	matches = structs.NewIDList()
	summoners = structs.NewIDList()
	store = structs.NewMatchStore("demo.db")
	ui = display.NewDisplay()

	summoners.Add(50669460)

	// Rate limit channel
	go func() {
		for {
			rateLimit <- true
			time.Sleep(1 * time.Minute / RequestsPerMinute)
		}
	}()

	// Load all existing matches and summoners in parallel
	go func() {
		store.Each(func(m structs.Match) {
			ui.AddEvent(fmt.Sprintf("[ Match  ] Loading %d...", m.GameID))
			matches.Blacklist(m.GameID) // don't need to re-run matches

			for i := 0; i < len(m.Participants); i++ {
				// Some duplicates, many new folks as well
				summoners.Add(m.Participants[i].AccountID)
			}

			ui.UpdateQueuedSummoners(summoners.Filled())
			ui.UpdateTotalSummoners(summoners.Known())
		})
	}()

	/* Load all existing matches */
	request()

	<-done
}

func request() {
	// Requests must ALWAYS be queued earliest first. This order is assumed for rps counting.
	requestLog := make(chan time.Time, 100000)
	lastRps := time.Now()
	rpsInterval := 5 * time.Second // recompute @ this interval
	rpsWindow := 30                // compute rps using records within this window (seconds)

	for i := 0; i < MaxSimultaneousRequests; i++ {
		/* Launch a goroutine to make requests */
		go func() {
			for {
				<-rateLimit                                      // Make sure we aren't rate limited
				if rand.Float32() > 0.1 && matches.Available() { // Request match
					getMatch()
					requestLog <- time.Now()
				} else if summoners.Available() { // Request summoner games
					getSummoner()
					requestLog <- time.Now()
				}
			}
		}()
	}

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
				// fmt.Println(fmt.Sprintf("updated rps: %.2f", rps))
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func getMatch() {
	match := matches.Next()
	ui.AddEvent(fmt.Sprintf("[ Match  ] Fetching %d...", match))

	url := fmt.Sprintf(
		"https://na1.api.riotgames.com/lol/match/v3/matches/%d?api_key=%s",
		match,
		os.Getenv("RIOT_API_KEY"),
	)

	api.Get(url, func(body []byte) {
		var full structs.APIMatch
		json.Unmarshal(body, &full)

		// Store the match
		match := structs.ToMatch(full)
		store.Add(match)

		// Add all account ID's to the summoner queue.
		for i := 0; i < len(match.Participants); i++ {
			summoners.Add(match.Participants[i].AccountID)
		}

		ui.UpdateQueuedSummoners(summoners.Filled())
		ui.UpdateTotalSummoners(summoners.Known())
	})
}

// Fetch a new set of match ID's for a new summoner. All returned match ID's are queued
// up for future requests.
func getSummoner() {
	summoner := summoners.Next()
	ui.AddEvent(fmt.Sprintf("[Summoner] Fetching %d...", summoner))

	url := fmt.Sprintf(
		"https://na1.api.riotgames.com/lol/match/v3/matchlists/by-account/%d",
		summoner,
	)

	api.Get(url, func(body []byte) {
		summaries := struct {
			Matches []structs.MatchSummary `json:"matches"`
		}{
			Matches: make([]structs.MatchSummary, 0),
		}

		json.Unmarshal(body, &summaries)

		// Add all summoners to the
		for i := 0; i < len(summaries.Matches); i++ {
			matches.Add(summaries.Matches[i].GameID)
		}

		ui.UpdateQueuedMatches(matches.Filled())
		ui.UpdateTotalMatches(matches.Known())
	})
}

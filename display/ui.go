package display

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

type Display struct {
	title *termui.Par
	// Total number of RECORDED matches. From MatchStore.Count()
	statTotalMatches *termui.Par
	// Total size of the 'summoner universe' - the number of summoner ID's we've seen
	statTotalSumms *termui.Par
	// Requests per second (all request types)
	statRps *termui.Par
	// How long the application has been running
	statAliveTimer *termui.Par
	// How full the summoner queue is
	summoners *termui.Gauge
	// How full the match queue is
	matches *termui.Gauge
	// Recent events
	events *termui.List

	_created time.Time
}

const (
	eventCount = 30
)

func NewDisplay(cleanup func()) *Display {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	d := &Display{
		title:            termui.NewPar("Matchgrab (q to exit)"),
		statTotalMatches: termui.NewPar("---"),
		statTotalSumms:   termui.NewPar("---"),
		statRps:          termui.NewPar("---"),
		statAliveTimer:   termui.NewPar("---"),
		summoners:        termui.NewGauge(),
		matches:          termui.NewGauge(),
		events:           termui.NewList(),

		_created: time.Now(),
	}

	d.title.Border = false

	d.statAliveTimer.BorderLabel = "Alive for"
	d.statAliveTimer.Height = 3
	d.statRps.BorderLabel = "Requests per second"
	d.statRps.Height = 3
	d.statTotalMatches.BorderLabel = "Stored matches"
	d.statTotalMatches.Height = 3
	d.statTotalSumms.BorderLabel = "Summoner universe"
	d.statTotalSumms.Height = 3

	d.summoners.BorderLabel = "Summoner queue"
	// d.summoners.Percent = 20

	d.matches.BorderLabel = "Match queue"
	// d.matches.Percent = 20

	d.events.BorderLabel = "Events"
	d.events.Items = []string{}
	d.events.ItemFgColor = termui.ColorWhite
	d.events.Height = eventCount

	// Setup the layout
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, d.title),
		),
		termui.NewRow(
			termui.NewCol(3, 0, d.statAliveTimer),
			termui.NewCol(3, 0, d.statRps),
			termui.NewCol(3, 0, d.statTotalMatches),
			termui.NewCol(3, 0, d.statTotalSumms),
		),
		termui.NewRow(
			termui.NewCol(6, 0, d.summoners),
			termui.NewCol(6, 0, d.matches),
		),
		termui.NewRow(
			termui.NewCol(12, 0, d.events),
		),
	)

	// press q to quit
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
		termui.Close()

		cleanup() // callback for main process cleanup
	})

	termui.Handle("/timer/1s", func(termui.Event) {
		d.statAliveTimer.Text = " " + time.Since(d._created).String()

		termui.Body.Align() // TODO: necessary?
		termui.Render(termui.Body)
	})

	termui.Body.Align()
	termui.Render(termui.Body)

	go func() {
		termui.Loop()
	}()

	return d
}

func (d *Display) AddEvent(ev string) {
	d.events.Items = append(d.events.Items, ev)

	if len(d.events.Items) > eventCount {
		d.events.Items = d.events.Items[len(d.events.Items)-eventCount:]
	}
}

func (d *Display) UpdateQueuedSummoners(pct float32) {
	d.summoners.Percent = int(pct)
}

func (d *Display) UpdateQueuedMatches(pct float32) {
	d.matches.Percent = int(pct)
}

func (d *Display) UpdateTotalSummoners(n int) {
	d.statTotalSumms.Text = fmt.Sprintf(" %d", n)
}

func (d *Display) UpdateTotalMatches(n int) {
	d.statTotalMatches.Text = fmt.Sprintf(" %d", n)
}

func (d *Display) UpdateRequestsPerSecond(n float32) {
	d.statRps.Text = fmt.Sprintf(" %.1f", n)
}

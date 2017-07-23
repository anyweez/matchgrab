package structs

import (
	"testing"
)

type champStatus struct {
	Banned bool
	Picked bool
	Won    bool
}

func outcomes() map[RiotID]champStatus {
	return map[RiotID]champStatus{
		122: champStatus{Banned: true, Picked: false, Won: false},
		164: champStatus{Banned: true, Picked: false, Won: false},
		23:  champStatus{Banned: true, Picked: false, Won: false},
		50:  champStatus{Banned: false, Picked: false, Won: false},
		3:   champStatus{Banned: false, Picked: true, Won: true},
		19:  champStatus{Banned: true, Picked: false, Won: false},
	}
}

// Make sure we are encoding + decoding to storage format correctly.
func TestToFromProto(t *testing.T) {
	samples := rawSamples()

	for _, sample := range samples {
		match := ToMatch(sample)

		buf := match.Bytes()
		m2 := MakeMatch(buf)

		if match.GameID != m2.GameID {
			t.Fail()
		}

		if match.SeasonID != m2.SeasonID {
			t.Fail()
		}

		if match.GameCreation != m2.GameCreation {
			t.Fail()
		}

		if match.GameDuration != m2.GameDuration {
			t.Fail()
		}

		// TODO: Check participants and ban lists as well.
	}
}

// TODO: save some raw json match data for next tests

// Ensure we correctly parse raw API responses into Match's.
func TestCanParse(t *testing.T) {
	matches := rawSamples()

	for _, sample := range matches {
		match := ToMatch(sample)

		if match.GameID != sample.GameID {
			t.Fail()
		}

		if match.SeasonID != sample.SeasonID {
			t.Fail()
		}

		if match.GameCreation != sample.GameCreation {
			t.Fail()
		}

		if match.GameDuration != sample.GameDuration {
			t.Fail()
		}
	}
}

// Test Banned(), Picked(), and Won() with unpacked matches.
func TestCountsUnpacked(t *testing.T) {
	samples := rawSamples()
	matches := make([]Match, 0, len(samples))

	for _, sample := range samples {
		matches = append(matches, ToMatch(sample))
	}

	oc := outcomes()

	for id, outcome := range oc {
		if matches[0].Banned(id) != outcome.Banned {
			t.Fail()
		}

		if matches[0].Picked(id) != outcome.Picked {
			t.Fail()
		}

		if matches[0].Won(id) != outcome.Won {
			t.Fail()
		}
	}
}

// TODO: test Banned(), Picked(), and Won() both packed and unpacked

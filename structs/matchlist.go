package structs

import "sync"

const MaxIDListSize = 100000

type IDList struct {
	Queue chan RiotID
	// Overflow []RiotID
	ids map[RiotID]bool

	// Number of available values
	// count int
	// Number of recorded or queued values
	// known int

	// Concurrency mutexes
	lockIDs      sync.Mutex
	lockOverflow sync.Mutex
}

func NewIDList() *IDList {
	ml := IDList{
		Queue: make(chan RiotID, MaxIDListSize),
		// Overflow: make([]RiotID, 0, MaxIDListSize),
		ids: make(map[RiotID]bool),
		// count: 0,
		// known: 0,
	}

	return &ml
}

func (ml *IDList) Add(m RiotID) bool {
	ml.lockIDs.Lock()
	_, exists := ml.ids[m]
	ml.lockIDs.Unlock()

	if !exists {
		// Try to add it to the queue; if it doesn't work then skip
		select {
		case ml.Queue <- m:
			ml.Blacklist(m)

			return true
		}
	}

	return false
}

func (ml *IDList) Blacklist(m RiotID) {
	ml.lockIDs.Lock()
	ml.ids[m] = true
	// ml.known++
	ml.lockIDs.Unlock()
}

func (ml *IDList) Next() (RiotID, bool) {
	if len(ml.Queue) > 0 {
		return <-ml.Queue, true
	}

	return -1, false
}

// Available : Returns true if there are any items in the list.
func (ml *IDList) Available() bool {
	return len(ml.Queue) > 0
}

// Filled : Returns the percentage of the list capacity that's filled
func (ml *IDList) Filled() float32 {
	return (float32(len(ml.Queue)) / float32(MaxIDListSize)) * 100
}

// Known : Return the number of ID's that have ever been blacklisted on this list.
// func (ml *IDList) Known() int {
// 	return ml.known
// }

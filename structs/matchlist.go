package structs

import "sync"

const MaxIDListSize = 100000

type IDList struct {
	Queue    chan RiotID
	Overflow []RiotID
	ids      map[RiotID]bool

	// Number of available values
	count int
	// Number of recorded or queued values
	known int

	// Concurrency mutexes
	lockIDs      sync.Mutex
	lockOverflow sync.Mutex
}

func NewIDList() *IDList {
	ml := IDList{
		Queue:    make(chan RiotID, 1000),
		Overflow: make([]RiotID, 0, MaxIDListSize),
		ids:      make(map[RiotID]bool),
		count:    0,
		known:    0,
	}

	return &ml
}

func (ml *IDList) Add(m RiotID) bool {
	ml.lockIDs.Lock()
	if _, exists := ml.ids[m]; !exists && ml.count < MaxIDListSize {
		ml.lockIDs.Unlock()

		select {
		case ml.Queue <- m:
		default:
			ml.lockOverflow.Lock()
			ml.Overflow = append(ml.Overflow, m)
			ml.lockOverflow.Unlock()
		}

		ml.Blacklist(m)
		ml.count++

		return true
	}
	ml.lockIDs.Unlock()

	return false
}

func (ml *IDList) Blacklist(m RiotID) {
	ml.lockIDs.Lock()
	ml.ids[m] = true
	ml.known++
	ml.lockIDs.Unlock()
}

func (ml *IDList) Next() RiotID {
	if ml.count > 0 {
		ml.count--

		n := <-ml.Queue

		if len(ml.Overflow) > 0 {
			var refill RiotID

			ml.lockOverflow.Lock()
			refill, ml.Overflow = ml.Overflow[0], ml.Overflow[1:]
			ml.Queue <- refill
			ml.lockOverflow.Unlock()
		}

		return n
	}

	return -1
}

func (ml *IDList) Available() bool {
	return ml.count > 0
}

// Filled : Returns the percentage of the list capacity that's filled
func (ml *IDList) Filled() float32 {
	return (float32(ml.count) / float32(MaxIDListSize+1000)) * 100
}

// Known : Return the number of ID's that have ever been blacklisted on this list.
func (ml *IDList) Known() int {
	return ml.known
}

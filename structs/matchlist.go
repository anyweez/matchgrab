package structs

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/willf/bloom"
)

const (
	MaxIDListSize = 100000
	BloomGrowBy   = 2     // double in size on each grow()
	BloomFpRate   = 0.001 // Increasing means more duplicate requests, decreasing means bloom filter consumes more memory.
)

// IDList : A queue-like data structure that only allows items to be added once; if an item
// has already been added then attempts to re-add it will be rejected. IDLists are safe to
// use concurrently.
type IDList struct {
	Queue chan RiotID

	blacklist   *bloom.BloomFilter
	blcap       uint       // Maximum amount of records blacklist can hold
	blacklisted uint       // Total number of values that have been blacklisted
	lockIDs     sync.Mutex // Concurrency mutexes

	_growCount int // [debugging] count number of grow() calls
}

// NewIDList : Create an empty IDList.
func NewIDList() *IDList {
	blsize := uint(MaxIDListSize * 10)

	ml := IDList{
		Queue:       make(chan RiotID, MaxIDListSize),
		blacklisted: 0,
		blacklist:   bloom.NewWithEstimates(blsize, BloomFpRate),
		blcap:       blsize,
	}

	return &ml
}

// Add : Add a new item to the list if it hasn't been added before. Items are
// added in order and cannot be added to the same list twice.
// Note that `IDList` uses a bloom filter to track which elements have been
// added and some false positives occur, meaning that some items will be
// incorrectly blocked.
func (ml *IDList) Add(m RiotID) bool {
	if !ml.Blacklisted(m) {
		// Try to add it to the queue; if it doesn't work then skip
		select {
		case ml.Queue <- m:
			ml.Blacklist(m)

			return true
		default:
			return false
		}
	}

	return false
}

// Blacklisted : Returns a boolean indicating whether the specified ID exists in
// the list. Note that `IDList` uses a bloom filter to track which elements have
// been added so some false positives will occur.
func (ml *IDList) Blacklisted(m RiotID) bool {
	return ml.blacklist.Test(m.Bytes())
}

// Blacklist : Add a new item to the blacklist. This is automatically called by
// Add() and shouldn't be called externally unless you want to blacklist *without*
// adding to the list.
func (ml *IDList) Blacklist(m RiotID) {
	// Check to make sure it isn't blacklisted (primarily to keep the
	// count accurate).
	if !ml.Blacklisted(m) {
		ml.lockIDs.Lock()

		// Grow before we add if we're at capacity
		if ml.blacklisted >= ml.blcap {
			ml.grow() // MUST hold lockIDs lock
		}

		// Add byte buffer to blacklist.
		ml.blacklist.Add(m.Bytes())

		ml.blacklisted++
		ml.lockIDs.Unlock()
	}
}

// grow : Increase the capacity of the bloom filter.
func (ml *IDList) grow() {
	log.Println(fmt.Sprintf("growing @ %s", time.Now()))
	// Old blacklist; outgrown
	retired := ml.blacklist

	// Replace with a new blacklist struct that's bigger
	ml.blcap = ml.blcap * BloomGrowBy
	ml.blacklist = bloom.NewWithEstimates(ml.blcap, BloomFpRate)

	// Merge in old data so that we don't lose any data.
	// Side note: figure out how this works.
	ml.blacklist.Merge(retired)

	ml._growCount++
}

// Next : Get the next item from the list if anything is available. The second
// return value will be true whenever an actual value is returned and false otherwise.
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
	return (float32(len(ml.Queue)) / float32(MaxIDListSize))
}

// Shuffle : Randomly distributes all items currently in the queue. Note that
// this only applies to items *currently in the queue* and will not affect
// insertion order for new items.
func (ml *IDList) Shuffle() {
	rand.Seed(time.Now().UnixNano())

	ids := make([]RiotID, len(ml.Queue))

	// Copy data out of the channel
	for i := 0; i < len(ids); i++ {
		ids[i] = <-ml.Queue
	}

	// Shuffle data
	for here, _ := range ids {
		there := rand.Intn(len(ids))

		ids[here], ids[there] = ids[there], ids[here]
	}

	// Add it back to the channel
	for i := 0; i < len(ids); i++ {
		ml.Queue <- ids[i]
	}
}

package structs

import (
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	copyInterval = 10 * time.Minute
)

// MatchStore : Represents a persistent data store for match data. Implements a thin layer over
// a LevelDB instance and is capable of reading and writing match data to the database. All
// writes are serialized and its therefore safe to call `Add()` from multiple goroutines.
type MatchStore struct {
	queue     chan Match
	db        *leveldb.DB
	count     int
	countInit bool // becomes true if the count is accurate
	active    bool // stop secondary goroutines when this becomes false

	dbLock sync.Mutex
}

// NewMatchStore : Create a new MatchStore that automatically records data and sync it
// to a snapshot instance.
func NewMatchStore(filename string) *MatchStore {
	ms := makeMs(filename, true)

	return ms
}

// makeMs : Internal method for creating a MatchStore with pre-populated defaults.
func makeMs(filename string, makeSnapshot bool) *MatchStore {
	ms := &MatchStore{
		queue:     make(chan Match, 10),
		count:     0,
		countInit: false,
		active:    true,
	}

	var err error
	ms.db, err = leveldb.OpenFile(filename, nil)

	if err != nil {
		panic("Cannot open LevelDB records: " + err.Error())
	}

	// Goroutine that asynchronously writes match data until the matchstore is closed.
	// Once MatchStore.Close() is called, this goroutine finishes writing all queued
	// changes and then closes the database, releasing the lock.
	go func() {
		for m := range ms.queue {
			err := ms.db.Put(m.GameID.Bytes(), m.Bytes(), nil)

			if err != nil {
				panic("Error writing record: " + err.Error())
			}

			ms.count++
		}

		ms.db.Close()
	}()

	if makeSnapshot {
		// Periodically copy all data over to a second database that can be accessed while
		// new data is being downloaded to the primary.
		go func() {
			// Back up everything once per period (defined as const above).
			for ms.active {
				time.Sleep(copyInterval)

				// Open, take snapshot, and then close. Keep the lock for as little time as possible.
				backup := makeMs(filename+"-snapshot", false)
				ms.Each(func(m *Match) {
					backup.Add(*m)
				})
				backup.Close()
			}
		}()
	}

	return ms
}

// Count : Returns the total number of records written to disk. Inaccurate unless Each() has been
// called at least once.
func (ms *MatchStore) Count() int {
	return ms.count
}

// Add : Queue up a new match to be written asynchronously.
func (ms *MatchStore) Add(m Match) {
	ms.queue <- m
}

// Each : Extract matches one by one.
func (ms *MatchStore) Each(fn func(*Match)) {
	iter := ms.db.NewIterator(nil, nil)

	for iter.Next() {
		match := MakeMatch(iter.Value())
		fn(match)

		if !ms.countInit {
			ms.count++
		}
	}

	ms.countInit = true
}

// Close : Clean up all related resources. No reads or writes are allowed after this
// function is called.
func (ms *MatchStore) Close() {
	close(ms.queue) // triggers closing of db once queue is empty
	ms.active = false
}

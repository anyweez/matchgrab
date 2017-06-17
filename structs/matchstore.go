package structs

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type MatchStore struct {
	queue     chan Match
	db        *leveldb.DB
	count     int
	countInit bool // becomes true if the count is accurate
	open      bool

	dbLock sync.Mutex
}

func NewMatchStore(filename string) *MatchStore {
	ms := &MatchStore{
		queue:     make(chan Match, 10),
		count:     0,
		countInit: false,
		open:      false,
	}

	var err error
	ms.db, err = leveldb.OpenFile(filename, nil)

	if err != nil {
		panic("Cannot open LevelDB records: " + err.Error())
	}
	ms.open = true

	// Goroutine that asynchronously writes match data.
	go func() {
		for m := range ms.queue {
			if ms.open {
				err := ms.db.Put(m.GameID.Bytes(), m.Bytes(), nil)

				if err != nil {
					panic("Error writing record: " + err.Error())
				}

				ms.count++
			}
		}
	}()

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
func (ms *MatchStore) Each(fn func(Match)) {
	// ms.dbLock.Lock()
	iter := ms.db.NewIterator(nil, nil)

	for iter.Next() {
		value := iter.Value()

		buf := bytes.NewBuffer(value)
		dec := gob.NewDecoder(buf)

		match := Match{}
		dec.Decode(&match)

		fn(match)

		if !ms.countInit {
			ms.count++
		}
	}
	// ms.dbLock.Unlock()

	ms.countInit = true
}

func (ms *MatchStore) Close() {
	ms.open = false
	ms.db.Close()
}

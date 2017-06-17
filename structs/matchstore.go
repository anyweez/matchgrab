package structs

import (
	"bytes"
	"encoding/gob"

	"github.com/syndtr/goleveldb/leveldb"
)

type MatchStore struct {
	queue     chan Match
	db        *leveldb.DB
	count     int
	countInit bool // becomes true if the count is accurate
}

func NewMatchStore(filename string) *MatchStore {
	ms := &MatchStore{
		queue:     make(chan Match, 10),
		count:     0,
		countInit: false,
	}

	ms.db, _ = leveldb.OpenFile(filename, nil)

	// Goroutine that asynchronously writes match data.
	go func() {
		for m := range ms.queue {
			ms.db.Put(m.GameID.Bytes(), m.Bytes(), nil)
			ms.count++
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

	ms.countInit = true
}

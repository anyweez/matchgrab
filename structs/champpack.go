package structs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// packedChampID : Only used in packed data structures related to champions.
type packedChampID int

// ChampPack : Low-level mapping struct used to convert between sparse RiotID's and dense packedChampID's. This
// struct keeps a direct mapping in memory and can convert between the two in a single array lookup, which provides
// roughly a 5.2x speedup in go1.7.1 (see packedarray_test.go benchmarks for experiment).
type ChampPack struct {
	toPacked      []packedChampID
	toUnpacked    []RiotID
	toPackedAdded []bool // has a value been set for toPacked[i]?

	MaxID   RiotID
	MaxSize int
}

// NewChampPack : Return a new ChampPack instance with a max (packed) size of `count` and a maximum
// ID value of `maxID`. For example, NewChampPack(5, 10) means there will be at most five mappings
// added, with the max RiotID being 10.
func NewChampPack(count int, maxID RiotID) *ChampPack {
	return &ChampPack{
		toPacked:      make([]packedChampID, maxID+1), // +1 so the max ID can actually fit
		toPackedAdded: make([]bool, maxID+1),
		toUnpacked:    make([]RiotID, 0, count),

		MaxID:   maxID,
		MaxSize: count,
	}
}

// NewRiotChampPack : Create a new champpack populated with summoner stats. This function makes a
// request to Riot's API for the latest champion list.
func NewRiotChampPack() *ChampPack {
	resp, err := http.Get("http://ddragon.leagueoflegends.com/cdn/7.15.1/data/en_US/champion.json")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var rc struct {
		Data map[string]struct {
			Key  string
			Name string
		}
	}

	json.Unmarshal(body, &rc)

	count := len(rc.Data)
	max := 0

	// Find the max champion ID so we can create the champpack
	for _, champ := range rc.Data {
		id, _ := strconv.Atoi(champ.Key)

		if id > max {
			max = id
		}
	}

	cp := NewChampPack(count, RiotID(max))

	for _, champ := range rc.Data {
		id, _ := strconv.Atoi(champ.Key)

		cp.AddRiotID(RiotID(id))
	}

	return cp
}

// AddRiotID : Add a new Riot ID to the mapping. Returns the corresponding packedChampID.
func (cp *ChampPack) AddRiotID(id RiotID) packedChampID {
	if id <= cp.MaxID && cp.PackedSize() < cp.MaxSize {
		cp.toUnpacked = append(cp.toUnpacked, id)

		packed := packedChampID(len(cp.toUnpacked) - 1)

		cp.toPacked[id] = packed
		cp.toPackedAdded[id] = true

		return packed
	}

	return -1
}

// GetPacked : Get a packedChampID for a previously-added RiotID. The boolean return value indicates whether
// the specified RiotID is known, and if not then the first value should not be trusted.
func (cp *ChampPack) GetPacked(id RiotID) (packedChampID, bool) {
	if int(id) < len(cp.toPacked) {
		return cp.toPacked[id], cp.toPackedAdded[id]
	}

	return 0, false
}

// GetUnpacked : Get previously-added RiotID corresponding to a packedChampID. The boolean return value indicates
// whether the specified RiotID is known, and if not then the first value should not be trusted.
func (cp *ChampPack) GetUnpacked(id packedChampID) (RiotID, bool) {
	if int(id) < len(cp.toUnpacked) {
		return cp.toUnpacked[id], true
	}

	return 0, false
}

// PackedSize : Returns the current number of champions packed in.
func (cp *ChampPack) PackedSize() int {
	return len(cp.toUnpacked)
}

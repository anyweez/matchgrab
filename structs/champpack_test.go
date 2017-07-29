package structs

import (
	"fmt"
	"log"
	"testing"
)

// Test to ensure that we're able to convert Riot ID's to packed and back again.
func TestPackUnpack(t *testing.T) {
	mapping := make(map[RiotID]packedChampID, 20)
	cp := NewChampPack(20, 100)

	ids := []RiotID{1, 7, 4, 13, 21, 80, 81, 54, 19}

	for _, id := range ids {
		mapping[id] = cp.AddRiotID(id)

		// Check that all values are known; precise values checked later.
		_, known := cp.GetPacked(id)
		if !known {
			log.Println(fmt.Sprintf("Added value not recoverable: %d", id))
			t.Fail()
		}
	}

	// Check that the packed <=> unpacked mapping is perfect.
	for up, p := range mapping {
		if unpacked, _ := cp.GetUnpacked(p); up != unpacked {
			t.Fail()
		}

		if packed, _ := cp.GetPacked(up); p != packed {
			t.Fail()
		}
	}
}

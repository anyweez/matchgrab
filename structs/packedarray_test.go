package structs

import "testing"

/** Basic functionality test **/

func TestBasicIO(t *testing.T) {
	cp := NewChampPack(10, RiotID(20))
	pairs := make(map[RiotID]packedChampID, 10)

	riotIDs := []RiotID{1, 2, 3, 6, 9, 10, 11, 14, 15, 20}

	for _, id := range riotIDs {
		pairs[id] = cp.AddRiotID(id)
	}

	for riot, packed := range pairs {
		unpacked, exists := cp.GetUnpacked(packed)

		if !exists || unpacked != riot {
			t.Fail()
		}
	}
}

func createPacker() *ChampPack {
	cp := NewChampPack(10, RiotID(30))

	cp.AddRiotID(2)
	cp.AddRiotID(4)
	cp.AddRiotID(5)
	cp.AddRiotID(8)
	cp.AddRiotID(10)
	cp.AddRiotID(14)
	cp.AddRiotID(15)
	cp.AddRiotID(16)
	cp.AddRiotID(17)
	cp.AddRiotID(18)

	return cp
}

func exec(pack bool, b *testing.B) {
	b.StopTimer()

	packer := createPacker()
	store := NewMatchStore("/Volumes/LOLMatches/matches.db")

	count := 0
	store.Each(func(m *Match) {
		b.StartTimer()

		if pack {
			m.Pack(packer)
		}

		for i := 0; i < 150; i++ {
			m.Banned(RiotID(i))
			m.Picked(RiotID(i))
			m.Won(RiotID(i))
		}

		b.StopTimer()

		count++

		if count >= b.N {
			return
		}
	})
	store.Close()
}

func BenchmarkUnpacked(b *testing.B) {
	exec(false, b)
}

func BenchmarkPacked(b *testing.B) {
	exec(true, b)
}

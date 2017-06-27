package structs

import "errors"

type PackedChampBooleanArray struct {
	packer *ChampPack

	champs []bool
}

func NewPackedChampBooleanArray(packer *ChampPack) *PackedChampBooleanArray {
	return &PackedChampBooleanArray{
		packer: packer,
		champs: make([]bool, packer.MaxSize),
	}
}

// Get : Returns the boolean value at the specified index, as well as a second boolean
// indicating whether the value exists. If the second value is false then the first should
// not be trusted.
func (pcba *PackedChampBooleanArray) Get(id RiotID) (bool, bool) {
	index, exists := pcba.packer.GetPacked(id)

	if exists && int(index) < len(pcba.champs) {
		return pcba.champs[index], true
	}

	return false, false
}

func (pcba *PackedChampBooleanArray) Set(id RiotID, val bool) error {
	index, exists := pcba.packer.GetPacked(id)

	if exists {
		pcba.champs[index] = val
		return nil
	}

	return errors.New("Nonexistant Riot ID specified.")
}

func (pcba *PackedChampBooleanArray) Each(fn func(id RiotID, val bool)) {
	for packed, val := range pcba.champs {
		unpacked, exists := pcba.packer.GetUnpacked(packedChampID(packed))

		// TODO: what to do if this value doesn't exist?
		if exists {
			fn(
				unpacked,
				val, // boolean value
			)
		}
	}
}

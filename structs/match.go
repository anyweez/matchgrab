package structs

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"time"
)

// APIMatch : raw data returned from Riot's API. Converted to Match using ToMatch function
type APIMatch struct {
	GameID       RiotID `json:"gameId"`
	SeasonID     int    `json:"seasonId"`
	GameCreation int64  `json:"gameCreation"`
	GameDuration int    `json:"gameDuration"`

	Participants []struct {
		TeamID     int    `json:"teamId"`
		ChampionID RiotID `json:"championId"`
		Stats      struct {
			Win bool `json:"win"`
		} `json:"stats"`
	} `json:"participants"`

	ParticipantIdentities []struct {
		Player struct {
			AccountID    RiotID `json:"accountId"`
			SummonerName string `json:"summonerName"`
			SummonerID   RiotID `json:"summonerId"`
			ProfileIcon  int    `json:"profileIcon"`
		} `json:"player"`
	}

	Teams []struct {
		Bans []struct {
			ChampionID RiotID `json:"championId"`
		} `json:"bans"`
	}

	GameMode string `json:"gameMode"`
	MapID    int    `json:"mapId"`
	GameType string `json:"gameType"`
}

// RiotID : Canonical identifier for everything that comes from Riot, including summoner ID's, champion ID's,
// and account ID's.
type RiotID int64

func (r RiotID) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r)

	return buf.Bytes()
}

type Match struct {
	GameID       RiotID `json:"gameId"`
	SeasonID     int    `json:"seasonId"`
	GameCreation int64  `json:"gameCreation"`
	GameDuration int    `json:"gameDuration"`

	Participants []Participant
	Bans         []RiotID

	GameMode string `json:"gameMode"`
	MapID    int    `json:"mapId"`
	GameType string `json:"gameType"`

	packed       bool
	packedBans   *PackedChampBooleanArray
	packedPicked *PackedChampBooleanArray
	packedWon    *PackedChampBooleanArray
}

func (m *Match) Pack(packer *ChampPack) {
	if m.packed {
		return
	}

	m.packed = true

	m.packedBans = NewPackedChampBooleanArray(packer)
	for _, b := range m.Bans {
		// TODO: currently getting random -1's in ban list. Remove @ retrieval?
		if b > 0 {
			m.packedBans.Set(b, true)
		}
	}

	m.packedPicked = NewPackedChampBooleanArray(packer)
	for _, p := range m.Participants {
		if p.ChampionID > 0 {
			m.packedPicked.Set(p.ChampionID, true)
		}
	}

	m.packedWon = NewPackedChampBooleanArray(packer)
	for _, p := range m.Participants {
		if p.ChampionID > 0 {
			m.packedWon.Set(p.ChampionID, p.Winner)
		}
	}
}

func (m *Match) When() time.Time {
	return time.Unix(m.GameCreation/1000, 0)
}

func (m *Match) Banned(id RiotID) bool {
	// Constant time if packed, linear if not packed
	if m.packed {
		val, exists := m.packedBans.Get(id)

		return exists && val
	}

	for _, ban := range m.Bans {
		if ban == id {
			return true
		}
	}

	return false
}

// Picked : Returns a boolean indicating whether the specified champion played in this game.
func (m *Match) Picked(id RiotID) bool {
	// Constant time if packed, linear if not packed
	if m.packed {
		val, exists := m.packedPicked.Get(id)

		return exists && val
	}

	for _, p := range m.Participants {
		if id == p.ChampionID {
			return true
		}
	}

	return false
}

// Won : Returns a boolean indicating whether the specified champion won the game.
func (m *Match) Won(id RiotID) bool {
	if m.packed {
		val, exists := m.packedWon.Get(id)

		return exists && val
	}

	for _, p := range m.Participants {
		if id == p.ChampionID && p.Winner {
			return true
		}
	}

	return false
}

func (m Match) Bytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	enc.Encode(m)
	return buf.Bytes()
}

type Participant struct {
	SummonerName string `json:"summonerName"`
	AccountID    RiotID `json:"accountId"`
	ProfileIcon  int    `json:"profileIcon"`
	SummonerID   RiotID `json:"summonerId"`
	ChampionID   RiotID `json:"championId"`
	TeamID       int    `json:"teamId"`

	Winner bool `json:"winner"`
}

// ToMatch : Convert raw API data to a Match object
func ToMatch(raw APIMatch) Match {
	var match Match

	match.GameID = raw.GameID
	match.SeasonID = raw.SeasonID
	match.GameCreation = raw.GameCreation
	match.GameDuration = raw.GameDuration

	match.GameMode = raw.GameMode
	match.MapID = raw.MapID
	match.GameType = raw.GameType

	match.Participants = make([]Participant, len(raw.Participants))

	for i := 0; i < len(raw.Participants); i++ {
		match.Participants[i] = Participant{
			TeamID:     raw.Participants[i].TeamID,
			ChampionID: raw.Participants[i].ChampionID,
			AccountID:  raw.ParticipantIdentities[i].Player.AccountID,
			SummonerID: raw.ParticipantIdentities[i].Player.SummonerID,

			ProfileIcon:  raw.ParticipantIdentities[i].Player.ProfileIcon,
			SummonerName: raw.ParticipantIdentities[i].Player.SummonerName,
			Winner:       raw.Participants[i].Stats.Win,
		}
	}

	match.Bans = make([]RiotID, 0)
	for _, team := range raw.Teams {
		for _, ban := range team.Bans {
			match.Bans = append(match.Bans, ban.ChampionID)
		}
	}

	return match
}

package structs

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/anyweez/matchgrab/config"
	protostruct "github.com/anyweez/matchgrab/proto"
	"github.com/golang/protobuf/proto"
)

// This file contains definitions for the core data structures in matchgrab including Match and
// Participant, as well as all functions that encode to / decode from those two structs. There are
// also a handful of utility functions defined here, such as Match.Won(RiotID) and Match.Pack().
//
// A lot of this code is somewhat redundant because I have to copy between APIMatch, Match, and
// protostruct.Match. They're all fairly similar in the information they hold, but the protobuf
// version should be internal to this module only. Ideally we could cut back some of this but I'm
// not sure its going to be a good idea in the long run.

type rawMastery struct {
	MasteryID int32 `json:"masteryId"`
}

type rawRune struct {
	RuneID int32 `json:"runeId"`
}

// APIMatch : Raw data returned from Riot's API. Converted to Match using ToMatch() function.
type APIMatch struct {
	GameID       RiotID `json:"gameId"`
	SeasonID     int    `json:"seasonId"`
	GameCreation int64  `json:"gameCreation"`
	GameDuration int    `json:"gameDuration"`

	Participants []struct {
		TeamID     int          `json:"teamId"`
		ChampionID RiotID       `json:"championId"`
		Masteries  []rawMastery `json:"masteries"`
		Runes      []rawRune    `json:"runes"`

		Stats struct {
			Win   bool `json:"win"`
			Item0 int32
			Item1 int32
			Item2 int32
			Item3 int32
			Item4 int32
			Item5 int32
			Item6 int32

			Kills                           int32
			Deaths                          int32
			Assists                         int32
			LargestKillingSpree             int32
			LargestMultiKill                int32
			KillingSprees                   int32
			LongestTimeSpentLiving          int32
			DoubleKills                     int32
			TripleKills                     int32
			QuadraKills                     int32
			PentaKills                      int32
			UnrealKills                     int32
			TotalDamageDealt                int32
			MagicDamageDealt                int32
			PhysicalDamageDealt             int32
			TrueDamageDealt                 int32
			LargestCriticalStrike           int32
			TotalDamageDealtToChampions     int32
			MagicDamageDealtToChampions     int32
			PhysicalDamageDealtToChampions  int32
			TrueDamageDealtToChampions      int32
			TotalHeal                       int32
			TotalUnitsHealed                int32
			DamageSelfMitigated             int32
			DamageDealtToObjectives         int32
			DamageDealtToTurrets            int32
			VisionScore                     int32
			TimeCCingOthers                 int32
			TotalDamageTaken                int32
			MagicalDamageTaken              int32
			PhysicalDamageTaken             int32
			TrueDamageTaken                 int32
			GoldEarned                      int32
			GoldSpent                       int32
			TurretKills                     int32
			InhibitorKills                  int32
			TotalMinionsKilled              int32
			NeutralMinionsKilled            int32
			NeutralMinionsKilledTeamJungle  int32
			NeutralMinionsKilledEnemyJungle int32
			TotalTimeCrowdControlDealt      int32
			ChampLevel                      int32
			VisionWardsBoughtInGame         int32
			SightWardsBoughtInGame          int32
			WardsPlaced                     int32
			WardsKilled                     int32
			FirstBloodKill                  bool
			FirstBloodAssist                bool
			FirstTowerKill                  bool
			FirstTowerAssist                bool
			FirstInhibitorKill              bool
			FirstInhibitorAssist            bool
			CombatPlayerScore               int32
			ObjectivePlayerScore            int32
			TotalPlayerScore                int32
			TotalScoreRank                  int32
		} `json:"stats"`
	}

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

// RiotID : Canonical identifier for everything that comes from Riot, including summoner ID's,
// champion ID's, and account ID's.
type RiotID int64

// Bytes : Encode RiotID as bytes.
func (r RiotID) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r)

	return buf.Bytes()
}

// Match : Primary structure used to store match information. Generated from APIMatch's using
// ToMatch(), and can be encoded into a compact binary format for storage using Match.Bytes().
//
// This struct stores all information related to an individual match, including summoner stats
// if Config.KeepStats is enabled.
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

// Pack : Improve lookup rates for bans, picks, and wins.
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

// Bytes : Output as protocol buffer-encoded byte array.
func (m Match) Bytes() []byte {
	bans := make([]int64, 0, len(m.Bans))

	for _, b := range m.Bans {
		bans = append(bans, int64(b))
	}

	participants := make([]*protostruct.Participant, 0, len(m.Participants))

	for _, p := range m.Participants {
		var stats *protostruct.ParticipantStats
		if p.Stats != nil {
			stats = &protostruct.ParticipantStats{
				Masteries: p.Masteries,
				Runes:     p.Runes,
				Items:     p.Items,

				Kills:                           p.Stats.Kills,
				Deaths:                          p.Stats.Deaths,
				Assists:                         p.Stats.Assists,
				LargestKillingSpree:             p.Stats.LargestKillingSpree,
				LargestMultiKill:                p.Stats.LargestMultiKill,
				KillingSprees:                   p.Stats.KillingSprees,
				LongestTimeSpentLiving:          p.Stats.LongestTimeSpentLiving,
				DoubleKills:                     p.Stats.DoubleKills,
				TripleKills:                     p.Stats.TripleKills,
				QuadraKills:                     p.Stats.QuadraKills,
				PentaKills:                      p.Stats.PentaKills,
				UnrealKills:                     p.Stats.UnrealKills,
				TotalDamageDealt:                p.Stats.TotalDamageDealt,
				MagicDamageDealt:                p.Stats.MagicDamageDealt,
				PhysicalDamageDealt:             p.Stats.PhysicalDamageDealt,
				TrueDamageDealt:                 p.Stats.TrueDamageDealt,
				LargestCriticalStrike:           p.Stats.LargestCriticalStrike,
				TotalDamageDealtToChampions:     p.Stats.TotalDamageDealtToChampions,
				MagicDamageDealtToChampions:     p.Stats.MagicDamageDealtToChampions,
				PhysicalDamageDealtToChampions:  p.Stats.PhysicalDamageDealtToChampions,
				TrueDamageDealtToChampions:      p.Stats.TrueDamageDealtToChampions,
				TotalHeal:                       p.Stats.TotalHeal,
				TotalUnitsHealed:                p.Stats.TotalUnitsHealed,
				DamageSelfMitigated:             p.Stats.DamageSelfMitigated,
				DamageDealtToObjectives:         p.Stats.DamageDealtToObjectives,
				DamageDealtToTurrets:            p.Stats.DamageDealtToTurrets,
				VisionScore:                     p.Stats.VisionScore,
				TimeCCingOthers:                 p.Stats.TimeCCingOthers,
				TotalDamageTaken:                p.Stats.TotalDamageTaken,
				MagicalDamageTaken:              p.Stats.MagicalDamageTaken,
				PhysicalDamageTaken:             p.Stats.PhysicalDamageTaken,
				TrueDamageTaken:                 p.Stats.TrueDamageTaken,
				GoldEarned:                      p.Stats.GoldEarned,
				GoldSpent:                       p.Stats.GoldSpent,
				TurretKills:                     p.Stats.TurretKills,
				InhibitorKills:                  p.Stats.InhibitorKills,
				TotalMinionsKilled:              p.Stats.TotalMinionsKilled,
				NeutralMinionsKilled:            p.Stats.NeutralMinionsKilled,
				NeutralMinionsKilledTeamJungle:  p.Stats.NeutralMinionsKilledTeamJungle,
				NeutralMinionsKilledEnemyJungle: p.Stats.NeutralMinionsKilledEnemyJungle,
				TotalTimeCrowdControlDealt:      p.Stats.TotalTimeCrowdControlDealt,
				ChampLevel:                      p.Stats.ChampLevel,
				VisionWardsBoughtInGame:         p.Stats.VisionWardsBoughtInGame,
				SightWardsBoughtInGame:          p.Stats.SightWardsBoughtInGame,
				WardsPlaced:                     p.Stats.WardsPlaced,
				WardsKilled:                     p.Stats.WardsKilled,
				FirstBloodKill:                  p.Stats.FirstBloodKill,
				FirstBloodAssist:                p.Stats.FirstBloodAssist,
				FirstTowerKill:                  p.Stats.FirstTowerKill,
				FirstTowerAssist:                p.Stats.FirstTowerAssist,
				FirstInhibitorKill:              p.Stats.FirstInhibitorKill,
				FirstInhibitorAssist:            p.Stats.FirstInhibitorAssist,
				CombatPlayerScore:               p.Stats.CombatPlayerScore,
				ObjectivePlayerScore:            p.Stats.ObjectivePlayerScore,
				TotalPlayerScore:                p.Stats.TotalPlayerScore,
				TotalScoreRank:                  p.Stats.TotalScoreRank,
			}
		}

		participants = append(participants, &protostruct.Participant{
			SummonerName: p.SummonerName,
			AccountID:    int64(p.AccountID),
			ProfileIcon:  int32(p.ProfileIcon),
			SummonerID:   int64(p.SummonerID),
			ChampionID:   int64(p.ChampionID),
			TeamID:       int32(p.TeamID),
			Winner:       p.Winner,

			Stats: stats,
		})
	}

	p := &protostruct.Match{
		GameID:       int64(m.GameID),
		SeasonID:     int32(m.SeasonID),
		GameCreation: m.GameCreation,
		GameDuration: int32(m.GameDuration),
		Participants: participants,
		Bans:         bans,

		GameMode: m.GameMode,
		MapID:    int32(m.MapID),
		GameType: m.GameType,
	}

	buf, _ := proto.Marshal(p)

	return buf
}

// MakeMatch : Convert an encoded byte array back into a match. This is the inverse
// of Match.Bytes().
func MakeMatch(buf []byte) *Match {
	pm := protostruct.Match{}

	proto.Unmarshal(buf, &pm)

	// Convert ban list
	bans := make([]RiotID, 0, len(pm.Bans))
	for _, b := range pm.Bans {
		bans = append(bans, RiotID(b))
	}

	// Convert participant list
	participants := make([]Participant, 0, len(pm.Participants))
	for _, p := range pm.Participants {
		var stats *ParticipantStats

		if p.Stats != nil {
			stats = &ParticipantStats{
				Kills:                           p.Stats.GetKills(),
				Deaths:                          p.Stats.GetDeaths(),
				Assists:                         p.Stats.GetAssists(),
				LargestKillingSpree:             p.Stats.GetLargestKillingSpree(),
				LargestMultiKill:                p.Stats.GetLargestMultiKill(),
				KillingSprees:                   p.Stats.GetKillingSprees(),
				LongestTimeSpentLiving:          p.Stats.GetLongestTimeSpentLiving(),
				DoubleKills:                     p.Stats.GetDoubleKills(),
				TripleKills:                     p.Stats.GetTripleKills(),
				QuadraKills:                     p.Stats.GetQuadraKills(),
				PentaKills:                      p.Stats.GetPentaKills(),
				UnrealKills:                     p.Stats.GetUnrealKills(),
				TotalDamageDealt:                p.Stats.GetTotalDamageDealt(),
				MagicDamageDealt:                p.Stats.GetMagicDamageDealt(),
				PhysicalDamageDealt:             p.Stats.GetPhysicalDamageDealt(),
				TrueDamageDealt:                 p.Stats.GetTrueDamageDealt(),
				LargestCriticalStrike:           p.Stats.GetLargestCriticalStrike(),
				TotalDamageDealtToChampions:     p.Stats.GetTotalDamageDealtToChampions(),
				MagicDamageDealtToChampions:     p.Stats.GetMagicDamageDealtToChampions(),
				PhysicalDamageDealtToChampions:  p.Stats.GetPhysicalDamageDealtToChampions(),
				TrueDamageDealtToChampions:      p.Stats.GetTrueDamageDealtToChampions(),
				TotalHeal:                       p.Stats.GetTotalHeal(),
				TotalUnitsHealed:                p.Stats.GetTotalUnitsHealed(),
				DamageSelfMitigated:             p.Stats.GetDamageSelfMitigated(),
				DamageDealtToObjectives:         p.Stats.GetDamageDealtToObjectives(),
				DamageDealtToTurrets:            p.Stats.GetDamageDealtToTurrets(),
				VisionScore:                     p.Stats.GetVisionScore(),
				TimeCCingOthers:                 p.Stats.GetTimeCCingOthers(),
				TotalDamageTaken:                p.Stats.GetTotalDamageTaken(),
				MagicalDamageTaken:              p.Stats.GetMagicalDamageTaken(),
				PhysicalDamageTaken:             p.Stats.GetPhysicalDamageTaken(),
				TrueDamageTaken:                 p.Stats.GetTrueDamageTaken(),
				GoldEarned:                      p.Stats.GetGoldEarned(),
				GoldSpent:                       p.Stats.GetGoldSpent(),
				TurretKills:                     p.Stats.GetTurretKills(),
				InhibitorKills:                  p.Stats.GetInhibitorKills(),
				TotalMinionsKilled:              p.Stats.GetTotalMinionsKilled(),
				NeutralMinionsKilled:            p.Stats.GetNeutralMinionsKilled(),
				NeutralMinionsKilledTeamJungle:  p.Stats.GetNeutralMinionsKilledTeamJungle(),
				NeutralMinionsKilledEnemyJungle: p.Stats.GetNeutralMinionsKilledEnemyJungle(),
				TotalTimeCrowdControlDealt:      p.Stats.GetTotalTimeCrowdControlDealt(),
				ChampLevel:                      p.Stats.GetChampLevel(),
				VisionWardsBoughtInGame:         p.Stats.GetVisionWardsBoughtInGame(),
				SightWardsBoughtInGame:          p.Stats.GetSightWardsBoughtInGame(),
				WardsPlaced:                     p.Stats.GetWardsPlaced(),
				WardsKilled:                     p.Stats.GetWardsKilled(),
				FirstBloodKill:                  p.Stats.GetFirstBloodKill(),
				FirstBloodAssist:                p.Stats.GetFirstBloodAssist(),
				FirstTowerKill:                  p.Stats.GetFirstTowerKill(),
				FirstTowerAssist:                p.Stats.GetFirstTowerAssist(),
				FirstInhibitorKill:              p.Stats.GetFirstInhibitorKill(),
				FirstInhibitorAssist:            p.Stats.GetFirstInhibitorAssist(),
				CombatPlayerScore:               p.Stats.GetCombatPlayerScore(),
				ObjectivePlayerScore:            p.Stats.GetObjectivePlayerScore(),
				TotalPlayerScore:                p.Stats.GetTotalPlayerScore(),
				TotalScoreRank:                  p.Stats.GetTotalScoreRank(),
			}
		}

		participants = append(participants, Participant{
			SummonerName: p.GetSummonerName(),
			AccountID:    RiotID(p.GetAccountID()),
			ProfileIcon:  int(p.GetProfileIcon()),
			SummonerID:   RiotID(p.GetSummonerID()),
			ChampionID:   RiotID(p.GetChampionID()),
			TeamID:       int(p.GetTeamID()),
			Winner:       p.GetWinner(),

			Stats: stats,
		})
	}

	m := &Match{
		GameID:       RiotID(pm.GetGameID()),
		SeasonID:     int(pm.GetSeasonID()),
		GameCreation: pm.GetGameCreation(),
		GameDuration: int(pm.GetGameDuration()),
		Participants: participants,
		Bans:         bans,
		GameMode:     pm.GetGameMode(),
		MapID:        int(pm.GetMapID()),
		GameType:     pm.GetGameType(),
	}

	return m
}

// Participant : Stores information about individual players, including stats if requested.
type Participant struct {
	SummonerName string `json:"summonerName"`
	AccountID    RiotID `json:"accountId"`
	ProfileIcon  int    `json:"profileIcon"`
	SummonerID   RiotID `json:"summonerId"`
	ChampionID   RiotID `json:"championId"`
	TeamID       int    `json:"teamId"`

	Winner bool `json:"winner"`

	Masteries []int32
	Runes     []int32
	Items     []int32

	Stats *ParticipantStats
}

type ParticipantStats struct {
	Kills                           int32
	Deaths                          int32
	Assists                         int32
	LargestKillingSpree             int32
	LargestMultiKill                int32
	KillingSprees                   int32
	LongestTimeSpentLiving          int32
	DoubleKills                     int32
	TripleKills                     int32
	QuadraKills                     int32
	PentaKills                      int32
	UnrealKills                     int32
	TotalDamageDealt                int32
	MagicDamageDealt                int32
	PhysicalDamageDealt             int32
	TrueDamageDealt                 int32
	LargestCriticalStrike           int32
	TotalDamageDealtToChampions     int32
	MagicDamageDealtToChampions     int32
	PhysicalDamageDealtToChampions  int32
	TrueDamageDealtToChampions      int32
	TotalHeal                       int32
	TotalUnitsHealed                int32
	DamageSelfMitigated             int32
	DamageDealtToObjectives         int32
	DamageDealtToTurrets            int32
	VisionScore                     int32
	TimeCCingOthers                 int32
	TotalDamageTaken                int32
	MagicalDamageTaken              int32
	PhysicalDamageTaken             int32
	TrueDamageTaken                 int32
	GoldEarned                      int32
	GoldSpent                       int32
	TurretKills                     int32
	InhibitorKills                  int32
	TotalMinionsKilled              int32
	NeutralMinionsKilled            int32
	NeutralMinionsKilledTeamJungle  int32
	NeutralMinionsKilledEnemyJungle int32
	TotalTimeCrowdControlDealt      int32
	ChampLevel                      int32
	VisionWardsBoughtInGame         int32
	SightWardsBoughtInGame          int32
	WardsPlaced                     int32
	WardsKilled                     int32
	FirstBloodKill                  bool
	FirstBloodAssist                bool
	FirstTowerKill                  bool
	FirstTowerAssist                bool
	FirstInhibitorKill              bool
	FirstInhibitorAssist            bool
	CombatPlayerScore               int32
	ObjectivePlayerScore            int32
	TotalPlayerScore                int32
	TotalScoreRank                  int32
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

	for i, p := range raw.Participants {
		pi := raw.ParticipantIdentities[i]

		var stats *ParticipantStats

		// Keep stats if desired.
		if config.Config.KeepStats {
			stats = &ParticipantStats{
				Kills:                           p.Stats.Kills,
				Deaths:                          p.Stats.Deaths,
				Assists:                         p.Stats.Assists,
				LargestKillingSpree:             p.Stats.LargestKillingSpree,
				LargestMultiKill:                p.Stats.LargestMultiKill,
				KillingSprees:                   p.Stats.KillingSprees,
				LongestTimeSpentLiving:          p.Stats.LongestTimeSpentLiving,
				DoubleKills:                     p.Stats.DoubleKills,
				TripleKills:                     p.Stats.TripleKills,
				QuadraKills:                     p.Stats.QuadraKills,
				PentaKills:                      p.Stats.PentaKills,
				UnrealKills:                     p.Stats.UnrealKills,
				TotalDamageDealt:                p.Stats.TotalDamageDealt,
				MagicDamageDealt:                p.Stats.MagicDamageDealt,
				PhysicalDamageDealt:             p.Stats.PhysicalDamageDealt,
				TrueDamageDealt:                 p.Stats.TrueDamageDealt,
				LargestCriticalStrike:           p.Stats.LargestCriticalStrike,
				TotalDamageDealtToChampions:     p.Stats.TotalDamageDealtToChampions,
				MagicDamageDealtToChampions:     p.Stats.MagicDamageDealtToChampions,
				PhysicalDamageDealtToChampions:  p.Stats.PhysicalDamageDealtToChampions,
				TrueDamageDealtToChampions:      p.Stats.TrueDamageDealtToChampions,
				TotalHeal:                       p.Stats.TotalHeal,
				TotalUnitsHealed:                p.Stats.TotalUnitsHealed,
				DamageSelfMitigated:             p.Stats.DamageSelfMitigated,
				DamageDealtToObjectives:         p.Stats.DamageDealtToObjectives,
				DamageDealtToTurrets:            p.Stats.DamageDealtToTurrets,
				VisionScore:                     p.Stats.VisionScore,
				TimeCCingOthers:                 p.Stats.TimeCCingOthers,
				TotalDamageTaken:                p.Stats.TotalDamageTaken,
				MagicalDamageTaken:              p.Stats.MagicalDamageTaken,
				PhysicalDamageTaken:             p.Stats.PhysicalDamageTaken,
				TrueDamageTaken:                 p.Stats.TrueDamageTaken,
				GoldEarned:                      p.Stats.GoldEarned,
				GoldSpent:                       p.Stats.GoldSpent,
				TurretKills:                     p.Stats.TurretKills,
				InhibitorKills:                  p.Stats.InhibitorKills,
				TotalMinionsKilled:              p.Stats.TotalMinionsKilled,
				NeutralMinionsKilled:            p.Stats.NeutralMinionsKilled,
				NeutralMinionsKilledTeamJungle:  p.Stats.NeutralMinionsKilledTeamJungle,
				NeutralMinionsKilledEnemyJungle: p.Stats.NeutralMinionsKilledEnemyJungle,
				TotalTimeCrowdControlDealt:      p.Stats.TotalTimeCrowdControlDealt,
				ChampLevel:                      p.Stats.ChampLevel,
				VisionWardsBoughtInGame:         p.Stats.VisionWardsBoughtInGame,
				SightWardsBoughtInGame:          p.Stats.SightWardsBoughtInGame,
				WardsPlaced:                     p.Stats.WardsPlaced,
				WardsKilled:                     p.Stats.WardsKilled,
				FirstBloodKill:                  p.Stats.FirstBloodKill,
				FirstBloodAssist:                p.Stats.FirstBloodAssist,
				FirstTowerKill:                  p.Stats.FirstTowerKill,
				FirstTowerAssist:                p.Stats.FirstTowerAssist,
				FirstInhibitorKill:              p.Stats.FirstInhibitorKill,
				FirstInhibitorAssist:            p.Stats.FirstInhibitorAssist,
				CombatPlayerScore:               p.Stats.CombatPlayerScore,
				TotalPlayerScore:                p.Stats.TotalPlayerScore,
				TotalScoreRank:                  p.Stats.TotalScoreRank,
			}
		}

		match.Participants[i] = Participant{
			TeamID:     p.TeamID,
			ChampionID: p.ChampionID,
			AccountID:  pi.Player.AccountID,
			SummonerID: pi.Player.SummonerID,

			ProfileIcon:  pi.Player.ProfileIcon,
			SummonerName: pi.Player.SummonerName,
			Winner:       p.Stats.Win,

			Masteries: make([]int32, 0),
			Runes:     make([]int32, 0),
			Items: []int32{
				p.Stats.Item0,
				p.Stats.Item1,
				p.Stats.Item2,
				p.Stats.Item3,
				p.Stats.Item4,
				p.Stats.Item5,
				p.Stats.Item6,
			},

			Stats: stats,
		}

		for _, m := range p.Masteries {
			match.Participants[i].Masteries = append(match.Participants[i].Masteries, m.MasteryID)
		}

		for _, r := range p.Runes {
			match.Participants[i].Runes = append(match.Participants[i].Runes, r.RuneID)
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

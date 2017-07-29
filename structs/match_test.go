package structs

import (
	"testing"

	"github.com/anyweez/matchgrab/config"
)

type champStatus struct {
	Banned bool
	Picked bool
	Won    bool
}

func outcomes() map[RiotID]champStatus {
	return map[RiotID]champStatus{
		122: champStatus{Banned: true, Picked: false, Won: false},
		164: champStatus{Banned: true, Picked: false, Won: false},
		23:  champStatus{Banned: true, Picked: false, Won: false},
		50:  champStatus{Banned: false, Picked: false, Won: false},
		3:   champStatus{Banned: false, Picked: true, Won: true},
		19:  champStatus{Banned: true, Picked: false, Won: false},
	}
}

// Make sure we are encoding + decoding to storage format correctly.
func TestToFromProto(t *testing.T) {
	samples := rawSamples()

	for _, sample := range samples {
		match := ToMatch(sample)

		buf := match.Bytes()
		m2 := MakeMatch(buf)

		if match.GameID != m2.GameID {
			t.Fail()
		}

		if match.SeasonID != m2.SeasonID {
			t.Fail()
		}

		if match.GameCreation != m2.GameCreation {
			t.Fail()
		}

		if match.GameDuration != m2.GameDuration {
			t.Fail()
		}

		// TODO: Check participants and ban lists as well.
	}
}

// TODO: save some raw json match data for next tests

// Ensure we correctly parse raw API responses into Match's.
func TestCanParse(t *testing.T) {
	matches := rawSamples()

	for _, sample := range matches {
		match := ToMatch(sample)

		if match.GameID != sample.GameID {
			t.Fail()
		}

		if match.SeasonID != sample.SeasonID {
			t.Fail()
		}

		if match.GameCreation != sample.GameCreation {
			t.Fail()
		}

		if match.GameDuration != sample.GameDuration {
			t.Fail()
		}
	}
}

// Test Banned(), Picked(), and Won() with unpacked matches.
func TestCountsUnpacked(t *testing.T) {
	samples := rawSamples()
	matches := make([]Match, 0, len(samples))

	for _, sample := range samples {
		matches = append(matches, ToMatch(sample))
	}

	oc := outcomes()

	for id, outcome := range oc {
		if matches[0].Banned(id) != outcome.Banned {
			t.Fail()
		}

		if matches[0].Picked(id) != outcome.Picked {
			t.Fail()
		}

		if matches[0].Won(id) != outcome.Won {
			t.Fail()
		}
	}
}

// TestFetchStats : Ensure we at least create a stats struct when we specify we
// want to create stats.
func TestFetchStats(t *testing.T) {
	config.Setup()
	config.Config.KeepStats = true

	raws := rawSamples()

	for _, raw := range raws {
		m := ToMatch(raw)

		for _, p := range m.Participants {
			if p.Stats == nil {
				t.Fail()
			}
		}
	}
}

// TestNoStats : Ensure we don't create a stats struct when we specify we don't want stats.
func TestNoStats(t *testing.T) {
	config.Setup()
	config.Config.KeepStats = false

	raws := rawSamples()

	for _, raw := range raws {
		m := ToMatch(raw)

		for _, p := range m.Participants {
			if p.Stats != nil {
				t.Fail()
			}
		}
	}
}

func TestPackStats(t *testing.T) {
	cp := NewRiotChampPack()

	for _, r := range rawSamples() {
		packed := ToMatch(r)
		unpacked := ToMatch(r)

		packed.Pack(cp)

		// Make sure all the same champs were picked, banned, and marked as winners.
		for i := 1; i < 300; i++ {
			riotID := RiotID(i)

			if packed.Picked(riotID) != unpacked.Picked(riotID) {
				t.Fail()
			}

			if packed.Banned(riotID) != unpacked.Banned(riotID) {
				t.Fail()
			}

			if packed.Won(riotID) != unpacked.Won(riotID) {
				t.Fail()
			}
		}
	}
}

// TestAPIMatchMapping : Ensure that ToMatch() copies all individual stats from APIMatch to
// Match objects.
func TestAPIMatchMapping(t *testing.T) {
	config.Setup()
	config.Config.KeepStats = true

	raws := rawSamples()

	for _, raw := range raws {
		m := ToMatch(raw)

		for i, p := range m.Participants {
			rp := m.Participants[i]

			if p.Stats.Assists != rp.Stats.Assists {
				t.Fail()
			}

			if p.Stats.ChampLevel != rp.Stats.ChampLevel {
				t.Fail()
			}

			if p.Stats.CombatPlayerScore != rp.Stats.CombatPlayerScore {
				t.Fail()
			}

			if p.Stats.DamageDealtToObjectives != rp.Stats.DamageDealtToObjectives {
				t.Fail()
			}

			if p.Stats.DamageDealtToTurrets != rp.Stats.DamageDealtToTurrets {
				t.Fail()
			}

			if p.Stats.DamageSelfMitigated != rp.Stats.DamageSelfMitigated {
				t.Fail()
			}

			if p.Stats.Deaths != rp.Stats.Deaths {
				t.Fail()
			}

			if p.Stats.DoubleKills != rp.Stats.DoubleKills {
				t.Fail()
			}

			if p.Stats.FirstBloodAssist != rp.Stats.FirstBloodAssist {
				t.Fail()
			}

			if p.Stats.FirstBloodKill != rp.Stats.FirstBloodKill {
				t.Fail()
			}

			if p.Stats.FirstInhibitorAssist != rp.Stats.FirstInhibitorAssist {
				t.Fail()
			}

			if p.Stats.FirstInhibitorKill != rp.Stats.FirstInhibitorKill {
				t.Fail()
			}

			if p.Stats.FirstTowerAssist != rp.Stats.FirstTowerAssist {
				t.Fail()
			}

			if p.Stats.FirstTowerKill != rp.Stats.FirstTowerKill {
				t.Fail()
			}

			if p.Stats.GoldEarned != rp.Stats.GoldEarned {
				t.Fail()
			}

			if p.Stats.GoldSpent != rp.Stats.GoldSpent {
				t.Fail()
			}

			if p.Stats.InhibitorKills != rp.Stats.InhibitorKills {
				t.Fail()
			}

			if p.Stats.KillingSprees != rp.Stats.KillingSprees {
				t.Fail()
			}

			if p.Stats.Kills != rp.Stats.Kills {
				t.Fail()
			}

			if p.Stats.LargestCriticalStrike != rp.Stats.LargestCriticalStrike {
				t.Fail()
			}

			if p.Stats.LargestKillingSpree != rp.Stats.LargestKillingSpree {
				t.Fail()
			}

			if p.Stats.LargestMultiKill != rp.Stats.LargestMultiKill {
				t.Fail()
			}

			if p.Stats.LongestTimeSpentLiving != rp.Stats.LongestTimeSpentLiving {
				t.Fail()
			}

			if p.Stats.MagicalDamageTaken != rp.Stats.MagicalDamageTaken {
				t.Fail()
			}

			if p.Stats.MagicDamageDealt != rp.Stats.MagicDamageDealt {
				t.Fail()
			}

			if p.Stats.MagicDamageDealtToChampions != rp.Stats.MagicDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilled != rp.Stats.NeutralMinionsKilled {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilledEnemyJungle != rp.Stats.NeutralMinionsKilledEnemyJungle {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilledTeamJungle != rp.Stats.NeutralMinionsKilledTeamJungle {
				t.Fail()
			}

			if p.Stats.ObjectivePlayerScore != rp.Stats.ObjectivePlayerScore {
				t.Fail()
			}

			if p.Stats.PentaKills != rp.Stats.PentaKills {
				t.Fail()
			}

			if p.Stats.PhysicalDamageDealt != rp.Stats.PhysicalDamageDealt {
				t.Fail()
			}

			if p.Stats.PhysicalDamageDealtToChampions != rp.Stats.PhysicalDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.PhysicalDamageTaken != rp.Stats.PhysicalDamageTaken {
				t.Fail()
			}

			if p.Stats.QuadraKills != rp.Stats.QuadraKills {
				t.Fail()
			}

			if p.Stats.SightWardsBoughtInGame != rp.Stats.SightWardsBoughtInGame {
				t.Fail()
			}

			if p.Stats.TimeCCingOthers != rp.Stats.TimeCCingOthers {
				t.Fail()
			}

			if p.Stats.TotalDamageDealt != rp.Stats.TotalDamageDealt {
				t.Fail()
			}

			if p.Stats.TotalDamageDealtToChampions != rp.Stats.TotalDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.TotalDamageTaken != rp.Stats.TotalDamageTaken {
				t.Fail()
			}

			if p.Stats.TotalHeal != rp.Stats.TotalHeal {
				t.Fail()
			}

			if p.Stats.TotalMinionsKilled != rp.Stats.TotalMinionsKilled {
				t.Fail()
			}

			if p.Stats.TotalPlayerScore != rp.Stats.TotalPlayerScore {
				t.Fail()
			}

			if p.Stats.TotalScoreRank != rp.Stats.TotalScoreRank {
				t.Fail()
			}

			if p.Stats.TotalTimeCrowdControlDealt != rp.Stats.TotalTimeCrowdControlDealt {
				t.Fail()
			}

			if p.Stats.TotalUnitsHealed != rp.Stats.TotalUnitsHealed {
				t.Fail()
			}

			if p.Stats.TripleKills != rp.Stats.TripleKills {
				t.Fail()
			}

			if p.Stats.TrueDamageDealt != rp.Stats.TrueDamageDealt {
				t.Fail()
			}

			if p.Stats.TrueDamageDealtToChampions != rp.Stats.TrueDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.TrueDamageTaken != rp.Stats.TrueDamageTaken {
				t.Fail()
			}

			if p.Stats.TurretKills != rp.Stats.TurretKills {
				t.Fail()
			}

			if p.Stats.UnrealKills != rp.Stats.UnrealKills {
				t.Fail()
			}

			if p.Stats.VisionScore != rp.Stats.VisionScore {
				t.Fail()
			}

			if p.Stats.VisionWardsBoughtInGame != rp.Stats.VisionWardsBoughtInGame {
				t.Fail()
			}

			if p.Stats.WardsKilled != rp.Stats.WardsKilled {
				t.Fail()
			}

			if p.Stats.WardsPlaced != rp.Stats.WardsPlaced {
				t.Fail()
			}
		}
	}
}

// TODO: test Banned(), Picked(), and Won() both packed and unpacked
func TestMatchToProto(t *testing.T) {
	config.Setup()
	config.Config.KeepStats = true

	raws := rawSamples()

	for _, raw := range raws {
		before := ToMatch(raw)
		after := MakeMatch(before.Bytes())

		// Match fields
		if before.GameID != after.GameID {
			t.Fail()
		}

		if before.GameCreation != after.GameCreation {
			t.Fail()
		}

		if before.GameDuration != after.GameDuration {
			t.Fail()
		}

		if before.GameMode != after.GameMode {
			t.Fail()
		}

		if before.GameType != after.GameType {
			t.Fail()
		}

		if before.MapID != after.MapID {
			t.Fail()
		}

		if before.SeasonID != after.SeasonID {
			t.Fail()
		}

		// Participant fields
		for i, p := range before.Participants {
			rp := after.Participants[i]

			if p.Stats.Assists != rp.Stats.Assists {
				t.Fail()
			}

			if p.Stats.ChampLevel != rp.Stats.ChampLevel {
				t.Fail()
			}

			if p.Stats.CombatPlayerScore != rp.Stats.CombatPlayerScore {
				t.Fail()
			}

			if p.Stats.DamageDealtToObjectives != rp.Stats.DamageDealtToObjectives {
				t.Fail()
			}

			if p.Stats.DamageDealtToTurrets != rp.Stats.DamageDealtToTurrets {
				t.Fail()
			}

			if p.Stats.DamageSelfMitigated != rp.Stats.DamageSelfMitigated {
				t.Fail()
			}

			if p.Stats.Deaths != rp.Stats.Deaths {
				t.Fail()
			}

			if p.Stats.DoubleKills != rp.Stats.DoubleKills {
				t.Fail()
			}

			if p.Stats.FirstBloodAssist != rp.Stats.FirstBloodAssist {
				t.Fail()
			}

			if p.Stats.FirstBloodKill != rp.Stats.FirstBloodKill {
				t.Fail()
			}

			if p.Stats.FirstInhibitorAssist != rp.Stats.FirstInhibitorAssist {
				t.Fail()
			}

			if p.Stats.FirstInhibitorKill != rp.Stats.FirstInhibitorKill {
				t.Fail()
			}

			if p.Stats.FirstTowerAssist != rp.Stats.FirstTowerAssist {
				t.Fail()
			}

			if p.Stats.FirstTowerKill != rp.Stats.FirstTowerKill {
				t.Fail()
			}

			if p.Stats.GoldEarned != rp.Stats.GoldEarned {
				t.Fail()
			}

			if p.Stats.GoldSpent != rp.Stats.GoldSpent {
				t.Fail()
			}

			if p.Stats.InhibitorKills != rp.Stats.InhibitorKills {
				t.Fail()
			}

			if p.Stats.KillingSprees != rp.Stats.KillingSprees {
				t.Fail()
			}

			if p.Stats.Kills != rp.Stats.Kills {
				t.Fail()
			}

			if p.Stats.LargestCriticalStrike != rp.Stats.LargestCriticalStrike {
				t.Fail()
			}

			if p.Stats.LargestKillingSpree != rp.Stats.LargestKillingSpree {
				t.Fail()
			}

			if p.Stats.LargestMultiKill != rp.Stats.LargestMultiKill {
				t.Fail()
			}

			if p.Stats.LongestTimeSpentLiving != rp.Stats.LongestTimeSpentLiving {
				t.Fail()
			}

			if p.Stats.MagicalDamageTaken != rp.Stats.MagicalDamageTaken {
				t.Fail()
			}

			if p.Stats.MagicDamageDealt != rp.Stats.MagicDamageDealt {
				t.Fail()
			}

			if p.Stats.MagicDamageDealtToChampions != rp.Stats.MagicDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilled != rp.Stats.NeutralMinionsKilled {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilledEnemyJungle != rp.Stats.NeutralMinionsKilledEnemyJungle {
				t.Fail()
			}

			if p.Stats.NeutralMinionsKilledTeamJungle != rp.Stats.NeutralMinionsKilledTeamJungle {
				t.Fail()
			}

			if p.Stats.ObjectivePlayerScore != rp.Stats.ObjectivePlayerScore {
				t.Fail()
			}

			if p.Stats.PentaKills != rp.Stats.PentaKills {
				t.Fail()
			}

			if p.Stats.PhysicalDamageDealt != rp.Stats.PhysicalDamageDealt {
				t.Fail()
			}

			if p.Stats.PhysicalDamageDealtToChampions != rp.Stats.PhysicalDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.PhysicalDamageTaken != rp.Stats.PhysicalDamageTaken {
				t.Fail()
			}

			if p.Stats.QuadraKills != rp.Stats.QuadraKills {
				t.Fail()
			}

			if p.Stats.SightWardsBoughtInGame != rp.Stats.SightWardsBoughtInGame {
				t.Fail()
			}

			if p.Stats.TimeCCingOthers != rp.Stats.TimeCCingOthers {
				t.Fail()
			}

			if p.Stats.TotalDamageDealt != rp.Stats.TotalDamageDealt {
				t.Fail()
			}

			if p.Stats.TotalDamageDealtToChampions != rp.Stats.TotalDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.TotalDamageTaken != rp.Stats.TotalDamageTaken {
				t.Fail()
			}

			if p.Stats.TotalHeal != rp.Stats.TotalHeal {
				t.Fail()
			}

			if p.Stats.TotalMinionsKilled != rp.Stats.TotalMinionsKilled {
				t.Fail()
			}

			if p.Stats.TotalPlayerScore != rp.Stats.TotalPlayerScore {
				t.Fail()
			}

			if p.Stats.TotalScoreRank != rp.Stats.TotalScoreRank {
				t.Fail()
			}

			if p.Stats.TotalTimeCrowdControlDealt != rp.Stats.TotalTimeCrowdControlDealt {
				t.Fail()
			}

			if p.Stats.TotalUnitsHealed != rp.Stats.TotalUnitsHealed {
				t.Fail()
			}

			if p.Stats.TripleKills != rp.Stats.TripleKills {
				t.Fail()
			}

			if p.Stats.TrueDamageDealt != rp.Stats.TrueDamageDealt {
				t.Fail()
			}

			if p.Stats.TrueDamageDealtToChampions != rp.Stats.TrueDamageDealtToChampions {
				t.Fail()
			}

			if p.Stats.TrueDamageTaken != rp.Stats.TrueDamageTaken {
				t.Fail()
			}

			if p.Stats.TurretKills != rp.Stats.TurretKills {
				t.Fail()
			}

			if p.Stats.UnrealKills != rp.Stats.UnrealKills {
				t.Fail()
			}

			if p.Stats.VisionScore != rp.Stats.VisionScore {
				t.Fail()
			}

			if p.Stats.VisionWardsBoughtInGame != rp.Stats.VisionWardsBoughtInGame {
				t.Fail()
			}

			if p.Stats.WardsKilled != rp.Stats.WardsKilled {
				t.Fail()
			}

			if p.Stats.WardsPlaced != rp.Stats.WardsPlaced {
				t.Fail()
			}
		}
	}
}

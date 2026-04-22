package seed

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
)

func TestNormalizeCurrentSeasonSyncType(t *testing.T) {
	t.Run("defaults empty to all", func(t *testing.T) {
		got, err := normalizeCurrentSeasonSyncType("")
		if err != nil {
			t.Fatalf("normalizeCurrentSeasonSyncType returned error: %v", err)
		}
		if got != currentSeasonSyncAll {
			t.Fatalf("expected %q, got %q", currentSeasonSyncAll, got)
		}
	})

	t.Run("accepts known value", func(t *testing.T) {
		got, err := normalizeCurrentSeasonSyncType("stats")
		if err != nil {
			t.Fatalf("normalizeCurrentSeasonSyncType returned error: %v", err)
		}
		if got != currentSeasonSyncStats {
			t.Fatalf("expected %q, got %q", currentSeasonSyncStats, got)
		}
	})

	t.Run("rejects unknown value", func(t *testing.T) {
		if _, err := normalizeCurrentSeasonSyncType("unknown"); err == nil {
			t.Fatal("expected error for unknown sync_type")
		}
	})
}

func TestResolveCurrentSeasonFromJob(t *testing.T) {
	t.Run("uses explicit season first", func(t *testing.T) {
		job := &db.ETLJob{
			Scope: map[string]any{"season": 2026, "years": []int{2025}},
		}
		got := resolveCurrentSeasonFromJob(job)
		if got != 2026 {
			t.Fatalf("expected season=2026, got %d", got)
		}
	})

	t.Run("falls back to years scope", func(t *testing.T) {
		job := &db.ETLJob{
			Scope: map[string]any{"years": []any{2024.0}},
		}
		got := resolveCurrentSeasonFromJob(job)
		if got != 2024 {
			t.Fatalf("expected season=2024, got %d", got)
		}
	})

	t.Run("falls back to current year", func(t *testing.T) {
		job := &db.ETLJob{Scope: map[string]any{}}
		got := resolveCurrentSeasonFromJob(job)
		if got != time.Now().Year() {
			t.Fatalf("expected season=%d, got %d", time.Now().Year(), got)
		}
	})
}

func TestBuildCurrentSeasonBattingRow(t *testing.T) {
	split := core.MLBStatSplit{
		Season: "2026",
		Player: &core.MLBPlayerRef{ID: 1001},
		Team:   &core.MLBTeamRef{ID: 117},
		Stat: core.MLBStatGroup{
			MLBCommonStatGroup: core.MLBCommonStatGroup{
				GamesPlayed:      json.RawMessage(`12`),
				AtBats:           json.RawMessage(`"44"`),
				Hits:             json.RawMessage(`15`),
				HomeRuns:         json.RawMessage(`3`),
				PlateAppearances: json.RawMessage(`50`),
				Avg:              json.RawMessage(`".341"`),
				OPS:              json.RawMessage(`"1.001"`),
			},
		},
	}
	crosswalk := currentSeasonCrosswalk{
		playerByMLBID: map[int]string{1001: "judgeaa01"},
		teamByMLBID: map[int]currentSeasonTeamCrosswalk{
			117: {TeamID: "NYA"},
		},
	}

	row, ok, reason := buildCurrentSeasonBattingRow(split, 2026, crosswalk)
	if !ok {
		t.Fatalf("expected row to parse, got reason=%s", reason)
	}
	if row.PlayerID == nil || *row.PlayerID != "judgeaa01" {
		t.Fatalf("unexpected player_id: %#v", row.PlayerID)
	}
	if row.TeamID == nil || *row.TeamID != "NYA" {
		t.Fatalf("unexpected team_id: %#v", row.TeamID)
	}
	if row.Season != 2026 {
		t.Fatalf("expected season=2026, got %d", row.Season)
	}
	if row.G == nil || *row.G != 12 {
		t.Fatalf("unexpected games played: %#v", row.G)
	}
	if row.AVG == nil || *row.AVG != 0.341 {
		t.Fatalf("unexpected avg: %#v", row.AVG)
	}
}

func TestBuildCurrentSeasonStandingRow_NilStreakDoesNotPanic(t *testing.T) {
	record := core.MLBStandingsRecord{
		Season:   "2026",
		Division: &core.MLBDivisionRef{ID: 201, Name: "AL East"},
	}
	teamRecord := core.MLBTeamRecord{
		Team:              &core.MLBTeamRef{ID: 117},
		Wins:              10,
		Losses:            5,
		WinningPercentage: ".667",
		LastTenRecords: []core.MLBSplitRecord{
			{Type: "lastTen", Wins: 7, Losses: 3},
		},
	}
	crosswalk := currentSeasonCrosswalk{
		teamByMLBID: map[int]currentSeasonTeamCrosswalk{
			117: {TeamID: "NYA", FranchiseID: "NYY"},
		},
	}

	row, ok, reason := buildCurrentSeasonStandingRow(record, teamRecord, 2026, crosswalk)
	if !ok {
		t.Fatalf("expected row to parse, got reason=%s", reason)
	}
	if row.Streak != nil {
		t.Fatalf("expected nil streak when API streak missing, got %#v", row.Streak)
	}
	if row.L10 == nil || *row.L10 != "7-3" {
		t.Fatalf("unexpected l10: %#v", row.L10)
	}
}

func TestBuildCurrentSeasonGameRow_ScheduledGameLeavesScoresNil(t *testing.T) {
	dateBucket := core.MLBScheduleDate{Date: "2026-04-01"}
	game := core.MLBGame{
		GamePk:   123,
		Season:   "2026",
		Status:   &core.MLBGameStatus{DetailedState: "Scheduled"},
		DayNight: "night",
		Teams: &core.MLBGameTeams{
			Away: &core.MLBGameTeamWrapper{Team: &core.MLBTeamRef{ID: 117}, Score: 0},
			Home: &core.MLBGameTeamWrapper{Team: &core.MLBTeamRef{ID: 121}, Score: 0},
		},
	}
	crosswalk := currentSeasonCrosswalk{
		teamByMLBID: map[int]currentSeasonTeamCrosswalk{
			117: {TeamID: "NYA"},
			121: {TeamID: "NYN"},
		},
	}

	row, ok, reason := buildCurrentSeasonGameRow(dateBucket, game, 2026, crosswalk)
	if !ok {
		t.Fatalf("expected row to parse, got reason=%s", reason)
	}
	if row.AwayScore != nil || row.HomeScore != nil {
		t.Fatalf("expected nil scores for scheduled game, got away=%#v home=%#v", row.AwayScore, row.HomeScore)
	}
	if row.Venue != nil {
		t.Fatalf("expected nil venue when absent, got %#v", row.Venue)
	}
}

func TestClassifyCurrentSeasonSyncFailure_Network(t *testing.T) {
	class, retryable := classifyCurrentSeasonSyncFailure(&net.DNSError{IsTimeout: true})
	if class != "network" {
		t.Fatalf("expected class=network, got %s", class)
	}
	if !retryable {
		t.Fatal("expected retryable network failure")
	}
}

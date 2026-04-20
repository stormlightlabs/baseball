package seed

import (
	"testing"
	"time"

	"stormlightlabs.org/baseball/internal/db"
)

func TestRetrosheetMaterializedViewsContainCoreGameLogViews(t *testing.T) {
	required := []string{
		"player_game_batting_stats",
		"player_game_pitching_stats",
		"player_game_fielding_stats",
		"team_game_stats",
		"season_batting_leaders",
		"season_pitching_leaders",
		"career_batting_leaders",
		"career_pitching_leaders",
	}

	for _, view := range required {
		if !containsString(retrosheetMaterializedViews, view) {
			t.Fatalf("expected retrosheetMaterializedViews to include %q", view)
		}
	}
}

func TestMaterializedViewRefreshSetsHaveNoDuplicates(t *testing.T) {
	assertNoDuplicateStrings(t, retrosheetMaterializedViews, "retrosheetMaterializedViews")
	assertNoDuplicateStrings(t, supplementalPipelineMaterializedViews, "supplementalPipelineMaterializedViews")
}

func TestMaterializedViewRefreshSetsDoNotOverlap(t *testing.T) {
	supplemental := make(map[string]struct{}, len(supplementalPipelineMaterializedViews))
	for _, view := range supplementalPipelineMaterializedViews {
		supplemental[view] = struct{}{}
	}

	for _, view := range retrosheetMaterializedViews {
		if _, ok := supplemental[view]; ok {
			t.Fatalf("view %q appears in both refresh sets", view)
		}
	}
}

func assertNoDuplicateStrings(t *testing.T, values []string, label string) {
	t.Helper()

	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			t.Fatalf("%s contains duplicate value %q", label, value)
		}
		seen[value] = struct{}{}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestMaterializedViewAttemptSummary(t *testing.T) {
	summary := materializedViewAttemptSummary([]db.MaterializedViewRefreshAttempt{
		{
			Attempt:  1,
			Status:   "completed",
			Duration: 45 * time.Second,
		},
		{
			Attempt:  2,
			Status:   "completed",
			Duration: 12 * time.Second,
		},
		{
			Attempt: 1,
			Status:  "deferred_dependency",
		},
		{
			Attempt: 1,
			Status:  "failed",
		},
	}, 30*time.Second)

	if summary["attempts"] != 4 {
		t.Fatalf("expected attempts=4, got %d", summary["attempts"])
	}
	if summary["retries"] != 1 {
		t.Fatalf("expected retries=1, got %d", summary["retries"])
	}
	if summary["deferred"] != 1 {
		t.Fatalf("expected deferred=1, got %d", summary["deferred"])
	}
	if summary["failed"] != 1 {
		t.Fatalf("expected failed=1, got %d", summary["failed"])
	}
	if summary["slow_count"] != 1 {
		t.Fatalf("expected slow_count=1, got %d", summary["slow_count"])
	}
}

func TestMaterializedViewPhaseName(t *testing.T) {
	if got := materializedViewPhaseName("crosswalk/metadata"); got != "materialized_views.crosswalk_metadata" {
		t.Fatalf("unexpected phase name: %s", got)
	}
}

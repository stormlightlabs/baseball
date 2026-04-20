package seed

import "testing"

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

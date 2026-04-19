package db

import "testing"

func TestLoadNegroLeaguesTeamMappingFromEmbeddedCSV(t *testing.T) {
	mapping, err := loadNegroLeaguesTeamMapping()
	if err != nil {
		t.Fatalf("loadNegroLeaguesTeamMapping returned error: %v", err)
	}
	if len(mapping) == 0 {
		t.Fatal("expected non-empty Negro Leagues team mapping")
	}

	tests := map[string]string{
		"HOM": "NN2",
		"CAG": "NNL",
		"BIR": "NAL",
	}
	for teamID, wantLeague := range tests {
		gotLeague, ok := mapping[teamID]
		if !ok {
			t.Fatalf("missing mapping for %s", teamID)
		}
		if gotLeague != wantLeague {
			t.Fatalf("unexpected league for %s: got %q want %q", teamID, gotLeague, wantLeague)
		}
	}
}

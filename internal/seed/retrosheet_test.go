package seed

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestCleanupRetrosheetArtifactsDryRun(t *testing.T) {
	root := t.TempDir()

	keep := []string{
		"gameinfo.csv",
		"allplayers.csv",
		"gamelogs/GL2025.zip",
		"plays/2025plays.zip",
		"allstar/allstar.zip",
		"negroleagues/gameinfo.csv",
		"negroleagues/plays.csv",
	}
	remove := []string{
		"gamelogs/GL2025.TXT",
		"plays/2025.EVN",
		"ejections/ejections.csv",
		"scratch.tmp",
		"negroleagues/readme.txt",
	}

	for _, rel := range keep {
		mustWriteFile(t, filepath.Join(root, rel), "keep")
	}
	for _, rel := range remove {
		mustWriteFile(t, filepath.Join(root, rel), "remove")
	}

	result, err := CleanupRetrosheetArtifacts(root, true)
	if err != nil {
		t.Fatalf("CleanupRetrosheetArtifacts returned error: %v", err)
	}
	if len(result.Removed) != 0 {
		t.Fatalf("expected no removed files in dry-run, got %v", result.Removed)
	}

	wantCandidates := make([]string, 0, len(remove))
	for _, rel := range remove {
		wantCandidates = append(wantCandidates, filepath.Join(root, rel))
	}
	slices.Sort(wantCandidates)
	if !reflect.DeepEqual(wantCandidates, result.Candidates) {
		t.Fatalf("unexpected candidates\nwant: %v\ngot:  %v", wantCandidates, result.Candidates)
	}

	for _, rel := range append(keep, remove...) {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to exist after dry-run: %v", rel, err)
		}
	}
}

func TestCleanupRetrosheetArtifactsRemovesTransientFiles(t *testing.T) {
	root := t.TempDir()

	keep := []string{
		"gamelogs/GL2025.zip",
		"plays/2025plays.zip",
		"gameinfo.csv",
	}
	remove := []string{
		"plays/2025.EVA",
		"scratch.csv",
	}

	for _, rel := range keep {
		mustWriteFile(t, filepath.Join(root, rel), "keep")
	}
	for _, rel := range remove {
		mustWriteFile(t, filepath.Join(root, rel), "remove")
	}

	result, err := CleanupRetrosheetArtifacts(root, false)
	if err != nil {
		t.Fatalf("CleanupRetrosheetArtifacts returned error: %v", err)
	}
	if len(result.Removed) != len(remove) {
		t.Fatalf("expected %d removed files, got %d", len(remove), len(result.Removed))
	}

	for _, rel := range remove {
		if _, err := os.Stat(filepath.Join(root, rel)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, stat err=%v", rel, err)
		}
	}
	for _, rel := range keep {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected keep file %s to remain: %v", rel, err)
		}
	}
}

func TestCleanupRetrosheetArtifactsMissingDir(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "retrosheet")

	result, err := CleanupRetrosheetArtifacts(missing, false)
	if err != nil {
		t.Fatalf("expected no error for missing directory, got %v", err)
	}
	if len(result.Candidates) != 0 || len(result.Removed) != 0 {
		t.Fatalf("expected empty result for missing directory, got %#v", result)
	}
}

package seed

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestShouldAutoCloneDataRoot(t *testing.T) {
	tests := []struct {
		name string
		root string
		want bool
	}{
		{name: "relative default", root: "tools/data", want: true},
		{name: "absolute default", root: "/home/app/tools/data", want: true},
		{name: "blank root", root: "", want: true},
		{name: "legacy data root", root: "data", want: false},
		{name: "custom root", root: "/mnt/baseball-data", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := shouldAutoCloneDataRoot(tt.root)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDataAutoCloneEnabled(t *testing.T) {
	t.Setenv(DataAutoCloneEnvVar, "")
	if !dataAutoCloneEnabled() {
		t.Fatal("expected auto clone enabled by default")
	}

	t.Setenv(DataAutoCloneEnvVar, "false")
	if dataAutoCloneEnabled() {
		t.Fatal("expected auto clone disabled for false")
	}

	t.Setenv(DataAutoCloneEnvVar, "0")
	if dataAutoCloneEnabled() {
		t.Fatal("expected auto clone disabled for 0")
	}

	t.Setenv(DataAutoCloneEnvVar, "true")
	if !dataAutoCloneEnabled() {
		t.Fatal("expected auto clone enabled for true")
	}
}

func TestMissingPipelineDataArtifactsCompleteRoot(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "People.csv"), "id,name\n")
	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "Teams.csv"), "year,team\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "woba.csv"), "season,woba\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "pf", "2025.csv"), "season,team,pf\n")
	mustWriteFile(t, filepath.Join(root, "salaries", "summary.csv"), "Year,Total,Average,Median\n")
	mustWriteFile(t, filepath.Join(root, "retrosheet", "allplayers.zip"), "zip")
	mustWriteFile(t, filepath.Join(root, "retrosheet", "biodata.zip"), "zip")

	opts := PipelineOptions{
		DataRoot:          root,
		LahmanCSVDir:      filepath.Join(root, "lahman", "csv"),
		RetrosheetDataDir: filepath.Join(root, "retrosheet"),
		FanGraphsDir:      filepath.Join(root, "fangraphs"),
		SalaryDataDir:     filepath.Join(root, "salaries"),
	}

	missing, err := missingPipelineDataArtifacts(opts)
	if err != nil {
		t.Fatalf("missingPipelineDataArtifacts returned error: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected no missing files, got %v", missing)
	}
}

func TestMissingPipelineDataArtifactsReportsRequiredPaths(t *testing.T) {
	root := t.TempDir()

	opts := PipelineOptions{
		DataRoot:          root,
		LahmanCSVDir:      filepath.Join(root, "lahman", "csv"),
		RetrosheetDataDir: filepath.Join(root, "retrosheet"),
		FanGraphsDir:      filepath.Join(root, "fangraphs"),
		SalaryDataDir:     filepath.Join(root, "salaries"),
	}

	missing, err := missingPipelineDataArtifacts(opts)
	if err != nil {
		t.Fatalf("missingPipelineDataArtifacts returned error: %v", err)
	}

	want := []string{
		filepath.Join(root, "fangraphs", "pf", "*.csv"),
		filepath.Join(root, "fangraphs", "woba.csv"),
		filepath.Join(root, "lahman", "csv", "People.csv"),
		filepath.Join(root, "lahman", "csv", "Teams.csv"),
		filepath.Join(root, "retrosheet", "allplayers.zip"),
		filepath.Join(root, "retrosheet", "biodata.zip"),
		filepath.Join(root, "salaries", "summary.csv"),
	}

	slices.Sort(want)
	if !reflect.DeepEqual(want, missing) {
		t.Fatalf("unexpected missing files\nwant: %v\ngot:  %v", want, missing)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

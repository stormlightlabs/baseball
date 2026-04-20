package seed

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestResolveDataRootPrecedence(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, filepath.Join("env", "root"))
	if err := os.MkdirAll(DefaultDataRoot, 0755); err != nil {
		t.Fatalf("mkdir default root: %v", err)
	}
	if err := os.MkdirAll(LegacyDataRoot, 0755); err != nil {
		t.Fatalf("mkdir legacy root: %v", err)
	}

	got := ResolveDataRoot(filepath.Join("flag", "root"))
	want := filepath.Join("flag", "root")
	if got != want {
		t.Fatalf("expected explicit root %q, got %q", want, got)
	}
}

func TestResolveDataRootUsesEnvBeforeLocalDirs(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, filepath.Join("external", "snapshot"))
	if err := os.MkdirAll(DefaultDataRoot, 0755); err != nil {
		t.Fatalf("mkdir default root: %v", err)
	}
	if err := os.MkdirAll(LegacyDataRoot, 0755); err != nil {
		t.Fatalf("mkdir legacy root: %v", err)
	}

	got := ResolveDataRoot("")
	want := filepath.Join("external", "snapshot")
	if got != want {
		t.Fatalf("expected env root %q, got %q", want, got)
	}
}

func TestResolveDataRootPrefersDefaultDataRootWhenPresent(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, "")
	if err := os.MkdirAll(DefaultDataRoot, 0755); err != nil {
		t.Fatalf("mkdir default root: %v", err)
	}
	if err := os.MkdirAll(LegacyDataRoot, 0755); err != nil {
		t.Fatalf("mkdir legacy root: %v", err)
	}

	got := ResolveDataRoot("")
	if got != DefaultDataRoot {
		t.Fatalf("expected %q, got %q", DefaultDataRoot, got)
	}
}

func TestResolveDataRootFallsBackToLegacyData(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, "")
	if err := os.MkdirAll(LegacyDataRoot, 0755); err != nil {
		t.Fatalf("mkdir legacy root: %v", err)
	}

	got := ResolveDataRoot("")
	if got != LegacyDataRoot {
		t.Fatalf("expected %q, got %q", LegacyDataRoot, got)
	}
}

func TestResolveDataRootDefaultsToDefaultDataRootWhenNoPathsExist(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, "")

	got := ResolveDataRoot("")
	if got != DefaultDataRoot {
		t.Fatalf("expected %q, got %q", DefaultDataRoot, got)
	}
}

func TestEnsurePipelineDataRootCreatesWorkerDirs(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "People.csv"), "id,name\n")
	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "Teams.csv"), "year,team\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "woba.csv"), "season,woba\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "pf", "2025.csv"), "season,team,pf\n")
	mustWriteFile(t, filepath.Join(root, "salaries", "summary.csv"), "Year,Total,Average,Median\n")

	opts := PipelineOptions{
		DataRoot:          root,
		LahmanCSVDir:      filepath.Join(root, "lahman", "csv"),
		RetrosheetDataDir: filepath.Join(root, "retrosheet"),
		ChadwickDataDir:   filepath.Join(root, "chadwick"),
		FanGraphsDir:      filepath.Join(root, "fangraphs"),
		SalaryDataDir:     filepath.Join(root, "salaries"),
	}

	provision, err := ensurePipelineDataRoot(context.Background(), opts)
	if err != nil {
		t.Fatalf("ensurePipelineDataRoot returned error: %v", err)
	}
	if provision.rootPath != root {
		t.Fatalf("expected rootPath=%q, got %q", root, provision.rootPath)
	}

	for _, dir := range []string{opts.RetrosheetDataDir, opts.ChadwickDataDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %q to be created: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", dir)
		}
	}
}

func TestMissingPipelineDataArtifactsCompleteRoot(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "People.csv"), "id,name\n")
	mustWriteFile(t, filepath.Join(root, "lahman", "csv", "Teams.csv"), "year,team\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "woba.csv"), "season,woba\n")
	mustWriteFile(t, filepath.Join(root, "fangraphs", "pf", "2025.csv"), "season,team,pf\n")
	mustWriteFile(t, filepath.Join(root, "salaries", "summary.csv"), "Year,Total,Average,Median\n")

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

func withTempWorkingDir(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

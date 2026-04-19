package seed

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolvePipelineYearsDevDefault(t *testing.T) {
	years, err := resolvePipelineYears(PipelineProfileDev, nil, nil)
	if err != nil {
		t.Fatalf("resolvePipelineYears returned error: %v", err)
	}

	for _, mustContain := range []int{1914, 1935, 1968, 2001, 2017} {
		if !containsInt(years, mustContain) {
			t.Fatalf("expected dev years to include %d, got %v", mustContain, years)
		}
	}

	expectedEnd := time.Now().Year() - 1
	if !containsInt(years, expectedEnd) {
		t.Fatalf("expected dev years to include %d, got %v", expectedEnd, years)
	}
}

func TestResolvePipelineYearsProdDefault(t *testing.T) {
	years, err := resolvePipelineYears(PipelineProfileProd, nil, nil)
	if err != nil {
		t.Fatalf("resolvePipelineYears returned error: %v", err)
	}

	if len(years) == 0 {
		t.Fatal("expected non-empty years")
	}
	if years[0] != 1910 {
		t.Fatalf("expected prod years to start at 1910, got %d", years[0])
	}
	expectedEnd := time.Now().Year() - 1
	if years[len(years)-1] != expectedEnd {
		t.Fatalf("expected prod years to end at %d, got %d", expectedEnd, years[len(years)-1])
	}
}

func TestResolvePipelineYearsWithEraExpansion(t *testing.T) {
	years, err := resolvePipelineYears(PipelineProfileDev, []int{2024}, []string{"fed"})
	if err != nil {
		t.Fatalf("resolvePipelineYears returned error: %v", err)
	}

	for _, y := range []int{1914, 1915, 2024} {
		if !containsInt(years, y) {
			t.Fatalf("expected years to include %d, got %v", y, years)
		}
	}
}

func TestNormalizePipelineOptionsRejectsUnknownEra(t *testing.T) {
	_, err := NormalizePipelineOptions(PipelineOptions{
		Profile:  PipelineProfileDev,
		Mode:     PipelineModeIncremental,
		EraNames: []string{"nope"},
	})
	if err == nil {
		t.Fatal("expected error for unknown era")
	}
}

func TestNormalizePipelineOptionsUsesResolvedDataRoot(t *testing.T) {
	withTempWorkingDir(t)
	t.Setenv(DataRootEnvVar, "")
	if err := os.MkdirAll(DefaultDataRoot, 0755); err != nil {
		t.Fatalf("mkdir default data root: %v", err)
	}

	opts, err := NormalizePipelineOptions(PipelineOptions{
		Profile: PipelineProfileDev,
		Mode:    PipelineModeIncremental,
		Years:   []int{2024},
	})
	if err != nil {
		t.Fatalf("NormalizePipelineOptions returned error: %v", err)
	}

	if opts.DataRoot != DefaultDataRoot {
		t.Fatalf("expected data root %q, got %q", DefaultDataRoot, opts.DataRoot)
	}
	if opts.LahmanCSVDir != filepath.Join(DefaultDataRoot, "lahman", "csv") {
		t.Fatalf("unexpected Lahman CSV dir: %s", opts.LahmanCSVDir)
	}
	if opts.RetrosheetDataDir != filepath.Join(DefaultDataRoot, "retrosheet") {
		t.Fatalf("unexpected Retrosheet dir: %s", opts.RetrosheetDataDir)
	}
	if opts.FanGraphsDir != filepath.Join(DefaultDataRoot, "fangraphs") {
		t.Fatalf("unexpected FanGraphs dir: %s", opts.FanGraphsDir)
	}
	if opts.ChadwickDataDir != filepath.Join(DefaultDataRoot, "chadwick") {
		t.Fatalf("unexpected Chadwick dir: %s", opts.ChadwickDataDir)
	}
	if opts.SalaryDataDir != filepath.Join(DefaultDataRoot, "salaries") {
		t.Fatalf("unexpected salaries dir: %s", opts.SalaryDataDir)
	}
}

func TestNormalizePipelineOptionsHonorsExplicitDataRoot(t *testing.T) {
	withTempWorkingDir(t)
	t.Setenv(DataRootEnvVar, filepath.Join("ignored", "env"))

	opts, err := NormalizePipelineOptions(PipelineOptions{
		Profile:  PipelineProfileDev,
		Mode:     PipelineModeIncremental,
		Years:    []int{2024},
		DataRoot: filepath.Join("custom", "root"),
	})
	if err != nil {
		t.Fatalf("NormalizePipelineOptions returned error: %v", err)
	}

	wantRoot := filepath.Join("custom", "root")
	if opts.DataRoot != wantRoot {
		t.Fatalf("expected data root %q, got %q", wantRoot, opts.DataRoot)
	}
	if opts.LahmanCSVDir != filepath.Join(wantRoot, "lahman", "csv") {
		t.Fatalf("unexpected Lahman CSV dir: %s", opts.LahmanCSVDir)
	}
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

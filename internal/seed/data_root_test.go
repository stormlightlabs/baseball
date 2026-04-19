package seed

import (
	"os"
	"path/filepath"
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

func TestResolveDataRootPrefersVendorDataWhenPresent(t *testing.T) {
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

func TestResolveDataRootDefaultsToVendorDataWhenNoPathsExist(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv(DataRootEnvVar, "")

	got := ResolveDataRoot("")
	if got != DefaultDataRoot {
		t.Fatalf("expected %q, got %q", DefaultDataRoot, got)
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

package seed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFormatByteSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "bytes",
			size: 999,
			want: "999 B",
		},
		{
			name: "one kilobyte",
			size: 1024,
			want: "1.0 KB",
		},
		{
			name: "fractional kilobytes",
			size: 1536,
			want: "1.5 KB",
		},
		{
			name: "one megabyte",
			size: 1024 * 1024,
			want: "1.0 MB",
		},
		{
			name: "one gigabyte",
			size: 1024 * 1024 * 1024,
			want: "1.0 GB",
		},
		{
			name: "one terabyte",
			size: 1024 * 1024 * 1024 * 1024,
			want: "1.0 TB",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := formatByteSize(tt.size)
			if got != tt.want {
				t.Fatalf("formatByteSize(%d): want %q, got %q", tt.size, tt.want, got)
			}
		})
	}
}

func TestWriteChadwickManifest(t *testing.T) {
	root := t.TempDir()
	shardA := filepath.Join(root, "people-a.csv")
	shardB := filepath.Join(root, "people-b.csv")
	merged := filepath.Join(root, "people.csv")

	if err := os.WriteFile(shardA, []byte("key_mlbam,key_retro,key_bbref,name_first,name_last\n1,retro,bbref,First,Last\n"), 0644); err != nil {
		t.Fatalf("write shardA: %v", err)
	}
	if err := os.WriteFile(shardB, []byte("key_mlbam,key_retro,key_bbref,name_first,name_last\n2,retro2,bbref2,Second,Last\n"), 0644); err != nil {
		t.Fatalf("write shardB: %v", err)
	}
	if err := os.WriteFile(merged, []byte("key_mlbam,key_retro,key_bbref,name_first,name_last\n1,retro,bbref,First,Last\n2,retro2,bbref2,Second,Last\n"), 0644); err != nil {
		t.Fatalf("write merged: %v", err)
	}

	if err := writeChadwickManifest(root, merged, []string{shardB, shardA}); err != nil {
		t.Fatalf("writeChadwickManifest returned error: %v", err)
	}

	manifestPath := filepath.Join(root, chadwickManifestFilename)
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var manifest chadwickManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if manifest.SchemaVersion != 1 {
		t.Fatalf("expected schema version 1, got %d", manifest.SchemaVersion)
	}
	if manifest.SourceRepoURL != chadwickRegisterRepoURL {
		t.Fatalf("unexpected source repo URL: %s", manifest.SourceRepoURL)
	}
	if manifest.SourceShardURL != chadwickRegisterShardURL {
		t.Fatalf("unexpected shard URL template: %s", manifest.SourceShardURL)
	}
	if len(manifest.Files) != 3 {
		t.Fatalf("expected 3 manifest files, got %d", len(manifest.Files))
	}
	for _, file := range manifest.Files {
		if file.Path == "" || file.SHA256 == "" || file.SizeBytes <= 0 {
			t.Fatalf("invalid manifest file entry: %#v", file)
		}
	}

	gotColumns := slices.Clone(manifest.RequiredColumns)
	wantColumns := slices.Clone(chadwickRequiredColumns)
	slices.Sort(gotColumns)
	slices.Sort(wantColumns)
	if !slices.Equal(gotColumns, wantColumns) {
		t.Fatalf("unexpected required columns: %v", manifest.RequiredColumns)
	}
}

func TestFetchChadwickRegisterData_UsesCommittedPeopleCSV(t *testing.T) {
	root := t.TempDir()
	peoplePath := filepath.Join(root, "people.csv")
	content := []byte("key_mlbam,key_retro,key_bbref,name_first,name_last\n1,retro,bbref,First,Last\n")
	if err := os.WriteFile(peoplePath, content, 0644); err != nil {
		t.Fatalf("write people.csv: %v", err)
	}

	if err := FetchChadwickRegisterData(context.Background(), root, false); err != nil {
		t.Fatalf("FetchChadwickRegisterData returned error: %v", err)
	}

	manifestPath := filepath.Join(root, chadwickManifestFilename)
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Fatalf("did not expect manifest to be created/rewritten in no-op mode")
	}

	for _, shard := range chadwickShardKeys {
		shardPath := filepath.Join(root, "people-"+shard+".csv")
		if _, err := os.Stat(shardPath); !os.IsNotExist(err) {
			t.Fatalf("unexpected shard file present: %s", shardPath)
		}
	}
}

package commands

import (
	"strings"
	"testing"
)

func TestParseRecreateConnectionURL(t *testing.T) {
	dbName, adminCfg, err := parseRecreateConnection("postgres://postgres:secret@localhost:5432/bigfly_db?sslmode=disable")
	if err != nil {
		t.Fatalf("parseRecreateConnection returned error: %v", err)
	}
	if dbName != "bigfly_db" {
		t.Fatalf("expected dbName bigfly_db, got %q", dbName)
	}

	if adminCfg.Database != "postgres" {
		t.Fatalf("expected admin database postgres, got %q", adminCfg.Database)
	}
}

func TestParseRecreateConnectionDSN(t *testing.T) {
	dbName, adminCfg, err := parseRecreateConnection("host=postgres port=5432 user=postgres password=secret dbname=bigfly_db sslmode=disable")
	if err != nil {
		t.Fatalf("parseRecreateConnection returned error: %v", err)
	}
	if dbName != "bigfly_db" {
		t.Fatalf("expected dbName bigfly_db, got %q", dbName)
	}

	if adminCfg.Database != "postgres" {
		t.Fatalf("expected admin database postgres, got %q", adminCfg.Database)
	}
	if adminCfg.Host != "postgres" {
		t.Fatalf("expected host postgres, got %q", adminCfg.Host)
	}
}

func TestParseRecreateConnectionRequiresDatabaseName(t *testing.T) {
	_, _, err := parseRecreateConnection("host=localhost user=postgres sslmode=disable")
	if err == nil {
		t.Fatalf("expected error when database name is missing")
	}
	if !strings.Contains(err.Error(), "must include a database name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

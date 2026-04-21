package cli

import "testing"

func TestNewBaseballRootCmdDoesNotExposeETL(t *testing.T) {
	root := NewBaseballRootCmd()

	for _, cmd := range root.Commands() {
		if cmd.Name() == "etl" {
			t.Fatalf("expected primary baseball root to exclude etl command")
		}
	}
}

func TestNewETLRootCmdUseAndSubcommands(t *testing.T) {
	root := NewETLRootCmd()

	if got := root.Use; got != "baseball-etl" {
		t.Fatalf("expected etl root use baseball-etl, got %q", got)
	}

	required := map[string]bool{
		"fetch":    false,
		"load":     false,
		"cleanup":  false,
		"run":      false,
		"worker":   false,
		"validate": false,
		"status":   false,
	}

	for _, cmd := range root.Commands() {
		if _, ok := required[cmd.Name()]; ok {
			required[cmd.Name()] = true
		}
	}

	for name, seen := range required {
		if !seen {
			t.Fatalf("expected etl root subcommand %q to be present", name)
		}
	}
}

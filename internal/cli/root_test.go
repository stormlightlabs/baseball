package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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

func TestDbCommandContractExcludesLegacyPopulateAndReset(t *testing.T) {
	dbCmd := DbCmd()
	usage := dbCmd.UsageString()

	required := []string{"migrate", "recreate", "refresh-views"}
	for _, name := range required {
		if !strings.Contains(usage, name) {
			t.Fatalf("expected db help to include %q", name)
		}
	}

	disallowed := []string{"reset", "populate", "repopulate"}
	for _, name := range disallowed {
		if strings.Contains(usage, name) {
			t.Fatalf("expected db help to exclude %q", name)
		}
	}
}

func TestCLIHelpContractsForCanonicalFlow(t *testing.T) {
	baseballRoot := NewBaseballRootCmd()
	etlRoot := NewETLRootCmd()

	baseballExpectations := map[string][]string{
		"db":     {"migrate", "recreate", "refresh-views"},
		"server": {"start", "fetch", "health"},
	}
	for cmdName, mustContain := range baseballExpectations {
		cmd := findCommand(t, baseballRoot, cmdName)
		usage := cmd.UsageString()
		for _, token := range mustContain {
			if !strings.Contains(usage, token) {
				t.Fatalf("expected %s --help to contain %q", cmdName, token)
			}
		}
	}

	etlExpectations := map[string][]string{
		"run":      {"--job-type", "--enqueue-only"},
		"worker":   {"--max-active-jobs", "--poll-interval"},
		"validate": {"--profile", "--years"},
		"status":   {"--strict"},
	}
	for cmdName, mustContain := range etlExpectations {
		cmd := findCommand(t, etlRoot, cmdName)
		usage := cmd.UsageString()
		for _, token := range mustContain {
			if !strings.Contains(usage, token) {
				t.Fatalf("expected baseball-etl %s --help to contain %q", cmdName, token)
			}
		}
	}
}

func findCommand(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()

	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}

	t.Fatalf("expected command %q on %q", name, root.Use)
	return nil
}

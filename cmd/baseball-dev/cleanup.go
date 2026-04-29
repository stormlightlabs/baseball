package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type cleanupConfig struct {
	artifactsDir string
	generatedDir string
	recreate     bool
}

func newCleanupCmd() *cobra.Command {
	cfg := cleanupConfig{
		artifactsDir: "test/artifacts",
		generatedDir: "test/generated",
		recreate:     true,
	}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove generated artifacts and generated Hurl tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.artifactsDir, "artifacts-dir", cfg.artifactsDir, "Artifacts directory to remove")
	cmd.Flags().StringVar(&cfg.generatedDir, "generated-dir", cfg.generatedDir, "Generated tests directory to remove")
	cmd.Flags().BoolVar(&cfg.recreate, "recreate", cfg.recreate, "Recreate directories after cleanup")
	return cmd
}

func runCleanup(cfg cleanupConfig) error {
	for _, dir := range []string{cfg.artifactsDir, cfg.generatedDir} {
		if err := validateCleanupPath(dir); err != nil {
			return err
		}
	}

	targets := []string{cfg.artifactsDir, cfg.generatedDir}
	for _, dir := range targets {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
		fmt.Printf("removed %s\n", dir)
	}

	if cfg.recreate {
		for _, dir := range targets {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to recreate %s: %w", dir, err)
			}
			fmt.Printf("recreated %s\n", dir)
		}
	}

	return nil
}

func validateCleanupPath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return fmt.Errorf("refusing unsafe cleanup path: %q", path)
	}
	if !strings.HasPrefix(trimmed, "test/") && trimmed != "test" {
		return fmt.Errorf("cleanup path must be under test/: %q", path)
	}
	return nil
}

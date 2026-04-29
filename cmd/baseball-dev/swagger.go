package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type swaggerConfig struct {
	useInstalled bool
	swagVersion  string
}

func newSwaggerConf(u bool, v string) swaggerConfig {
	return swaggerConfig{useInstalled: u, swagVersion: v}
}

func newSwaggerCmd() *cobra.Command {
	cfg := newSwaggerConf(false, "v1.16.6")

	cmd := &cobra.Command{
		Use:     "swagger",
		Aliases: []string{"swag"},
		Short:   "Manage Swagger docs generation and formatting",
	}
	cmd.PersistentFlags().BoolVar(&cfg.useInstalled, "use-installed", cfg.useInstalled, "Use installed `swag` binary instead of `go run github.com/swaggo/swag/cmd/swag@...`")
	cmd.PersistentFlags().StringVar(&cfg.swagVersion, "swag-version", cfg.swagVersion, "Swag module version used when not using installed binary")

	cmd.AddCommand(newSwaggerGenerateCmd(&cfg))
	cmd.AddCommand(newSwaggerFixCmd())
	cmd.AddCommand(newSwaggerFmtCmd(&cfg))
	cmd.AddCommand(newSwaggerCleanCmd())
	return cmd
}

func newSwaggerGenerateCmd(cfg *swaggerConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate Swagger/OpenAPI docs into internal/docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==> Generating Swagger docs")
			if err := runSwagCommand(cfg, "init",
				"--dir", "./internal/api",
				"--generalInfo", "server.go",
				"--output", "./internal/docs",
				"--parseDependency",
				"--parseInternal",
				"--md", "./internal/api",
			); err != nil {
				return err
			}
			return runSwaggerFix()
		},
	}
}

func newSwaggerFixCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fix",
		Short: "Remove LeftDelim/RightDelim from generated Swagger docs.go",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwaggerFix()
		},
	}
}

func newSwaggerFmtCmd(cfg *swaggerConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "fmt",
		Short: "Format Swagger comments",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==> Formatting Swagger comments")
			return runSwagCommand(cfg, "fmt")
		},
	}
}

func newSwaggerCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove generated Swagger files",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets := []string{
				"internal/docs/docs.go",
				"internal/docs/swagger.json",
				"internal/docs/swagger.yaml",
			}
			for _, file := range targets {
				if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed removing %s: %w", file, err)
				}
				fmt.Printf("removed %s\n", file)
			}
			return nil
		},
	}
}

func runSwagCommand(cfg *swaggerConfig, args ...string) error {
	if cfg.useInstalled {
		if _, err := exec.LookPath("swag"); err != nil {
			return fmt.Errorf("missing required command: swag")
		}
		return runCmd("swag", args...)
	}

	module := "github.com/swaggo/swag/cmd/swag@" + strings.TrimSpace(cfg.swagVersion)
	runArgs := append([]string{"run", module}, args...)
	return runCmd("go", runArgs...)
}

func runSwaggerFix() error {
	path := "internal/docs/docs.go"
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		fmt.Println("⚠ internal/docs/docs.go not found, skipping fix")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	filtered := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		if strings.Contains(line, "LeftDelim:") || strings.Contains(line, "RightDelim:") {
			changed = true
			continue
		}
		filtered = append(filtered, line)
	}
	if !changed {
		fmt.Println("✓ Swagger docs already clean")
		return nil
	}

	out := strings.Join(filtered, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Println("✓ Fixed swagger docs (removed LeftDelim/RightDelim fields)")
	return nil
}

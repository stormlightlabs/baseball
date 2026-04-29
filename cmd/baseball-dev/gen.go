package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var hurlVarRe = regexp.MustCompile(`\{\{\s*([^}]+?)\s*\}\}`)

type genConf struct {
	inputPath           string
	cleanupTransitional bool
	artifactsDir        string
	generatedDir        string
	internalSwaggerYAML string
	internalSwaggerJSON string
	queryParamsMode     string
	pathParamsMode      string
}

func newGenerateCmd() *cobra.Command {
	cfg := genConf{
		artifactsDir:        "test/artifacts",
		generatedDir:        "test/generated",
		internalSwaggerYAML: "internal/docs/swagger.yaml",
		internalSwaggerJSON: "internal/docs/swagger.json",
		queryParamsMode:     "required",
		pathParamsMode:      "variables",
	}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Build test artifacts from Swagger and split Hurl files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if envInput := strings.TrimSpace(os.Getenv("SWAGGER_INPUT")); cfg.inputPath == "" && envInput != "" {
				cfg.inputPath = envInput
			}
			if cleanupEnv := strings.TrimSpace(os.Getenv("CLEANUP_TRANSITIONAL")); cleanupEnv != "" {
				v, err := parseBoolLike(cleanupEnv)
				if err != nil {
					return err
				}
				cfg.cleanupTransitional = v
			}
			return runGenerate(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.inputPath, "input", "", "Path to Swagger input (.yaml/.yml/.json)")
	cmd.Flags().BoolVar(&cfg.cleanupTransitional, "cleanup", false, "Remove transitional files after successful generation")
	cmd.Flags().StringVar(&cfg.artifactsDir, "artifacts-dir", cfg.artifactsDir, "Directory for OpenAPI and aggregate outputs")
	cmd.Flags().StringVar(&cfg.generatedDir, "generated-dir", cfg.generatedDir, "Directory for split generated Hurl files")
	cmd.Flags().StringVar(&cfg.queryParamsMode, "query-params", cfg.queryParamsMode, "openapi-to-hurl query param mode: none|required|all")
	cmd.Flags().StringVar(&cfg.pathParamsMode, "path-params", cfg.pathParamsMode, "openapi-to-hurl path param mode: default|variables")
	return cmd
}

func runGenerate(cfg genConf) error {
	if !isOneOf(cfg.queryParamsMode, "none", "required", "all") {
		return fmt.Errorf("invalid --query-params value %q (allowed: none|required|all)", cfg.queryParamsMode)
	}
	if !isOneOf(cfg.pathParamsMode, "default", "variables") {
		return fmt.Errorf("invalid --path-params value %q (allowed: default|variables)", cfg.pathParamsMode)
	}

	if cfg.inputPath == "" {
		switch {
		case fileExists(cfg.internalSwaggerYAML):
			cfg.inputPath = cfg.internalSwaggerYAML
		case fileExists(cfg.internalSwaggerJSON):
			cfg.inputPath = cfg.internalSwaggerJSON
		default:
			return fmt.Errorf("no Swagger input found; expected %s or %s", cfg.internalSwaggerYAML, cfg.internalSwaggerJSON)
		}
	}

	if !fileExists(cfg.inputPath) {
		return fmt.Errorf("swagger input does not exist: %s", cfg.inputPath)
	}

	if !hasAllowedExt(cfg.inputPath, ".yaml", ".yml", ".json") {
		return fmt.Errorf("unsupported swagger input extension: %s", cfg.inputPath)
	}

	for _, cmdName := range []string{"npx", "openapi-to-hurl"} {
		if _, err := exec.LookPath(cmdName); err != nil {
			return fmt.Errorf("missing required command: %s", cmdName)
		}
	}

	if err := os.MkdirAll(cfg.artifactsDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.generatedDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll("test/smoke", 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll("test/regression", 0o755); err != nil {
		return err
	}

	openapi30 := filepath.Join(cfg.artifactsDir, "openapi.3.0.yaml")
	openapi31 := filepath.Join(cfg.artifactsDir, "openapi.3.1.yaml")
	openapi31San := filepath.Join(cfg.artifactsDir, "openapi.3.1.sanitized.yaml")
	allGenerated := filepath.Join(cfg.artifactsDir, "all.generated.hurl")

	fmt.Println("==> Converting Swagger 2.0 to OpenAPI 3.0")
	if err := runCmd("npx", "-y", "swagger2openapi", "--yaml", "--outfile", openapi30, cfg.inputPath); err != nil {
		return err
	}

	fmt.Println("==> Upgrading OpenAPI 3.0 to OpenAPI 3.1")
	if err := runCmd("npx", "-y", "openapi-format", openapi30, "--convertTo", "3.1", "--output", openapi31); err != nil {
		return err
	}

	fmt.Println("==> Sanitizing OpenAPI for openapi-to-hurl compatibility")
	if err := sanitizeOpenAPI(openapi31, openapi31San); err != nil {
		return err
	}

	fmt.Println("==> Generating Hurl")
	if err := writeCmdOutput(
		allGenerated,
		"openapi-to-hurl",
		"--query-params", cfg.queryParamsMode,
		"--path-params", cfg.pathParamsMode,
		openapi31San,
	); err != nil {
		return err
	}

	fmt.Println("==> Splitting Hurl")
	if err := clearGeneratedHurl(cfg.generatedDir); err != nil {
		return err
	}
	count, err := splitHurlFile(allGenerated, cfg.generatedDir)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d hurl files to %s\n", count, cfg.generatedDir)

	if cfg.cleanupTransitional {
		fmt.Println("==> Cleaning up transitional files")
		if err := removeIfExists(openapi30, openapi31San, allGenerated); err != nil {
			return err
		}
	}

	fmt.Println("==> Done")
	fmt.Printf("Swagger input: %s\n", cfg.inputPath)
	fmt.Printf("OpenAPI 3.1:   %s\n", openapi31)
	if cfg.cleanupTransitional {
		fmt.Println("Sanitized:     removed")
		fmt.Println("Generated:     removed")
	} else {
		fmt.Printf("Sanitized:     %s\n", openapi31San)
		fmt.Printf("Generated:     %s\n", allGenerated)
	}
	fmt.Printf("Split files:   %s\n", cfg.generatedDir)
	return nil
}

func removeIfExists(paths ...string) error {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func hasAllowedExt(path string, allowed ...string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(allowed, ext)
}

func slugifyURL(value string) string {
	out := strings.TrimSpace(value)
	out = strings.TrimPrefix(out, "{{host}}")
	out = strings.TrimPrefix(out, "{{ host }}")
	out = strings.TrimPrefix(out, "{{BASE_URL}}")
	out = strings.TrimPrefix(out, "{{ BASE_URL }}")
	out = strings.TrimPrefix(out, "http://")
	out = strings.TrimPrefix(out, "https://")
	if idx := strings.Index(out, "/"); idx >= 0 {
		out = out[idx:]
	}
	if out == "" {
		out = "/"
	}

	pathPart, queryPart, _ := strings.Cut(out, "?")
	pathPart = hurlVarRe.ReplaceAllString(pathPart, "$1")
	pathPart = strings.ReplaceAll(pathPart, "{", "")
	pathPart = strings.ReplaceAll(pathPart, "}", "")

	slug := pathPart
	if queryPart != "" {
		queryKeys := make([]string, 0)
		for _, pair := range strings.Split(queryPart, "&") {
			key, _, _ := strings.Cut(pair, "=")
			key = strings.TrimSpace(hurlVarRe.ReplaceAllString(key, "$1"))
			if key == "" {
				continue
			}
			queryKeys = append(queryKeys, key)
		}
		if len(queryKeys) > 0 {
			slug += "__q-" + strings.Join(queryKeys, "-")
		}
	}

	out = slug
	clean := make([]rune, 0, len(out))
	for _, r := range out {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '/' || r == '-' {
			clean = append(clean, r)
		} else {
			clean = append(clean, '-')
		}
	}
	out = strings.Trim(string(clean), "-/")
	out = strings.ReplaceAll(out, "/", "__")
	if out == "" {
		return "request"
	}
	return out
}

func splitHurlFile(inputPath, outputDir string) (int, error) {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	blocks := make([][]string, 0)
	current := make([]string, 0)

	for _, line := range lines {
		if isRequestLine(line) && len(current) > 0 {
			blocks = append(blocks, current)
			current = []string{line}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, current)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return 0, err
	}

	count := 0
	for _, block := range blocks {
		requestLine := ""
		for _, line := range block {
			if isRequestLine(line) {
				requestLine = line
				break
			}
		}
		if requestLine == "" {
			continue
		}

		method, url := parseRequestLine(requestLine)
		if method == "" {
			continue
		}
		count++
		name := fmt.Sprintf("%03d-%s-%s.hurl", count, strings.ToLower(method), slugifyURL(url))
		path := filepath.Join(outputDir, name)
		content := strings.TrimSpace(strings.Join(block, "\n")) + "\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return count, err
		}
	}
	return count, nil
}

func isRequestLine(line string) bool {
	methods := []string{"GET ", "POST ", "PUT ", "PATCH ", "DELETE ", "HEAD ", "OPTIONS ", "TRACE ", "CONNECT "}
	for _, m := range methods {
		if strings.HasPrefix(line, m) {
			return true
		}
	}
	return false
}

func parseRequestLine(line string) (method, url string) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func clearGeneratedHurl(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.hurl"))
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

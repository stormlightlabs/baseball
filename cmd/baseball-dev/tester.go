package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/utils"
)

type testRunRes struct {
	path     string
	duration time.Duration
	output   string
	err      error
}

type testRunConf struct {
	host        string
	dir         string
	match       string
	concurrency int
	hurlBin     string
}

type styles struct {
	title   lipgloss.Style
	ok      lipgloss.Style
	fail    lipgloss.Style
	muted   lipgloss.Style
	summary lipgloss.Style
}

func stylesheet() styles {
	return styles{
		title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")),
		ok:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
		fail:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
		muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		summary: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")),
	}
}

func newRunCmd() *cobra.Command {
	cfg := testRunConf{
		host:        "http://localhost:8080",
		dir:         "test/generated",
		match:       "",
		concurrency: runtime.NumCPU(),
		hurlBin:     "hurl",
	}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run generated Hurl tests concurrently",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTestRunner(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.host, "host", cfg.host, "Host URL used for Hurl variables (host and BASE_URL)")
	cmd.Flags().StringVar(&cfg.dir, "dir", cfg.dir, "Directory containing split .hurl files")
	cmd.Flags().StringVar(&cfg.match, "match", cfg.match, "Fuzzy filter on test file name")
	cmd.Flags().IntVar(&cfg.concurrency, "concurrency", cfg.concurrency, "Number of tests to run concurrently")
	cmd.Flags().StringVar(&cfg.hurlBin, "hurl-bin", cfg.hurlBin, "Hurl executable name/path")
	return cmd
}

func runTestRunner(cfg testRunConf) error {
	if cfg.concurrency < 1 {
		cfg.concurrency = 1
	}
	if _, err := exec.LookPath(cfg.hurlBin); err != nil {
		return fmt.Errorf("missing required command: %s", cfg.hurlBin)
	}
	return runTestsWithOutput(cfg)
}

func runTestsWithOutput(cfg testRunConf) error {
	s := stylesheet()
	files, err := discoverTests(cfg.dir)
	if err != nil {
		return err
	}
	if cfg.match != "" {
		files = filterTests(files, cfg.match)
	}
	if len(files) == 0 {
		if cfg.match == "" {
			return fmt.Errorf("no .hurl files found in %s", cfg.dir)
		}
		return fmt.Errorf("no tests matched pattern %q in %s", cfg.match, cfg.dir)
	}

	fmt.Println(s.title.Render("Big Fly / Baseball Dev Tools"))
	fmt.Println(s.muted.Render(fmt.Sprintf("Host: %s", cfg.host)))
	fmt.Println(s.muted.Render(fmt.Sprintf("Dir: %s", cfg.dir)))
	if cfg.match != "" {
		fmt.Println(s.muted.Render(fmt.Sprintf("Match: %s", cfg.match)))
	}
	fmt.Println(s.muted.Render(fmt.Sprintf("Concurrency: %d", cfg.concurrency)))
	fmt.Println()

	start := time.Now()
	results := runTests(files, cfg)
	totalDuration := time.Since(start)

	failed := 0
	for _, file := range files {
		result := results[file]
		name := filepath.Base(file)
		if result.err == nil {
			fmt.Printf("%s %s %s\n", s.ok.Render("PASS"), name, s.muted.Render(fmt.Sprintf("(%s)", result.duration.Round(time.Millisecond))))
			continue
		}

		failed++
		fmt.Printf("%s %s %s\n", s.fail.Render("FAIL"), name, s.muted.Render(fmt.Sprintf("(%s)", result.duration.Round(time.Millisecond))))
		output := strings.TrimSpace(result.output)
		if output != "" {
			fmt.Println(utils.IndentLines(utils.TruncateLines(output, 24), "  "))
		}
		fmt.Println()
	}

	passed := len(files) - failed
	fmt.Println(s.summary.Render(
		fmt.Sprintf("Summary: total=%d passed=%d failed=%d elapsed=%s", len(files), passed, failed, totalDuration.Round(time.Millisecond)),
	))
	if failed > 0 {
		return errors.New("one or more tests failed")
	}
	return nil
}

func discoverTests(dir string) ([]string, error) {
	pattern := filepath.Join(dir, "*.hurl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob %s: %w", pattern, err)
	}
	sort.Strings(files)
	return files, nil
}

func filterTests(files []string, query string) []string {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return files
	}

	filtered := make([]string, 0, len(files))
	for _, file := range files {
		name := strings.ToLower(filepath.Base(file))
		if strings.Contains(name, q) || isFuzzySubsequence(q, name) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func runTests(files []string, cfg testRunConf) map[string]testRunRes {
	jobs := make(chan string)
	resultsCh := make(chan testRunRes, len(files))

	var wg sync.WaitGroup
	for i := 0; i < cfg.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				resultsCh <- runOneTest(file, cfg)
			}
		}()
	}

	go func() {
		for _, file := range files {
			jobs <- file
		}
		close(jobs)
		wg.Wait()
		close(resultsCh)
	}()

	results := make(map[string]testRunRes, len(files))
	for result := range resultsCh {
		results[result.path] = result
	}
	return results
}

func runOneTest(file string, cfg testRunConf) testRunRes {
	start := time.Now()
	cmd := exec.Command(
		cfg.hurlBin,
		"--test",
		"--variable", "host="+cfg.host,
		"--variable", "BASE_URL="+cfg.host,
		file,
	)
	out, err := cmd.CombinedOutput()
	return testRunRes{
		path:     file,
		duration: time.Since(start),
		output:   string(out),
		err:      err,
	}
}

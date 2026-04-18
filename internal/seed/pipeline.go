package seed

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

type PipelineProfile string

const (
	PipelineProfileDev  PipelineProfile = "dev"
	PipelineProfileProd PipelineProfile = "prod"
)

type PipelineMode string

const (
	PipelineModeIncremental PipelineMode = "incremental"
	PipelineModeFull        PipelineMode = "full"
)

// PipelineOptions configures a full ETL pipeline run.
type PipelineOptions struct {
	Profile           PipelineProfile
	Mode              PipelineMode
	Years             []int
	EraNames          []string
	LahmanCSVDir      string
	RetrosheetDataDir string
	FanGraphsDir      string
	SalaryDataDir     string
}

// PipelineRunResult summarizes a completed pipeline run.
type PipelineRunResult struct {
	RunID      int64
	Profile    PipelineProfile
	Mode       PipelineMode
	Years      []int
	TotalRows  int64
	Validation PipelineValidationResult
}

// PipelineValidationIssue reports one validation finding.
type PipelineValidationIssue struct {
	Severity string
	Dataset  string
	Message  string
}

// PipelineValidationResult captures validation outcome.
type PipelineValidationResult struct {
	Profile PipelineProfile
	Issues  []PipelineValidationIssue
}

func (r PipelineValidationResult) Errors() []PipelineValidationIssue {
	out := make([]PipelineValidationIssue, 0, len(r.Issues))
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			out = append(out, issue)
		}
	}
	return out
}

func (r PipelineValidationResult) Warnings() []PipelineValidationIssue {
	out := make([]PipelineValidationIssue, 0, len(r.Issues))
	for _, issue := range r.Issues {
		if issue.Severity == "warning" {
			out = append(out, issue)
		}
	}
	return out
}

func (r PipelineValidationResult) OK() bool { return len(r.Errors()) == 0 }

// NormalizePipelineOptions applies defaults and validates inputs.
func NormalizePipelineOptions(opts PipelineOptions) (PipelineOptions, error) {
	if opts.Profile == "" {
		opts.Profile = PipelineProfileDev
	}
	if opts.Mode == "" {
		opts.Mode = PipelineModeIncremental
	}
	if opts.LahmanCSVDir == "" {
		opts.LahmanCSVDir = filepath.Join("data", "lahman", "csv")
	}
	if opts.RetrosheetDataDir == "" {
		opts.RetrosheetDataDir = filepath.Join("data", "retrosheet")
	}
	if opts.FanGraphsDir == "" {
		opts.FanGraphsDir = filepath.Join("data", "fangraphs")
	}
	if opts.SalaryDataDir == "" {
		opts.SalaryDataDir = filepath.Join("data", "salaries")
	}

	if opts.Profile != PipelineProfileDev && opts.Profile != PipelineProfileProd {
		return opts, fmt.Errorf("invalid profile %q (valid: dev, prod)", opts.Profile)
	}
	if opts.Mode != PipelineModeIncremental && opts.Mode != PipelineModeFull {
		return opts, fmt.Errorf("invalid mode %q (valid: incremental, full)", opts.Mode)
	}

	years, err := resolvePipelineYears(opts.Profile, opts.Years, opts.EraNames)
	if err != nil {
		return opts, err
	}
	opts.Years = years
	return opts, nil
}

// RunPipeline executes the default ETL pipeline.
func RunPipeline(ctx context.Context, database *db.DB, opts PipelineOptions) (PipelineRunResult, error) {
	opts, err := NormalizePipelineOptions(opts)
	if err != nil {
		return PipelineRunResult{}, err
	}

	params := map[string]any{
		"years": opts.Years,
		"eras":  opts.EraNames,
	}

	runID, err := database.StartETLRun(ctx, string(opts.Profile), string(opts.Mode), params)
	if err != nil {
		return PipelineRunResult{}, err
	}

	result := PipelineRunResult{
		RunID:   runID,
		Profile: opts.Profile,
		Mode:    opts.Mode,
		Years:   slices.Clone(opts.Years),
	}

	runStatus := "completed"
	runErrMsg := ""
	defer func() {
		if err := database.FinishETLRun(ctx, runID, runStatus, runErrMsg); err != nil {
			echo.Infof("⚠ Failed to finalize ETL run %d: %v", runID, err)
		}
	}()

	force := opts.Mode == PipelineModeFull
	skipLahman := opts.Mode == PipelineModeIncremental

	echo.Header("ETL Pipeline Run")
	echo.Infof("Run ID: %d", runID)
	echo.Infof("Profile: %s", opts.Profile)
	echo.Infof("Mode: %s", opts.Mode)
	echo.Infof("Years: %s", summarizeYears(opts.Years))

	rows, stepErr := runPipelineStep(ctx, database, runID, "extract.retrosheet", func() (int64, error) {
		return 0, FetchRetrosheetData(ctx, opts.RetrosheetDataDir, opts.Years, force)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "extract.negroleagues", func() (int64, error) {
		return 0, FetchNegroLeaguesData(ctx, filepath.Join(opts.RetrosheetDataDir, "negroleagues"), force)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.lahman", func() (int64, error) {
		return LoadLahman(ctx, database, LahmanOptions{CSVDir: opts.LahmanCSVDir, Skip: skipLahman})
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.retrosheet", func() (int64, error) {
		loaded, err := LoadRetrosheet(ctx, database, RetrosheetOptions{
			DataDir: opts.RetrosheetDataDir,
			Years:   opts.Years,
			Force:   force,
		})
		if err != nil {
			return 0, err
		}
		return loaded.GameRows + loaded.PlayRows, nil
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.fangraphs", func() (int64, error) {
		return LoadFanGraphsData(ctx, database, opts.FanGraphsDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.salary", func() (int64, error) {
		return LoadSalaryData(ctx, database, opts.SalaryDataDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.retrosheet_players", func() (int64, error) {
		csvPath, err := EnsureRetrosheetPlayersCSV(opts.RetrosheetDataDir)
		if err != nil {
			return 0, err
		}
		return LoadRetrosheetPlayers(ctx, database, csvPath)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.biodata", func() (int64, error) {
		tmpDir, cleanup, err := ExtractBiodataArchive(opts.RetrosheetDataDir)
		if err != nil {
			return 0, err
		}
		defer cleanup()
		return LoadBiodata(ctx, database, tmpDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.weather", func() (int64, error) {
		return LoadWeatherData(ctx, database, filepath.Join(opts.RetrosheetDataDir, "gameinfo.csv"))
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.parks", func() (int64, error) {
		return LoadParksData(ctx, database)
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.allstar", func() (int64, error) {
		return LoadAllStarData(ctx, database, filepath.Join(opts.RetrosheetDataDir, "allstar", "allstar.zip"))
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "refresh.materialized_views", func() (int64, error) {
		views, err := database.RefreshMaterializedViews(ctx, nil)
		return int64(views), err
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "validate", func() (int64, error) {
		validation, err := ValidatePipeline(ctx, database, opts.Profile, opts.Years)
		if err != nil {
			return 0, err
		}
		result.Validation = validation
		if !validation.OK() {
			return int64(len(validation.Issues)), fmt.Errorf("validation failed: %d error(s)", len(validation.Errors()))
		}
		return int64(len(validation.Issues)), nil
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	echo.Successf("✓ ETL pipeline complete (run_id=%d)", runID)
	return result, nil
}

func runPipelineStep(ctx context.Context, database *db.DB, runID int64, step string, fn func() (int64, error)) (int64, error) {
	stepID, err := database.StartETLStep(ctx, runID, step)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	echo.Info("")
	echo.Infof("Step: %s", step)

	rowCount, stepErr := fn()
	if stepErr != nil {
		_ = database.FinishETLStep(ctx, stepID, "failed", rowCount, stepErr.Error())
		return rowCount, stepErr
	}

	if err := database.FinishETLStep(ctx, stepID, "completed", rowCount, ""); err != nil {
		return rowCount, err
	}

	echo.Successf("✓ %s (%s)", step, time.Since(start).Round(time.Second))
	return rowCount, nil
}

// ValidatePipeline verifies completeness for the chosen profile.
func ValidatePipeline(ctx context.Context, database *db.DB, profile PipelineProfile, expectedYears []int) (PipelineValidationResult, error) {
	if profile == "" {
		profile = PipelineProfileDev
	}

	result := PipelineValidationResult{Profile: profile}
	addErr := func(dataset, msg string) {
		result.Issues = append(result.Issues, PipelineValidationIssue{
			Severity: "error",
			Dataset:  dataset,
			Message:  msg,
		})
	}
	addWarn := func(dataset, msg string) {
		result.Issues = append(result.Issues, PipelineValidationIssue{
			Severity: "warning",
			Dataset:  dataset,
			Message:  msg,
		})
	}

	count := func(query string, args ...any) (int64, error) {
		var n int64
		if err := database.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}

	people, err := count(`SELECT COUNT(*) FROM "People"`)
	if err != nil {
		return result, fmt.Errorf("lahman check failed: %w", err)
	}
	teams, err := count(`SELECT COUNT(*) FROM "Teams"`)
	if err != nil {
		return result, fmt.Errorf("lahman check failed: %w", err)
	}
	games, err := count(`SELECT COUNT(*) FROM games`)
	if err != nil {
		return result, fmt.Errorf("retrosheet check failed: %w", err)
	}
	plays, err := count(`SELECT COUNT(*) FROM plays`)
	if err != nil {
		return result, fmt.Errorf("retrosheet check failed: %w", err)
	}
	woba, err := count(`SELECT COUNT(*) FROM woba_constants`)
	if err != nil {
		return result, fmt.Errorf("fangraphs check failed: %w", err)
	}
	leagueConstants, err := count(`SELECT COUNT(*) FROM league_constants`)
	if err != nil {
		return result, fmt.Errorf("fangraphs check failed: %w", err)
	}
	parkFactors, err := count(`SELECT COUNT(*) FROM park_factors`)
	if err != nil {
		return result, fmt.Errorf("fangraphs check failed: %w", err)
	}
	salarySummary, err := count(`SELECT COUNT(*) FROM salary_summary`)
	if err != nil {
		return result, fmt.Errorf("salary check failed: %w", err)
	}
	retrosheetPlayers, err := count(`SELECT COUNT(*) FROM retrosheet_players`)
	if err != nil {
		return result, fmt.Errorf("retrosheet players check failed: %w", err)
	}
	bio, err := count(`SELECT COUNT(*) FROM player_bio_extended`)
	if err != nil {
		return result, fmt.Errorf("biodata check failed: %w", err)
	}
	relatives, err := count(`SELECT COUNT(*) FROM player_relatives`)
	if err != nil {
		return result, fmt.Errorf("biodata check failed: %w", err)
	}
	coaches, err := count(`SELECT COUNT(*) FROM coaches`)
	if err != nil {
		return result, fmt.Errorf("biodata check failed: %w", err)
	}
	umpires, err := count(`SELECT COUNT(*) FROM umpires`)
	if err != nil {
		return result, fmt.Errorf("biodata check failed: %w", err)
	}
	weatherRows, err := count(`SELECT COUNT(*) FROM games WHERE temp_f IS NOT NULL OR wind_speed_mph IS NOT NULL OR sky IS NOT NULL OR precip IS NOT NULL OR field_condition IS NOT NULL`)
	if err != nil {
		return result, fmt.Errorf("weather check failed: %w", err)
	}
	parkMap, err := count(`SELECT COUNT(*) FROM park_map`)
	if err != nil {
		return result, fmt.Errorf("parks check failed: %w", err)
	}
	allStarGames, err := count(`SELECT COUNT(*) FROM games WHERE game_type = 'allstar'`)
	if err != nil {
		return result, fmt.Errorf("all-star check failed: %w", err)
	}
	allStarPlays, err := count(`SELECT COUNT(*) FROM plays WHERE gid LIKE 'ALS%'`)
	if err != nil {
		return result, fmt.Errorf("all-star check failed: %w", err)
	}
	negroGames, err := count(`SELECT COUNT(*) FROM games WHERE league IN ('NAL', 'NNL', 'NN2', 'ECL', 'ANL', 'EWL', 'NSL', 'IND')`)
	if err != nil {
		return result, fmt.Errorf("negro leagues check failed: %w", err)
	}

	if people == 0 || teams == 0 {
		addErr("lahman", "Lahman tables are empty")
	}
	if games == 0 || plays == 0 {
		addErr("retrosheet", "Retrosheet games/plays are not fully loaded")
	}
	if woba == 0 || leagueConstants == 0 || parkFactors == 0 {
		addErr("fangraphs_constants", "FanGraphs constants are incomplete")
	}
	if salarySummary == 0 {
		addErr("salary_summary", "Salary summary is empty")
	}
	if retrosheetPlayers == 0 {
		addErr("retrosheet_players", "Retrosheet players crosswalk is empty")
	}
	if bio+relatives+coaches+umpires == 0 {
		addErr("biodata", "Retrosheet biodata tables are empty")
	}
	if weatherRows == 0 {
		addErr("weather", "No weather metadata was applied to games")
	}
	if parkMap == 0 {
		addErr("parks_metadata", "park_map is empty")
	}
	if allStarGames == 0 || allStarPlays == 0 {
		addErr("allstar", "All-Star games/plays are incomplete")
	}
	if negroGames == 0 {
		addErr("negroleagues", "Negro Leagues games were not loaded")
	}

	var minDate, maxDate string
	if err := database.QueryRowContext(ctx, `SELECT COALESCE(MIN(date), ''), COALESCE(MAX(date), '') FROM games`).Scan(&minDate, &maxDate); err != nil {
		return result, fmt.Errorf("retrosheet coverage check failed: %w", err)
	}
	minYear := yearFromDate(minDate)
	maxYear := yearFromDate(maxDate)

	switch profile {
	case PipelineProfileProd:
		if minYear == 0 || minYear > 1910 {
			addErr("coverage", fmt.Sprintf("expected historical coverage to start by 1910, got %d", minYear))
		}
		expectedMax := time.Now().Year() - 1
		if maxYear == 0 || maxYear < expectedMax {
			addErr("coverage", fmt.Sprintf("expected modern coverage through at least %d, got %d", expectedMax, maxYear))
		}
	default:
		if len(expectedYears) == 0 {
			expectedYears = devRepresentativeYears()
		}
		missingYears, err := missingGameYears(ctx, database, expectedYears)
		if err != nil {
			return result, err
		}
		if len(missingYears) > 0 {
			addWarn("coverage", fmt.Sprintf("representative years not present: %s", summarizeYears(missingYears)))
		}
	}

	return result, nil
}

func resolvePipelineYears(profile PipelineProfile, explicitYears []int, eraNames []string) ([]int, error) {
	var years []int
	if len(explicitYears) > 0 {
		years = append(years, explicitYears...)
	} else {
		switch profile {
		case PipelineProfileProd:
			years = prodExhaustiveYears()
		default:
			years = devRepresentativeYears()
		}
	}

	for _, eraName := range eraNames {
		normalized := strings.ToLower(strings.TrimSpace(eraName))
		if normalized == "" {
			continue
		}
		era := GetEra(normalized)
		if era == nil {
			return nil, fmt.Errorf("unknown era %q", eraName)
		}
		years = append(years, era.Years()...)
	}

	years = uniqueSortedYears(years)
	if len(years) == 0 {
		return nil, fmt.Errorf("no years resolved for ETL pipeline")
	}
	return years, nil
}

func devRepresentativeYears() []int {
	end := time.Now().Year() - 1
	start := end - 3
	recent := make([]int, 0, 4)
	for y := start; y <= end; y++ {
		recent = append(recent, y)
	}

	samples := []int{1914, 1935, 1955, 1968, 1985, 2001, 2010, 2017}
	return uniqueSortedYears(append(recent, samples...))
}

func prodExhaustiveYears() []int {
	end := time.Now().Year() - 1
	years := make([]int, 0, end-1910+1)
	for y := 1910; y <= end; y++ {
		years = append(years, y)
	}
	return years
}

func uniqueSortedYears(values []int) []int {
	if len(values) == 0 {
		return values
	}

	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func summarizeYears(years []int) string {
	years = uniqueSortedYears(years)
	if len(years) == 0 {
		return "0 years"
	}
	return fmt.Sprintf("%d years: %d-%d", len(years), years[0], years[len(years)-1])
}

func yearFromDate(value string) int {
	if len(value) < 4 {
		return 0
	}
	year, err := strconv.Atoi(value[:4])
	if err != nil {
		return 0
	}
	return year
}

func missingGameYears(ctx context.Context, database *db.DB, years []int) ([]int, error) {
	years = uniqueSortedYears(years)
	missing := make([]int, 0)
	for _, year := range years {
		var count int64
		if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM games WHERE SUBSTRING(date, 1, 4) = $1`, fmt.Sprintf("%04d", year)).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to verify coverage for %d: %w", year, err)
		}
		if count == 0 {
			missing = append(missing, year)
		}
	}
	return missing, nil
}

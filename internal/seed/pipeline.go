package seed

import (
	"context"
	"errors"
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
	Profile             PipelineProfile
	Mode                PipelineMode
	Years               []int
	EraNames            []string
	DataRoot            string
	LahmanCSVDir        string
	RetrosheetDataDir   string
	ChadwickDataDir     string
	FanGraphsDir        string
	SalaryDataDir       string
	NetworkRetryMax     int
	NetworkRetryBackoff time.Duration
	LoadChunkSize       int
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
	opts.DataRoot = ResolveDataRoot(opts.DataRoot)
	if opts.LahmanCSVDir == "" {
		opts.LahmanCSVDir = LahmanCSVDir(opts.DataRoot)
	}
	if opts.RetrosheetDataDir == "" {
		opts.RetrosheetDataDir = RetrosheetDir(opts.DataRoot)
	}
	if opts.FanGraphsDir == "" {
		opts.FanGraphsDir = FanGraphsDir(opts.DataRoot)
	}
	if opts.ChadwickDataDir == "" {
		opts.ChadwickDataDir = ChadwickDir(opts.DataRoot)
	}
	if opts.SalaryDataDir == "" {
		opts.SalaryDataDir = SalariesDir(opts.DataRoot)
	}
	if opts.NetworkRetryMax < 0 {
		opts.NetworkRetryMax = 0
	}
	if opts.NetworkRetryBackoff <= 0 {
		opts.NetworkRetryBackoff = defaultETLNetworkRetryBackoff
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

	provisioned, err := ensurePipelineDataRoot(ctx, opts)
	if err != nil {
		return PipelineRunResult{}, err
	}
	defer provisioned.cleanup()
	if provisioned.rootPath != opts.DataRoot {
		opts.DataRoot = provisioned.rootPath
		opts.LahmanCSVDir = LahmanCSVDir(opts.DataRoot)
		opts.RetrosheetDataDir = RetrosheetDir(opts.DataRoot)
		opts.FanGraphsDir = FanGraphsDir(opts.DataRoot)
		opts.ChadwickDataDir = ChadwickDir(opts.DataRoot)
		opts.SalaryDataDir = SalariesDir(opts.DataRoot)
	}

	params := map[string]any{
		"years":                 opts.Years,
		"eras":                  opts.EraNames,
		"data_root":             opts.DataRoot,
		"network_retry_max":     opts.NetworkRetryMax,
		"network_retry_backoff": opts.NetworkRetryBackoff.String(),
		"load_chunk_size":       opts.LoadChunkSize,
	}

	runID, err := database.StartETLRun(ctx, string(opts.Profile), string(opts.Mode), params)
	if err != nil {
		return PipelineRunResult{}, err
	}
	if err := database.MarkETLRunRunning(ctx, runID); err != nil {
		return PipelineRunResult{}, err
	}

	result := PipelineRunResult{
		RunID:   runID,
		Profile: opts.Profile,
		Mode:    opts.Mode,
		Years:   slices.Clone(opts.Years),
	}

	runStatus := "succeeded"
	runErrMsg := ""
	markRunFailure := func(stepErr error) {
		runErrMsg = stepErr.Error()
		if errors.Is(stepErr, context.Canceled) || errors.Is(stepErr, context.DeadlineExceeded) {
			runStatus = "cancelled"
			return
		}
		runStatus = "failed"
	}
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

	networkRetryPolicy := pipelineStepRetryPolicy{
		MaxRetries: opts.NetworkRetryMax,
		Backoff:    opts.NetworkRetryBackoff,
		RetryClass: "network",
	}

	rows, stepErr := runPipelineStepWithRetry(ctx, database, runID, "extract.retrosheet", networkRetryPolicy, func(stepCtx context.Context) (int64, error) {
		return 0, FetchRetrosheetData(stepCtx, opts.RetrosheetDataDir, opts.Years, force)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStepWithRetry(ctx, database, runID, "extract.chadwick", networkRetryPolicy, func(stepCtx context.Context) (int64, error) {
		return 0, FetchChadwickRegisterData(stepCtx, opts.ChadwickDataDir, force)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStepWithRetry(ctx, database, runID, "extract.negroleagues", networkRetryPolicy, func(stepCtx context.Context) (int64, error) {
		return 0, FetchNegroLeaguesData(stepCtx, filepath.Join(opts.RetrosheetDataDir, "negroleagues"), force)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.lahman", func(stepCtx context.Context) (int64, error) {
		return LoadLahman(stepCtx, database, LahmanOptions{CSVDir: opts.LahmanCSVDir, Skip: skipLahman})
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.retrosheet", func(stepCtx context.Context) (int64, error) {
		loaded, err := LoadRetrosheet(stepCtx, database, RetrosheetOptions{
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
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.fangraphs", func(stepCtx context.Context) (int64, error) {
		return LoadFanGraphsData(stepCtx, database, opts.FanGraphsDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.salary", func(stepCtx context.Context) (int64, error) {
		return LoadSalaryData(stepCtx, database, opts.SalaryDataDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.retrosheet_players", func(stepCtx context.Context) (int64, error) {
		csvPath, err := EnsureRetrosheetPlayersCSV(opts.RetrosheetDataDir)
		if err != nil {
			return 0, err
		}
		return LoadRetrosheetPlayers(stepCtx, database, csvPath)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.biodata", func(stepCtx context.Context) (int64, error) {
		tmpDir, cleanup, err := ExtractBiodataArchive(opts.RetrosheetDataDir)
		if err != nil {
			return 0, err
		}
		defer cleanup()
		return LoadBiodata(stepCtx, database, tmpDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.crosswalk.players_mlbam", func(stepCtx context.Context) (int64, error) {
		return LoadPlayerMLBAMMappings(stepCtx, database, opts.ChadwickDataDir)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.crosswalk.teams_mlbam", func(stepCtx context.Context) (int64, error) {
		return LoadTeamMLBAMMappings(stepCtx, database, opts.Years)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.weather", func(stepCtx context.Context) (int64, error) {
		return LoadWeatherData(stepCtx, database, filepath.Join(opts.RetrosheetDataDir, "gameinfo.csv"))
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.parks", func(stepCtx context.Context) (int64, error) {
		return LoadParksData(stepCtx, database)
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "load.allstar", func(stepCtx context.Context) (int64, error) {
		return LoadAllStarData(stepCtx, database, filepath.Join(opts.RetrosheetDataDir, "allstar", "allstar.zip"))
	})
	result.TotalRows += rows
	if stepErr != nil {
		markRunFailure(stepErr)
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "refresh.materialized_views", func(stepCtx context.Context) (int64, error) {
		refreshRunID := runID
		views, err := RefreshPipelineMaterializedViews(stepCtx, database, db.MaterializedViewRefreshOptions{
			RunID:              &refreshRunID,
			Step:               "refresh.materialized_views",
			ForceNonConcurrent: true,
		})
		return int64(views), err
	})
	result.TotalRows += rows
	if stepErr != nil {
		runStatus = "failed"
		runErrMsg = stepErr.Error()
		return result, stepErr
	}

	rows, stepErr = runPipelineStep(ctx, database, runID, "validate", func(stepCtx context.Context) (int64, error) {
		validation, err := ValidatePipeline(stepCtx, database, opts.Profile, opts.Years)
		if err != nil {
			return 0, err
		}
		result.Validation = validation
		if !validation.OK() {
			errors := validation.Errors()
			for _, issue := range errors {
				echo.Infof("✗ [%s] %s", issue.Dataset, issue.Message)
			}
			for _, issue := range validation.Warnings() {
				echo.Infof("⚠ [%s] %s", issue.Dataset, issue.Message)
			}

			parts := make([]string, 0, len(errors))
			for _, issue := range errors {
				parts = append(parts, fmt.Sprintf("[%s] %s", issue.Dataset, issue.Message))
			}
			return int64(len(validation.Issues)), fmt.Errorf("validation failed: %s", strings.Join(parts, "; "))
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

func runPipelineStep(ctx context.Context, database *db.DB, runID int64, step string, fn func(context.Context) (int64, error)) (int64, error) {
	stepID, err := database.StartETLStep(ctx, runID, step)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	echo.Info("")
	echo.Infof("Step: %s", step)

	stepCtx := withETLStepContext(ctx, runID, step)
	rowCount, stepErr := fn(stepCtx)
	if stepErr != nil {
		_ = database.FinishETLStep(ctx, stepID, "failed", rowCount, stepErr.Error())
		return rowCount, fmt.Errorf("%s: %w", step, stepErr)
	}

	if err := database.FinishETLStep(ctx, stepID, "completed", rowCount, ""); err != nil {
		return rowCount, err
	}

	echo.Successf("✓ %s (%s)", step, time.Since(start).Round(time.Second))
	return rowCount, nil
}

type pipelineStepRetryPolicy struct {
	MaxRetries int
	Backoff    time.Duration
	RetryClass string
}

func runPipelineStepWithRetry(
	ctx context.Context,
	database *db.DB,
	runID int64,
	step string,
	policy pipelineStepRetryPolicy,
	fn func(context.Context) (int64, error),
) (int64, error) {
	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}
	if policy.Backoff <= 0 {
		policy.Backoff = defaultETLNetworkRetryBackoff
	}

	var rows int64
	var stepErr error

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		rows, stepErr = runPipelineStep(ctx, database, runID, step, fn)
		if stepErr == nil {
			return rows, nil
		}
		if attempt == policy.MaxRetries {
			return rows, stepErr
		}

		delay := time.Duration(attempt+1) * policy.Backoff
		echo.Infof("  Retrying step=%s attempt=%d/%d in %s due to: %v", step, attempt+2, policy.MaxRetries+1, delay.Round(time.Millisecond), stepErr)

		recordETLPhaseEvent(
			withETLStepContext(ctx, runID, step),
			database,
			step,
			"retry_wait",
			"retrying",
			rows,
			time.Now(),
			map[string]any{
				"attempt":      attempt + 2,
				"max_attempts": policy.MaxRetries + 1,
				"retry_class":  policy.RetryClass,
				"delay_ms":     delay.Milliseconds(),
			},
			stepErr,
		)

		select {
		case <-ctx.Done():
			return rows, ctx.Err()
		case <-time.After(delay):
		}
	}

	return rows, stepErr
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
	playerMLBAMMap, err := count(`SELECT COUNT(*) FROM player_mlbam_map`)
	if err != nil {
		return result, fmt.Errorf("mlbam crosswalk check failed: %w", err)
	}
	teamMLBAMMap, err := count(`SELECT COUNT(*) FROM team_mlbam_map`)
	if err != nil {
		return result, fmt.Errorf("mlbam crosswalk check failed: %w", err)
	}
	allStarGames, err := count(`SELECT COUNT(*) FROM games WHERE game_type = 'allstar'`)
	if err != nil {
		return result, fmt.Errorf("all-star check failed: %w", err)
	}
	allStarPlays, err := count(`
		SELECT COUNT(*)
		FROM plays p
		JOIN games g ON g.game_id = p.gid
		WHERE g.game_type = 'allstar'
	`)
	if err != nil {
		return result, fmt.Errorf("all-star check failed: %w", err)
	}
	negroGames, err := count(`
		SELECT COUNT(*)
		FROM games
		WHERE home_team_league IN ('NAL', 'NNL', 'NN2', 'ECL', 'ANL', 'EWL', 'NSL', 'IND')
		   OR visiting_team_league IN ('NAL', 'NNL', 'NN2', 'ECL', 'ANL', 'EWL', 'NSL', 'IND')
	`)
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
	if playerMLBAMMap == 0 || teamMLBAMMap == 0 {
		addErr("mlbam_crosswalk", "MLBAM player/team crosswalk tables are empty")
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

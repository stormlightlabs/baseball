package etl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/cli/shared"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/seed"
	"stormlightlabs.org/baseball/internal/utils"
)

type cronCLIOptions struct {
	profile             string
	schedules           []string
	scheduleConfig      string
	disableScheduler    bool
	priority            int
	maxActiveJobs       int
	maxQueuedJobs       int
	batchDelay          time.Duration
	networkRetries      int
	networkRetryBackoff time.Duration
	loadChunkSize       int
	jobMaxRetries       int
	pollInterval        time.Duration
}

type currentSeasonCronConfig struct {
	CurrentSeason currentSeasonCronSection `toml:"current_season"`
}

type currentSeasonCronSection struct {
	Enabled     bool   `toml:"enabled"`
	Season      int    `toml:"season"`
	CronStats   string `toml:"cron_stats"`
	CronStand   string `toml:"cron_standings"`
	CronSched   string `toml:"cron_schedule"`
	CronRosters string `toml:"cron_rosters"`
	ActiveWin   string `toml:"active_window"`
}

type cronTask struct {
	SyncType     string
	ScheduleSpec string
	Season       int
	ActiveWindow string
	Source       string
}

func EtlCronCmd() *cobra.Command {
	opts := &cronCLIOptions{}
	cmd := &cobra.Command{
		Use:   "cron",
		Short: "Run ETL worker with current-season cron scheduler",
		Long: "Run a long-lived ETL worker loop and optionally register current-season cron schedules.\n" +
			"This command only schedules jobs; execution stays in the worker loop.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runETLCron(cmd, opts)
		},
	}
	addCronFlags(cmd, opts)
	return cmd
}

func addCronFlags(cmd *cobra.Command, opts *cronCLIOptions) {
	cmd.Flags().StringVar(&opts.profile, "profile", "current-season", "Profile used for queued current-season-sync jobs")
	cmd.Flags().StringArrayVar(&opts.schedules, "schedule", nil, "Cron expression (repeatable). Optional format sync_type=expr, e.g. stats=*/5 * * * *")
	cmd.Flags().StringVar(&opts.scheduleConfig, "schedule-config", "", "Optional schedule TOML path (defaults to --config)")
	cmd.Flags().BoolVar(&opts.disableScheduler, "disable-scheduler", false, "Disable scheduler registration and run worker loop only")
	cmd.Flags().IntVar(&opts.priority, "priority", 50, "Queue priority for scheduled jobs (lower runs first)")
	cmd.Flags().IntVar(&opts.maxActiveJobs, "max-active-jobs", 1, "Maximum active ETL jobs allowed concurrently")
	cmd.Flags().IntVar(&opts.maxQueuedJobs, "max-queued-jobs", 128, "Maximum queued+active ETL jobs before enqueue is rejected")
	cmd.Flags().DurationVar(&opts.batchDelay, "batch-delay", 0, "Delay between processed jobs (e.g. 2s)")
	cmd.Flags().IntVar(&opts.networkRetries, "network-retries", 2, "Retries for network/download extraction steps")
	cmd.Flags().DurationVar(&opts.networkRetryBackoff, "network-retry-backoff", 3*time.Second, "Backoff between network retries")
	cmd.Flags().IntVar(&opts.loadChunkSize, "load-chunk-size", 0, "Optional load chunk size metadata for downstream loaders")
	cmd.Flags().IntVar(&opts.jobMaxRetries, "job-max-retries", 2, "Maximum scheduled job retries after failure")
	cmd.Flags().DurationVar(&opts.pollInterval, "poll-interval", 5*time.Second, "Queue poll interval when no jobs are available")
}

func runETLCron(cmd *cobra.Command, opts *cronCLIOptions) error {
	runCtx, stopSignals := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	echo.Header("ETL Cron + Worker")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()
	echo.Success("✓ Connected to database")

	tasks, err := resolveCronTasks(cmd, opts)
	if err != nil {
		return err
	}

	workerOpts := seed.JobWorkerOptions{
		MaxActiveJobs:       opts.maxActiveJobs,
		MaxQueuedJobs:       opts.maxQueuedJobs,
		BatchDelay:          opts.batchDelay,
		NetworkRetryMax:     opts.networkRetries,
		NetworkRetryBackoff: opts.networkRetryBackoff,
		LoadChunkSize:       opts.loadChunkSize,
		MaxJobRetries:       opts.jobMaxRetries,
	}

	scheduler := cron.New()
	if opts.disableScheduler {
		echo.Info("Scheduler disabled by --disable-scheduler; worker loop will continue without cron registration.")
	} else {
		if len(tasks) == 0 {
			echo.Info("No cron schedules configured; worker loop will continue without cron registration.")
		}
		var enqueueMu sync.Mutex
		for _, task := range tasks {
			taskCopy := task
			if _, err := scheduler.AddFunc(task.ScheduleSpec, func() {
				enqueueMu.Lock()
				defer enqueueMu.Unlock()

				if enqErr := enqueueCurrentSeasonCronTask(runCtx, database, taskCopy, opts); enqErr != nil {
					echo.Errorf("cron enqueue failed sync_type=%s schedule=%q err=%v", taskCopy.SyncType, taskCopy.ScheduleSpec, enqErr)
				}
			}); err != nil {
				return fmt.Errorf("invalid cron schedule %q for sync_type=%s: %w", task.ScheduleSpec, task.SyncType, err)
			}
			echo.Infof("Registered cron task sync_type=%s season=%d schedule=%q source=%s", task.SyncType, task.Season, task.ScheduleSpec, task.Source)
		}
		if len(tasks) > 0 {
			scheduler.Start()
			echo.Successf("✓ Scheduler online tasks=%d", len(tasks))
		}
	}

	defer func() {
		stopCtx := scheduler.Stop()
		select {
		case <-stopCtx.Done():
		case <-time.After(5 * time.Second):
			echo.Info("Scheduler stop timed out after 5s; exiting.")
		}
	}()

	echo.Infof("Starting worker loop (poll=%s)...", opts.pollInterval)
	if err := seed.RunETLWorker(runCtx, database, workerOpts, opts.pollInterval); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	echo.Success("✓ Worker stopped")
	return nil
}

func enqueueCurrentSeasonCronTask(ctx context.Context, database *db.DB, task cronTask, opts *cronCLIOptions) error {
	if task.ActiveWindow != "" {
		window, err := utils.ParseMonthDayWindow(task.ActiveWindow)
		if err != nil {
			return fmt.Errorf("invalid active_window %q: %w", task.ActiveWindow, err)
		}
		if !window.Contains(time.Now()) {
			echo.Infof("Skipping cron tick sync_type=%s season=%d outside active_window=%s", task.SyncType, task.Season, task.ActiveWindow)
			return nil
		}
	}

	season := strconv.Itoa(task.Season)
	pending, err := database.HasPendingETLJob(ctx, db.ETLJobTypeCurrentSync, opts.profile, task.SyncType, season)
	if err != nil {
		return err
	}
	if pending {
		echo.Infof("Skipping enqueue sync_type=%s season=%d (pending queued/running job exists)", task.SyncType, task.Season)
		return nil
	}

	scope := map[string]any{
		"profile":       opts.profile,
		"mode":          "incremental",
		"sync_type":     task.SyncType,
		"season":        task.Season,
		"active_window": task.ActiveWindow,
		"enqueued_at":   time.Now().UTC().Format(time.RFC3339),
	}

	options := map[string]any{
		"schedule":        task.ScheduleSpec,
		"schedule_source": task.Source,
	}

	jobID, err := database.EnqueueETLJob(ctx, db.ETLJobSpec{
		JobType:    db.ETLJobTypeCurrentSync,
		Priority:   opts.priority,
		Profile:    opts.profile,
		Mode:       "incremental",
		Scope:      scope,
		Options:    options,
		MaxRetries: opts.jobMaxRetries,
	}, opts.maxQueuedJobs)
	if err != nil {
		if errors.Is(err, db.ErrETLQueueFull) {
			echo.Infof("Skipping enqueue sync_type=%s season=%d queue full: %v", task.SyncType, task.Season, err)
			return nil
		}
		return err
	}

	echo.Successf("✓ Enqueued current-season-sync job id=%d sync_type=%s season=%d", jobID, task.SyncType, task.Season)
	return nil
}

func resolveCronTasks(cmd *cobra.Command, opts *cronCLIOptions) ([]cronTask, error) {
	if opts.disableScheduler {
		return nil, nil
	}

	configPath := strings.TrimSpace(opts.scheduleConfig)
	if configPath == "" {
		configPath = strings.TrimSpace(shared.FindConfigPath(cmd))
	}

	cfg, cfgErr := loadCurrentSeasonCronConfig(configPath)
	if cfgErr != nil && len(opts.schedules) == 0 {
		return nil, cfgErr
	}
	if cfgErr != nil {
		echo.Infof("Unable to load schedule config from %q; proceeding with explicit --schedule values: %v", configPath, cfgErr)
	}

	tasks := make([]cronTask, 0, 8)
	if cfgErr == nil && cfg.CurrentSeason.Enabled {
		cfgTasks, err := cronTasksFromConfig(cfg)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, cfgTasks...)
	}

	flagTasks, err := cronTasksFromFlags(opts.schedules, cfg.CurrentSeason.Season, cfg.CurrentSeason.ActiveWin)
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, flagTasks...)
	if len(tasks) == 0 {
		return nil, nil
	}
	return dedupeCronTasks(tasks), nil
}

func loadCurrentSeasonCronConfig(path string) (currentSeasonCronConfig, error) {
	var cfg currentSeasonCronConfig
	path = strings.TrimSpace(path)
	if path == "" {
		return cfg, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed reading schedule config %q: %w", path, err)
	}

	if err := toml.Unmarshal(content, &cfg); err != nil {
		return cfg, fmt.Errorf("failed parsing TOML config %q: %w", path, err)
	}
	return cfg, nil
}

func cronTasksFromConfig(cfg currentSeasonCronConfig) ([]cronTask, error) {
	season := cfg.CurrentSeason.Season
	if season <= 0 {
		season = time.Now().Year()
	}
	activeWindow := strings.TrimSpace(cfg.CurrentSeason.ActiveWin)
	if activeWindow != "" {
		if _, err := utils.ParseMonthDayWindow(activeWindow); err != nil {
			return nil, fmt.Errorf("invalid current_season.active_window: %w", err)
		}
	}

	out := make([]cronTask, 0, 4)
	add := func(syncType, spec string) {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			return
		}
		out = append(out, cronTask{
			SyncType:     syncType,
			ScheduleSpec: spec,
			Season:       season,
			ActiveWindow: activeWindow,
			Source:       "config",
		})
	}

	add("stats", cfg.CurrentSeason.CronStats)
	add("standings", cfg.CurrentSeason.CronStand)
	add("schedule", cfg.CurrentSeason.CronSched)
	add("rosters", cfg.CurrentSeason.CronRosters)
	return out, nil
}

func cronTasksFromFlags(values []string, defaultSeason int, defaultWindow string) ([]cronTask, error) {
	season := defaultSeason
	if season <= 0 {
		season = time.Now().Year()
	}

	defaultWindow = strings.TrimSpace(defaultWindow)
	if defaultWindow != "" {
		if _, err := utils.ParseMonthDayWindow(defaultWindow); err != nil {
			return nil, fmt.Errorf("invalid active_window in config: %w", err)
		}
	}

	out := make([]cronTask, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		syncType := "all"
		spec := raw
		if strings.Contains(raw, "=") {
			parts := strings.SplitN(raw, "=", 2)
			syncType = strings.TrimSpace(parts[0])
			spec = strings.TrimSpace(parts[1])
		}
		if err := validateCurrentSeasonSyncType(syncType); err != nil {
			return nil, err
		}
		if spec == "" {
			return nil, fmt.Errorf("empty cron expression in --schedule value %q", raw)
		}

		out = append(out, cronTask{
			SyncType:     syncType,
			ScheduleSpec: spec,
			Season:       season,
			ActiveWindow: defaultWindow,
			Source:       "flag",
		})
	}
	return out, nil
}

func dedupeCronTasks(tasks []cronTask) []cronTask {
	seen := make(map[string]struct{}, len(tasks))
	out := make([]cronTask, 0, len(tasks))
	for _, task := range tasks {
		key := fmt.Sprintf("%s|%d|%s|%s", task.SyncType, task.Season, task.ScheduleSpec, task.ActiveWindow)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, task)
	}
	return out
}

func validateCurrentSeasonSyncType(value string) error {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "all", "stats", "standings", "schedule", "rosters":
		return nil
	default:
		return fmt.Errorf("invalid sync_type %q (allowed: all,stats,standings,schedule,rosters)", value)
	}
}

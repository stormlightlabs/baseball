package cli

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

func EtlJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Inspect and administer ETL queue jobs",
		Long:  "List ETL queue jobs and clear stuck running jobs without direct SQL access.",
	}
	cmd.AddCommand(EtlJobsLsCmd())
	cmd.AddCommand(EtlJobsClearCmd())
	return cmd
}

func EtlJobsLsCmd() *cobra.Command {
	var statusFilter string
	var jobTypeFilter string
	var profileFilter string
	var limit int

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List ETL jobs",
		Long:  "List ETL jobs with optional filters for status, job type, and profile.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listETLJobs(cmd, statusFilter, jobTypeFilter, profileFilter, limit)
		},
	}
	cmd.Flags().StringVar(&statusFilter, "status", "", "Comma-separated statuses (queued,started,running,retry_wait,succeeded,failed,cancelled)")
	cmd.Flags().StringVar(&jobTypeFilter, "job-type", "", "Job type filter (full-run,yearly-sync,validate-only,cleanup-only,maintenance)")
	cmd.Flags().StringVar(&profileFilter, "profile", "", "Profile filter (dev|prod)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum jobs to return (1-500)")
	return cmd
}

func EtlJobsClearCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear running ETL jobs",
		Long:  "Move currently running ETL jobs back to retry_wait so the queue can proceed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clearRunningETLJobs(cmd, reason)
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "cleared by operator", "Reason recorded in etl_jobs.last_error")
	return cmd
}

func listETLJobs(cmd *cobra.Command, rawStatuses, rawJobType, rawProfile string, limit int) error {
	statuses, err := parseETLJobStatuses(rawStatuses)
	if err != nil {
		return err
	}
	jobType, err := parseETLJobTypeFilter(rawJobType)
	if err != nil {
		return err
	}
	if limit < 1 || limit > 500 {
		return fmt.Errorf("error: --limit must be between 1 and 500")
	}

	echo.Header("ETL Jobs")
	echo.Info("Connecting to database...")
	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()
	echo.Success("✓ Connected to database")

	jobs, err := database.ListETLJobs(cmd.Context(), db.ETLJobListFilter{
		Statuses: statuses,
		JobType:  jobType,
		Profile:  strings.TrimSpace(rawProfile),
		Limit:    limit,
	})
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	if len(jobs) == 0 {
		echo.Info("No ETL jobs matched the provided filters.")
		return nil
	}

	echo.Infof("Showing %d job(s):", len(jobs))
	for _, job := range jobs {
		echo.Infof(
			"  id=%d status=%s type=%s profile=%s mode=%s attempts=%d/%d queued=%s started=%s worker=%s error=%s",
			job.ID,
			job.Status,
			job.JobType,
			blankAsDash(job.Profile),
			blankAsDash(job.Mode),
			job.Attempts,
			job.MaxRetries+1,
			job.QueuedAt.Format(time.RFC3339),
			formatNullableTime(job.StartedAt),
			blankAsDash(job.WorkerID),
			compactError(job.LastError),
		)
	}

	return nil
}

func clearRunningETLJobs(cmd *cobra.Command, reason string) error {
	echo.Header("Clear Running ETL Jobs")
	echo.Info("Connecting to database...")
	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()
	echo.Success("✓ Connected to database")

	cleared, err := database.ClearRunningETLJobs(cmd.Context(), reason)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	echo.Successf("✓ Cleared %d running ETL job(s)", cleared)
	return nil
}

func parseETLJobStatuses(raw string) ([]db.ETLJobStatus, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	out := make([]db.ETLJobStatus, 0, len(parts))
	for _, part := range parts {
		value := db.ETLJobStatus(strings.ToLower(strings.TrimSpace(part)))
		switch value {
		case db.ETLJobStatusQueued,
			db.ETLJobStatusStarted,
			db.ETLJobStatusRunning,
			db.ETLJobStatusRetryWait,
			db.ETLJobStatusSucceeded,
			db.ETLJobStatusFailed,
			db.ETLJobStatusCancelled:
			out = append(out, value)
		default:
			return nil, fmt.Errorf("error: invalid ETL status %q", part)
		}
	}

	slices.Sort(out)
	out = slices.Compact(out)
	return out, nil
}

func parseETLJobTypeFilter(raw string) (db.ETLJobType, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "", nil
	}

	value := db.ETLJobType(raw)
	switch value {
	case db.ETLJobTypeFullRun,
		db.ETLJobTypeYearlySync,
		db.ETLJobTypeValidate,
		db.ETLJobTypeCleanup,
		db.ETLJobTypeMaintenance:
		return value, nil
	default:
		return "", fmt.Errorf("error: invalid ETL job type %q", raw)
	}
}

func formatNullableTime(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return value.UTC().Format(time.RFC3339)
}

func compactError(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	const maxLen = 100
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}

func blankAsDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

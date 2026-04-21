package db

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

type testScanner struct {
	values []any
}

func (s testScanner) Scan(dest ...any) error {
	if len(dest) != len(s.values) {
		return fmt.Errorf("scan destination count mismatch: got %d want %d", len(dest), len(s.values))
	}
	for i := range dest {
		if err := assignScanValue(dest[i], s.values[i]); err != nil {
			return fmt.Errorf("assign scan value at index %d: %w", i, err)
		}
	}
	return nil
}

func assignScanValue(dest any, value any) error {
	switch d := dest.(type) {
	case *int64:
		if value == nil {
			*d = 0
			return nil
		}
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("expected int64, got %T", value)
		}
		*d = v
		return nil
	case *int:
		if value == nil {
			*d = 0
			return nil
		}
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
		*d = v
		return nil
	case *string:
		if value == nil {
			*d = ""
			return nil
		}
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		*d = v
		return nil
	case *[]byte:
		if value == nil {
			*d = nil
			return nil
		}
		v, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
		*d = v
		return nil
	case *ETLJobType:
		if value == nil {
			*d = ""
			return nil
		}
		v, ok := value.(ETLJobType)
		if !ok {
			return fmt.Errorf("expected ETLJobType, got %T", value)
		}
		*d = v
		return nil
	case *ETLJobStatus:
		if value == nil {
			*d = ""
			return nil
		}
		v, ok := value.(ETLJobStatus)
		if !ok {
			return fmt.Errorf("expected ETLJobStatus, got %T", value)
		}
		*d = v
		return nil
	case *time.Time:
		if value == nil {
			*d = time.Time{}
			return nil
		}
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
		*d = v
		return nil
	case *sql.NullString:
		if value == nil {
			d.Valid = false
			d.String = ""
			return nil
		}
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		d.String = v
		d.Valid = true
		return nil
	case *sql.NullInt64:
		if value == nil {
			d.Valid = false
			d.Int64 = 0
			return nil
		}
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("expected int64, got %T", value)
		}
		d.Int64 = v
		d.Valid = true
		return nil
	case *sql.NullTime:
		if value == nil {
			d.Valid = false
			d.Time = time.Time{}
			return nil
		}
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
		d.Time = v
		d.Valid = true
		return nil
	default:
		return fmt.Errorf("unsupported destination type %T", dest)
	}
}

func TestScanETLJob_NullFailureClassAndLastError(t *testing.T) {
	queuedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	scanner := testScanner{values: []any{
		int64(42),
		ETLJobTypeFullRun,
		1,
		ETLJobStatusQueued,
		"dev",
		"incremental",
		[]byte(`{"years":[2024,2025]}`),
		[]byte(`{"rebuild":false}`),
		3,
		1,
		nil,
		nil,
		int64(0),
		nil,
		queuedAt,
		nil,
		nil,
		nil,
		nil,
	}}

	job, err := scanETLJob(scanner)
	if err != nil {
		t.Fatalf("scanETLJob returned error: %v", err)
	}

	if job.FailureClass != "" {
		t.Fatalf("expected empty failure class, got %q", job.FailureClass)
	}
	if job.LastError != "" {
		t.Fatalf("expected empty last error, got %q", job.LastError)
	}
	if len(job.Scope) == 0 {
		t.Fatalf("expected decoded scope map")
	}
	if len(job.Options) == 0 {
		t.Fatalf("expected decoded options map")
	}
}

func TestScanETLJob_PopulatesNullableFields(t *testing.T) {
	queuedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	startedAt := queuedAt.Add(10 * time.Second)
	finishedAt := startedAt.Add(20 * time.Second)
	nextRetry := finishedAt.Add(30 * time.Second)
	scanner := testScanner{values: []any{
		int64(55),
		ETLJobTypeYearlySync,
		0,
		ETLJobStatusRetryWait,
		"prod",
		"full",
		[]byte(`{}`),
		[]byte(`{}`),
		5,
		2,
		"network",
		"temporary outage",
		int64(123),
		int64(88),
		queuedAt,
		startedAt,
		finishedAt,
		nextRetry,
		"etl-worker-1",
	}}

	job, err := scanETLJob(scanner)
	if err != nil {
		t.Fatalf("scanETLJob returned error: %v", err)
	}

	if job.FailureClass != "network" {
		t.Fatalf("expected failure class network, got %q", job.FailureClass)
	}
	if job.LastError != "temporary outage" {
		t.Fatalf("expected last error temporary outage, got %q", job.LastError)
	}
	if job.RunID == nil || *job.RunID != 88 {
		t.Fatalf("expected run ID 88, got %#v", job.RunID)
	}
	if job.WorkerID != "etl-worker-1" {
		t.Fatalf("expected worker ID etl-worker-1, got %q", job.WorkerID)
	}
	if job.StartedAt == nil || !job.StartedAt.Equal(startedAt) {
		t.Fatalf("unexpected started_at: %#v", job.StartedAt)
	}
	if job.FinishedAt == nil || !job.FinishedAt.Equal(finishedAt) {
		t.Fatalf("unexpected finished_at: %#v", job.FinishedAt)
	}
	if job.NextRetryAt == nil || !job.NextRetryAt.Equal(nextRetry) {
		t.Fatalf("unexpected next_retry_at: %#v", job.NextRetryAt)
	}
}

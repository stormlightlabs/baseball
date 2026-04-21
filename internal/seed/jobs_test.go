package seed

import (
	"errors"
	"testing"
)

func TestParseETLJobTypeAuto(t *testing.T) {
	jobType, err := ParseETLJobType("auto", PipelineModeIncremental, []int{2024})
	if err != nil {
		t.Fatalf("ParseETLJobType returned error: %v", err)
	}
	if jobType != "yearly-sync" {
		t.Fatalf("expected yearly-sync, got %s", jobType)
	}

	jobType, err = ParseETLJobType("auto", PipelineModeFull, []int{1910, 1911})
	if err != nil {
		t.Fatalf("ParseETLJobType returned error: %v", err)
	}
	if jobType != "full-run" {
		t.Fatalf("expected full-run, got %s", jobType)
	}
}

func TestParseETLJobTypeInvalid(t *testing.T) {
	if _, err := ParseETLJobType("nope", PipelineModeIncremental, nil); err == nil {
		t.Fatal("expected invalid job type error")
	}
}

func TestSplitYearsIntoBatches(t *testing.T) {
	years := []int{2021, 2022, 2023, 2024, 2025}
	batches := splitYearsIntoBatches(years, 2)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if len(batches[0]) != 2 || len(batches[1]) != 2 || len(batches[2]) != 1 {
		t.Fatalf("unexpected batch sizes: %#v", batches)
	}
}

func TestNormalizeJobWorkerOptions(t *testing.T) {
	opts := NormalizeJobWorkerOptions(JobWorkerOptions{})
	if opts.MaxActiveJobs < 1 {
		t.Fatalf("expected positive MaxActiveJobs, got %d", opts.MaxActiveJobs)
	}
	if opts.MaxQueuedJobs < 1 {
		t.Fatalf("expected positive MaxQueuedJobs, got %d", opts.MaxQueuedJobs)
	}
	if opts.NetworkRetryBackoff <= 0 {
		t.Fatalf("expected positive NetworkRetryBackoff, got %s", opts.NetworkRetryBackoff)
	}
	if opts.WorkerID == "" {
		t.Fatal("expected generated WorkerID")
	}
}

func TestClassifyPipelineFailure(t *testing.T) {
	class, retryable := classifyPipelineFailure(errors.New("extract.retrosheet: download failed after retries"))
	if class != "network" || !retryable {
		t.Fatalf("expected retryable network classification, got class=%s retryable=%v", class, retryable)
	}

	class, retryable = classifyPipelineFailure(errors.New("validation failed: [coverage] missing seasons"))
	if class != "validation" || retryable {
		t.Fatalf("expected non-retryable validation classification, got class=%s retryable=%v", class, retryable)
	}

	class, retryable = classifyPipelineFailure(errors.New("load.salary: relation does not exist"))
	if class != "db_write" || retryable {
		t.Fatalf("expected non-retryable db_write classification, got class=%s retryable=%v", class, retryable)
	}
}

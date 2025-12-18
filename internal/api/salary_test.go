package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/testutils"
)

func setupSalaryTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	ctx := context.Background()
	projectRoot, err := testutils.GetProjectRoot()
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}

	container, err := testutils.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}

	database, err := db.Connect(container.ConnStr)
	if err != nil {
		cleanup()
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(ctx); err != nil {
		cleanup()
		t.Fatalf("failed to run migrations: %v", err)
	}

	if err := container.SeedFromSQL(ctx, "salary_summary.sql"); err != nil {
		cleanup()
		t.Fatalf("failed to load salary summary fixtures: %v", err)
	}

	return NewServer(database.DB, nil), cleanup
}

func TestSalaryEndpoints(t *testing.T) {
	server, cleanup := setupSalaryTestServer(t)
	defer cleanup()

	t.Run("GET /v1/salaries/summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/salaries/summary", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp SalarySummaryResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Data) == 0 {
			t.Error("expected at least one salary summary")
		}

		if len(resp.Data) != 4 {
			t.Errorf("expected 4 salary summaries, got %d", len(resp.Data))
		}
	})

	t.Run("GET /v1/salaries/summary - verify order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/salaries/summary", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp SalarySummaryResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Data) > 1 {
			for i := 1; i < len(resp.Data); i++ {
				if resp.Data[i].Year > resp.Data[i-1].Year {
					t.Error("expected salary summaries to be in descending year order")
				}
			}
		}
	})

	t.Run("GET /v1/salaries/summary/{year}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/salaries/summary/2023", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var summary core.SalarySummary
		if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if summary.Year != 2023 {
			t.Errorf("expected year 2023, got %d", summary.Year)
		}

		if summary.Total != 4714113679.00 {
			t.Errorf("expected total 4714113679.00, got %f", summary.Total)
		}

		if summary.Average != 4920787.00 {
			t.Errorf("expected average 4920787.00, got %f", summary.Average)
		}

		if summary.Median != 1500000.00 {
			t.Errorf("expected median 1500000.00, got %f", summary.Median)
		}
	})

	t.Run("GET /v1/salaries/summary/{year} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/salaries/summary/1999", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if resp.Error == "" {
			t.Error("expected error message to be set")
		}
	})
}

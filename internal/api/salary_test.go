package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
)

func TestSalaryEndpoints(t *testing.T) {
	t.Run("GET /v1/salaries/summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/salaries/summary", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

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

		testServer.ServeHTTP(w, req)

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

		testServer.ServeHTTP(w, req)

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

		testServer.ServeHTTP(w, req)

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

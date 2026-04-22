package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStandingsEndpoint(t *testing.T) {
	t.Run("GET /v1/standings?season=2026", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/standings?season=2026", nil)
		w := httptest.NewRecorder()
		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp StandingsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Standings) == 0 {
			t.Fatal("expected standings rows")
		}
		if resp.Standings[0].Source != "current_season" {
			t.Fatalf("expected source=current_season, got %q", resp.Standings[0].Source)
		}
		if resp.LastUpdated == nil {
			t.Fatal("expected last_updated for current-season standings")
		}
	})

	t.Run("GET /v1/standings?season=2023", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/standings?season=2023", nil)
		w := httptest.NewRecorder()
		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp StandingsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Standings) == 0 {
			t.Fatal("expected historical standings rows")
		}
		if resp.Standings[0].Source != "lahman" {
			t.Fatalf("expected source=lahman, got %q", resp.Standings[0].Source)
		}
	})
}

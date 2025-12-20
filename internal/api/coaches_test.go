package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
)

func TestCoachEndpoints(t *testing.T) {
	t.Run("GET /v1/coaches", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected at least one coach")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/coaches with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 10 {
			t.Errorf("expected per_page 10, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/coaches/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/roberda07", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var coach core.Coach
		if err := json.NewDecoder(w.Body).Decode(&coach); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if coach.ID != "roberda07" {
			t.Errorf("expected coach ID roberda07, got %s", coach.ID)
		}

		if coach.LastName == nil || *coach.LastName != "Roberts" {
			t.Errorf("expected last name Roberts, got %v", coach.LastName)
		}
	})

	t.Run("GET /v1/coaches/{id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/nonexistent", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestCoachSeasonsEndpoint(t *testing.T) {
	t.Run("GET /v1/coaches/{id}/seasons", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp CoachSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.CoachID != "roberda07" {
			t.Errorf("expected coach ID roberda07, got %s", resp.CoachID)
		}

		if len(resp.Seasons) == 0 {
			t.Error("expected at least one coaching season for Dave Roberts")
		}
	})

	t.Run("GET /v1/coaches/{id}/seasons - verify season data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp CoachSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Seasons) > 0 {
			season := resp.Seasons[0]

			if season.TeamID == "" {
				t.Error("expected team_id to be present")
			}

			if season.Year == 0 {
				t.Error("expected year to be present")
			}

			if season.FirstGame == nil {
				t.Error("expected first_game to be present")
			}

			if season.LastGame == nil {
				t.Error("expected last_game to be present")
			}
		}
	})

	t.Run("GET /v1/coaches/{id}/seasons - coach with multiple seasons", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp CoachSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Seasons) < 2 {
			t.Error("expected Dave Roberts to have coached multiple seasons")
		}
	})

	t.Run("GET /v1/coaches/{id}/seasons - chronological order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/coaches/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp CoachSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Seasons) > 1 {
			for i := 1; i < len(resp.Seasons); i++ {
				if resp.Seasons[i].Year > resp.Seasons[i-1].Year {
					t.Error("expected seasons to be in descending chronological order")
				}
			}
		}
	})
}

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
)

func TestManagerEndpoints(t *testing.T) {
	t.Run("GET /v1/managers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers", nil)
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
			t.Error("expected at least one manager")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/managers with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers?page=1&per_page=10", nil)
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

	t.Run("GET /v1/managers/{manager_id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var manager core.Manager
		if err := json.NewDecoder(w.Body).Decode(&manager); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if manager.ID != "roberda07" {
			t.Errorf("expected manager ID roberda07, got %s", manager.ID)
		}

		if manager.LastName != "Roberts" {
			t.Errorf("expected last name Roberts, got %s", manager.LastName)
		}
	})

	t.Run("GET /v1/managers/{manager_id} - verify extended biodata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var manager core.Manager
		if err := json.NewDecoder(w.Body).Decode(&manager); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if manager.DebutGame == nil {
			t.Error("expected debut_game to be present for Dave Roberts")
		}

		if manager.LastGame == nil {
			t.Error("expected last_game to be present for Dave Roberts")
		}

		if manager.UseName == nil {
			t.Error("expected use_name to be present for Dave Roberts")
		}

		if manager.FullName == nil {
			t.Error("expected full_name to be present for Dave Roberts")
		}
	})

	t.Run("GET /v1/managers/{manager_id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/nonexistent", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestManagerSeasonsEndpoint(t *testing.T) {
	t.Run("GET /v1/managers/{manager_id}/seasons", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp ManagerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ManagerID != "roberda07" {
			t.Errorf("expected manager ID roberda07, got %s", resp.ManagerID)
		}

		if len(resp.Seasons) == 0 {
			t.Error("expected at least one managerial season for Dave Roberts")
		}
	})

	t.Run("GET /v1/managers/{manager_id}/seasons - verify season data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp ManagerSeasonsResponse
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

			if season.G == 0 {
				t.Error("expected games to be present")
			}

			if season.W == 0 {
				t.Error("expected wins to be present")
			}

			if season.L == 0 {
				t.Error("expected losses to be present")
			}
		}
	})

	t.Run("GET /v1/managers/{manager_id}/seasons - multiple seasons", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp ManagerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Seasons) < 2 {
			t.Error("expected Dave Roberts to have managed multiple seasons")
		}
	})

	t.Run("GET /v1/managers/{manager_id}/seasons - chronological order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp ManagerSeasonsResponse
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

	t.Run("GET /v1/managers/{manager_id}/seasons - verify win/loss calculations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/managers/roberda07/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp ManagerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Seasons) > 0 {
			season := resp.Seasons[0]

			if season.W+season.L > season.G {
				t.Errorf("wins (%d) + losses (%d) should not exceed games (%d)", season.W, season.L, season.G)
			}
		}
	})
}

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
)

func TestUmpireEndpoints(t *testing.T) {
	t.Run("GET /v1/umpires", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires", nil)
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
			t.Error("expected at least one umpire")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/umpires with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires?page=1&per_page=10", nil)
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

	t.Run("GET /v1/umpires/{umpire_id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var umpire core.Umpire
		if err := json.NewDecoder(w.Body).Decode(&umpire); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if umpire.ID != "abbof101" {
			t.Errorf("expected umpire ID abbof101, got %s", umpire.ID)
		}

		if umpire.LastName != "Abbott" {
			t.Errorf("expected last name Abbott, got %s", umpire.LastName)
		}
	})

	t.Run("GET /v1/umpires/{umpire_id} - verify biodata fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var umpire core.Umpire
		if err := json.NewDecoder(w.Body).Decode(&umpire); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if umpire.FirstGame == nil {
			t.Error("expected first_game to be present")
		}

		if umpire.LastGame == nil {
			t.Error("expected last_game to be present")
		}
	})

	t.Run("GET /v1/umpires/{umpire_id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/nonexistent", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestUmpireGamesEndpoint(t *testing.T) {
	t.Run("GET /v1/umpires/{umpire_id}/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101/games", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/umpires/{umpire_id}/games with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101/games?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/umpires/{umpire_id}/games with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101/games?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 5 {
			t.Errorf("expected per_page 5, got %d", resp.PerPage)
		}
	})
}

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestGameEndpoints(t *testing.T) {
	t.Run("GET /v1/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games", nil)
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
			t.Error("expected at least one game")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/games with date filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games?date=2023-04-01", nil)
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

	t.Run("GET /v1/games with team filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games?team_id=NYA", nil)
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

	t.Run("GET /v1/games/{id}", func(t *testing.T) {
		gameID := "NYA202309100"
		req := httptest.NewRequest(http.MethodGet, "/v1/games/"+gameID, nil)
		req.SetPathValue("id", gameID)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/games/{id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/INVALID999", nil)
		req.SetPathValue("id", "INVALID999")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("GET /v1/games/{id}/boxscore", func(t *testing.T) {
		gameID := "NYA202309100"
		req := httptest.NewRequest(http.MethodGet, "/v1/games/"+gameID+"/boxscore", nil)
		req.SetPathValue("id", gameID)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/schedule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/schedule", nil)
		req.SetPathValue("year", "2023")
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

	t.Run("GET /v1/games/date/{date}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/date/2023-09-10", nil)
		req.SetPathValue("date", "2023-09-10")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/teams/{id}/{year}/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams/NYA/2023/games", nil)
		req.SetPathValue("id", "NYA")
		req.SetPathValue("year", "2023")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

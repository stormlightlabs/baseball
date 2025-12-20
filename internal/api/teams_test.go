package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestTeamEndpoints(t *testing.T) {
	t.Run("GET /v1/teams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams", nil)
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
			t.Error("expected at least one team")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/teams with year filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams?year=2023", nil)
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

	t.Run("GET /v1/teams/{id}?year=2023", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams/NYA?year=2023", nil)
		req.SetPathValue("id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/teams/{id}?year=1800 - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams/ZZZ?year=1800", nil)
		req.SetPathValue("id", "ZZZ")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams", nil)
		req.SetPathValue("year", "2023")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/franchises", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/franchises", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/franchises/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/franchises/NYY", nil)
		req.SetPathValue("id", "NYY")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/teams?year=2023", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams?year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/roster", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/roster", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("team_id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/batting", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("team_id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/pitching", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("team_id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/fielding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/fielding", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("team_id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/schedule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/schedule", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("team_id", "NYA")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

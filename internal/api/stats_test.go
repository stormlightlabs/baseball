package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestStatsEndpoints(t *testing.T) {
	t.Run("GET /v1/stats/batting/leaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/leaders?stat=hr&year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting/leaders with min plate appearances", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/leaders?stat=avg&year=2023&min_pa=100", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/pitching/leaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/pitching/leaders?stat=era&year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/pitching/leaders with min innings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/pitching/leaders?stat=era&year=2023&min_ip=50", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting/query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/query?year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/pitching/query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/pitching/query?year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/fielding/query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/fielding/query?year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting/career/leaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/career/leaders?stat=hr", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/pitching/career/leaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/pitching/career/leaders?stat=w", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/teams/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/teams/batting?year=2023", nil)
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

	t.Run("GET /v1/stats/teams/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/teams/pitching?year=2023", nil)
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

	t.Run("GET /v1/stats/teams/fielding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/teams/fielding?year=2023", nil)
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

	t.Run("GET /v1/stats/batting/query with league filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/query?year=2023&league=AL", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting/query with team filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting/query?year=2023&team_id=NYA", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

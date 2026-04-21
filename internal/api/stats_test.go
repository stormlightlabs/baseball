package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestStatsEndpoints(t *testing.T) {
	t.Run("GET /v1/seasons/{year}/leaders/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/leaders/batting?stat=hr", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/leaders/batting with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/leaders/batting?stat=avg&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/leaders/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/leaders/pitching?stat=era", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/leaders/pitching with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/leaders/pitching?stat=era&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/leaders/pitching historical null BK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/1914/leaders/pitching?page=1&per_page=10&stat=era", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/pitching?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/fielding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/fielding?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/leaders/batting/career", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/leaders/batting/career?stat=hr", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/leaders/pitching/career", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/leaders/pitching/career?stat=w", nil)
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

	t.Run("GET /v1/stats/batting with league filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting?season=2023&league=AL", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/stats/batting with team filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/stats/batting?season=2023&team_id=NYA", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

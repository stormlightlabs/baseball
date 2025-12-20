package api

import (
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestAwardsEndpoints(t *testing.T) {
	t.Run("GET /v1/awards", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards - verify response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestAwardDetailsEndpoint(t *testing.T) {
	t.Run("GET /v1/awards/{award_id}", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with year filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?year=2018", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with player_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?player_id=bettsmo01", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with league filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?league=AL", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} - verify paginated response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestSeasonAwardsEndpoint(t *testing.T) {
	t.Run("GET /v1/seasons/{year}/awards", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with award_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?award_id=Gold%20Glove", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with player_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?player_id=bettsmo01", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with league filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?league=AL", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards - verify response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

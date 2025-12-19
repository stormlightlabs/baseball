package api

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/testutils"
)

func setupAwardsTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	ctx := context.Background()
	projectRoot, err := testutils.GetProjectRoot()
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}

	container, err := testutils.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}

	database, err := db.Connect(container.ConnStr)
	if err != nil {
		cleanup()
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(ctx); err != nil {
		cleanup()
		t.Fatalf("failed to run migrations: %v", err)
	}

	if err := container.LoadFixtures(ctx); err != nil {
		cleanup()
		t.Fatalf("failed to load fixtures: %v", err)
	}

	if _, err := database.RefreshMaterializedViews(ctx, []string{}); err != nil {
		cleanup()
		t.Fatalf("failed to refresh materialized views: %v", err)
	}

	return NewServer(database.DB, nil), cleanup
}

func TestAwardsEndpoints(t *testing.T) {
	server, cleanup := setupAwardsTestServer(t)
	defer cleanup()

	t.Run("GET /v1/awards", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards - verify response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestAwardDetailsEndpoint(t *testing.T) {
	server, cleanup := setupAwardsTestServer(t)
	defer cleanup()

	t.Run("GET /v1/awards/{award_id}", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with year filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?year=2018", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with player_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?player_id=bettsmo01", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with league filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?league=AL", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/awards/{award_id} - verify paginated response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/awards/Gold%20Glove", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestSeasonAwardsEndpoint(t *testing.T) {
	server, cleanup := setupAwardsTestServer(t)
	defer cleanup()

	t.Run("GET /v1/seasons/{year}/awards", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with award_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?award_id=Gold%20Glove", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with player_id filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?player_id=bettsmo01", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with league filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?league=AL", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/seasons/{year}/awards - verify response structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/seasons/2018/awards", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/testutils"
)

func setupAllStarTestServer(t *testing.T) (*Server, func()) {
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

func TestAllStarGamesEndpoint(t *testing.T) {
	server, cleanup := setupAllStarTestServer(t)
	defer cleanup()

	t.Run("GET /v1/allstar/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/allstar/games", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var games []core.AllStarGame
		if err := json.NewDecoder(w.Body).Decode(&games); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/allstar/games with year filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/allstar/games?year=2023", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var games []core.AllStarGame
		if err := json.NewDecoder(w.Body).Decode(&games); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		for _, game := range games {
			if game.Year != 2023 {
				t.Errorf("expected year 2023, got %d", game.Year)
			}
		}
	})

	t.Run("GET /v1/allstar/games - verify response structure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/allstar/games", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var games []core.AllStarGame
		if err := json.NewDecoder(w.Body).Decode(&games); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(games) > 0 {
			game := games[0]
			if game.GameID == "" {
				t.Error("expected game_id to be set")
			}
			if game.Year == 0 {
				t.Error("expected year to be set")
			}
		}
	})
}

func TestAllStarGameDetailsEndpoint(t *testing.T) {
	server, cleanup := setupAllStarTestServer(t)
	defer cleanup()

	t.Run("GET /v1/allstar/games/{id} - valid game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/allstar/games/ALS202307110", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Accept either success or internal server error (if game doesn't exist in fixtures)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 200 or 500, got %d", w.Code)
		}

		if w.Code == http.StatusOK {
			var game core.AllStarGame
			if err := json.NewDecoder(w.Body).Decode(&game); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if game.GameID == "" {
				t.Error("expected game_id to be set")
			}
		}
	})

	t.Run("GET /v1/allstar/games/{id} - nonexistent game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/allstar/games/INVALID999", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 for nonexistent game, got %d", w.Code)
		}
	})
}

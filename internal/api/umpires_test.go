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

func setupUmpireTestServer(t *testing.T) (*Server, func()) {
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
		t.Fatalf("failed to load umpire fixtures: %v", err)
	}

	if _, err := database.RefreshMaterializedViews(ctx, []string{}); err != nil {
		cleanup()
		t.Fatalf("failed to refresh materialized views: %v", err)
	}

	return NewServer(database.DB, nil), cleanup
}

func TestUmpireEndpoints(t *testing.T) {
	server, cleanup := setupUmpireTestServer(t)
	defer cleanup()

	t.Run("GET /v1/umpires", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

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

		server.ServeHTTP(w, req)

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

		server.ServeHTTP(w, req)

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

		server.ServeHTTP(w, req)

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

		server.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestUmpireGamesEndpoint(t *testing.T) {
	server, cleanup := setupUmpireTestServer(t)
	defer cleanup()

	t.Run("GET /v1/umpires/{umpire_id}/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101/games", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

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

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/umpires/{umpire_id}/games with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/umpires/abbof101/games?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

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

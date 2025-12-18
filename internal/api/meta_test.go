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

func setupTestServer(t *testing.T) (*Server, func()) {
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

	if err := container.SeedFromSQL(ctx, "woba_constants.sql", "league_constants.sql", "park_factors.sql"); err != nil {
		cleanup()
		t.Fatalf("failed to seed constants: %v", err)
	}

	return NewServer(database.DB, nil), cleanup
}

func TestMetaEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("GET /v1/meta", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp metaResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Version == "" {
			t.Error("expected version to be set")
		}

		if resp.GeneratedAt.IsZero() {
			t.Error("expected generated_at to be set")
		}

		if resp.Coverage == nil {
			t.Error("expected coverage to be set")
		}

		if resp.SchemaHashes == nil {
			t.Error("expected schema_hashes to be set")
		}

		if resp.Datasets == nil {
			t.Error("expected datasets to be set")
		}
	})

	t.Run("GET /v1/meta/datasets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/datasets", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var datasets []core.DatasetStatus
		if err := json.NewDecoder(w.Body).Decode(&datasets); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if datasets == nil {
			t.Error("expected datasets array, got nil")
		}
	})

	t.Run("GET /v1/meta/constants/woba", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/woba", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.WOBAConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) == 0 {
			t.Error("expected at least one wOBA constant")
		}

		for _, c := range constants {
			if c.Season == 0 {
				t.Error("expected season to be set")
			}
			if c.WOBA == 0 {
				t.Error("expected woba to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/woba with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/woba?season=2023", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.WOBAConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) != 1 {
			t.Errorf("expected 1 constant, got %d", len(constants))
		}

		if len(constants) > 0 && constants[0].Season != 2023 {
			t.Errorf("expected season 2023, got %d", constants[0].Season)
		}
	})

	t.Run("GET /v1/meta/constants/league", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/league", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.LeagueConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) == 0 {
			t.Error("expected at least one league constant")
		}

		for _, c := range constants {
			if c.Season == 0 {
				t.Error("expected season to be set")
			}
			if c.League == "" {
				t.Error("expected league to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/league with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/league?season=2024&league=AL", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.LeagueConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) != 1 {
			t.Errorf("expected 1 constant, got %d", len(constants))
		}

		if len(constants) > 0 {
			if constants[0].Season != 2024 {
				t.Errorf("expected season 2024, got %d", constants[0].Season)
			}
			if constants[0].League != "AL" {
				t.Errorf("expected league AL, got %s", constants[0].League)
			}
		}
	})

	t.Run("GET /v1/meta/constants/park-factors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/park-factors", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var factors []core.ParkFactorRow
		if err := json.NewDecoder(w.Body).Decode(&factors); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(factors) == 0 {
			t.Error("expected at least one park factor")
		}

		for _, f := range factors {
			if f.Season == 0 {
				t.Error("expected season to be set")
			}
			if f.ParkID == "" {
				t.Error("expected park_id to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/park-factors with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/park-factors?season=2023&team=BOS", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var factors []core.ParkFactorRow
		if err := json.NewDecoder(w.Body).Decode(&factors); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(factors) == 0 {
			t.Error("expected at least one park factor for BOS in 2023")
		}

		for _, f := range factors {
			if f.Season != 2023 {
				t.Errorf("expected season 2023, got %d", f.Season)
			}
			if f.TeamID != nil && *f.TeamID != "BOS" {
				t.Errorf("expected team BOS, got %s", *f.TeamID)
			}
		}
	})
}

func TestHealthEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("GET /v1/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("expected status 'ok', got '%s'", resp.Status)
		}
	})
}

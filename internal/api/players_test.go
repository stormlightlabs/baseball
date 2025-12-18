package api

import (
	"context"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/testutils"
)

func setupPlayerTestServer(t *testing.T) (*Server, func()) {
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

	// TODO: Load player fixture data
	// if err := container.SeedFromSQL(ctx, "players.sql", "batting.sql", "pitching.sql"); err != nil {
	// 	cleanup()
	// 	t.Fatalf("failed to seed player data: %v", err)
	// }

	return NewServer(database.DB, nil), cleanup
}

func TestPlayerEndpoints(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players with name filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players with debut_year filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players with pagination", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id} - not found", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})
}

func TestPlayerSeasonsEndpoint(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/seasons", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/seasons - player with both batting and pitching", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/seasons - player with only batting", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/seasons - player with only pitching", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})
}

func TestPlayerStatsEndpoints(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/stats/batting", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/stats/batting - verify career totals", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/stats/batting - verify calculated rates", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/stats/pitching", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/stats/pitching - verify career totals", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})

	t.Run("GET /v1/players/{id}/stats/pitching - verify calculated rates", func(t *testing.T) {
		t.Skip("TODO: implement after loading player fixtures")
	})
}

func TestPlayerAwardsEndpoints(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/awards", func(t *testing.T) {
		t.Skip("TODO: implement after loading award fixtures")
	})

	t.Run("GET /v1/players/{id}/awards with year filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading award fixtures")
	})

	t.Run("GET /v1/players/{id}/awards with pagination", func(t *testing.T) {
		t.Skip("TODO: implement after loading award fixtures")
	})

	t.Run("GET /v1/players/{id}/hall-of-fame", func(t *testing.T) {
		t.Skip("TODO: implement after loading hall of fame fixtures")
	})

	t.Run("GET /v1/players/{id}/hall-of-fame - player inducted", func(t *testing.T) {
		t.Skip("TODO: implement after loading hall of fame fixtures")
	})

	t.Run("GET /v1/players/{id}/hall-of-fame - player not inducted", func(t *testing.T) {
		t.Skip("TODO: implement after loading hall of fame fixtures")
	})
}

func TestPlayerGameLogsEndpoints(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/game-logs", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs with season filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs with pagination", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/batting", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with season filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with date range", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with min_hr filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with multiple stat filters", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with season filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with min_so filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with min_ip filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding with position filter", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding with date range", func(t *testing.T) {
		t.Skip("TODO: implement after loading game log fixtures")
	})
}

func TestPlayerAppearancesEndpoint(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/appearances", func(t *testing.T) {
		t.Skip("TODO: implement after loading appearance fixtures")
	})

	t.Run("GET /v1/players/{id}/appearances - multi-position player", func(t *testing.T) {
		t.Skip("TODO: implement after loading appearance fixtures")
	})
}

func TestPlayerTeamsEndpoint(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/teams", func(t *testing.T) {
		t.Skip("TODO: implement after loading team history fixtures")
	})

	t.Run("GET /v1/players/{id}/teams - player with multiple teams", func(t *testing.T) {
		t.Skip("TODO: implement after loading team history fixtures")
	})

	t.Run("GET /v1/players/{id}/teams - player with single team career", func(t *testing.T) {
		t.Skip("TODO: implement after loading team history fixtures")
	})
}

func TestPlayerSalariesEndpoint(t *testing.T) {
	_, cleanup := setupPlayerTestServer(t)
	defer cleanup()

	t.Run("GET /v1/players/{id}/salaries", func(t *testing.T) {
		t.Skip("TODO: implement after loading salary fixtures")
	})

	t.Run("GET /v1/players/{id}/salaries - verify chronological order", func(t *testing.T) {
		t.Skip("TODO: implement after loading salary fixtures")
	})
}

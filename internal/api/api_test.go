package api

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/testutils"
)

var (
	testServer  *Server
	testDB      *sql.DB
	testCleanup func()
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	projectRoot, err := testutils.GetProjectRoot()
	if err != nil {
		panic("failed to get project root: " + err.Error())
	}

	originalDir, err := os.Getwd()
	if err != nil {
		panic("failed to get current directory: " + err.Error())
	}

	if err := os.Chdir(projectRoot); err != nil {
		panic("failed to change to project root: " + err.Error())
	}

	container, err := testutils.NewPostgresContainer(ctx)
	if err != nil {
		panic("failed to create postgres container: " + err.Error())
	}

	testCleanup = func() {
		os.Chdir(originalDir)
		if err := container.Terminate(ctx); err != nil {
			panic("failed to terminate container: " + err.Error())
		}
	}

	database, err := db.Connect(container.ConnStr)
	if err != nil {
		testCleanup()
		panic("failed to connect to database: " + err.Error())
	}

	if err := database.Migrate(ctx); err != nil {
		testCleanup()
		panic("failed to run migrations: " + err.Error())
	}

	if err := container.LoadFixtures(ctx); err != nil {
		testCleanup()
		panic("failed to load fixtures: " + err.Error())
	}

	sqlFiles := []string{"woba_constants.sql", "league_constants.sql", "park_factors.sql", "salary_summary.sql"}
	if err := container.SeedFromSQL(ctx, sqlFiles...); err != nil {
		testCleanup()
		panic("failed to load SQL fixtures: " + err.Error())
	}

	if _, err := database.RefreshMaterializedViews(ctx, []string{}); err != nil {
		testCleanup()
		panic("failed to refresh materialized views: " + err.Error())
	}

	testDB = database.DB
	testServer = NewServer(database.DB, nil)

	code := m.Run()

	testCleanup()

	os.Exit(code)
}

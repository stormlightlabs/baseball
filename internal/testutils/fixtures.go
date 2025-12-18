package testutils

import (
	"context"
	"fmt"
	"path/filepath"
)

// LoadFixtures loads all CSV fixtures into the test database.
// It uses PostgreSQL's COPY command for efficient bulk loading.
func (c *PostgresContainer) LoadFixtures(ctx context.Context) error {
	fixtures := map[string]string{
		`"Teams"`:    "teams.csv",
		`"People"`:   "people.csv",
		`"Batting"`:  "batting.csv",
		`"Pitching"`: "pitching.csv",
		`"Fielding"`: "fielding.csv",
		`"Parks"`:    "parks.csv",
		"games":      "games.csv",
		"plays":      "plays.csv",
	}

	projectRoot, err := GetProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	fixturesDir := filepath.Join(projectRoot, "internal", "testutils", "testdata")

	for table, csvFile := range fixtures {
		csvPath := filepath.Join(fixturesDir, csvFile)
		if err := c.LoadCSV(ctx, table, csvPath); err != nil {
			return fmt.Errorf("failed to load %s: %w", csvFile, err)
		}
	}

	return nil
}

// LoadCSV loads a CSV file into a database table using COPY command.
// TODO: Implement efficient CSV loading based on ETL pipeline implementation.
func (c *PostgresContainer) LoadCSV(ctx context.Context, table, csvPath string) error {
	if !fileExists(csvPath) {
		return fmt.Errorf("CSV file not found: %s", csvPath)
	}

	query := fmt.Sprintf("COPY %s FROM STDIN WITH (FORMAT csv, HEADER true)", table)
	_ = query
	return fmt.Errorf("LoadCSV not yet implemented - use RunMigrations and Seed for now")
}

// LoadLeagueFixtures loads fixtures for specific leagues.
func (c *PostgresContainer) LoadLeagueFixtures(ctx context.Context, leagues ...string) error {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	fixturesDir := filepath.Join(projectRoot, "internal", "testutils", "testdata")

	for _, league := range leagues {
		var csvFile string
		switch league {
		case "FL":
			csvFile = "games_federal_league.csv"
			if err := c.LoadCSV(ctx, "games", filepath.Join(fixturesDir, csvFile)); err != nil {
				return fmt.Errorf("failed to load Federal League games: %w", err)
			}
			csvFile = "plays_federal_league.csv"
			if err := c.LoadCSV(ctx, "plays", filepath.Join(fixturesDir, csvFile)); err != nil {
				return fmt.Errorf("failed to load Federal League plays: %w", err)
			}
		case "NNL", "NAL":
			csvFile = "games_negro_leagues.csv"
			if err := c.LoadCSV(ctx, "games", filepath.Join(fixturesDir, csvFile)); err != nil {
				return fmt.Errorf("failed to load Negro Leagues games: %w", err)
			}
		default:
			return fmt.Errorf("unknown league: %s", league)
		}
	}

	return nil
}

// SeedFromSQL loads fixture data from SQL files instead of CSV.
func (c *PostgresContainer) SeedFromSQL(ctx context.Context, sqlFiles ...string) error {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	fixturesDir := filepath.Join(projectRoot, "internal", "testutils", "testdata")

	for _, sqlFile := range sqlFiles {
		sqlPath := filepath.Join(fixturesDir, sqlFile)
		if err := c.Seed(ctx, sqlPath); err != nil {
			return fmt.Errorf("failed to seed from %s: %w", sqlFile, err)
		}
	}

	return nil
}

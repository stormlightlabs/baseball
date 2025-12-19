package testutils

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadFixtures loads all CSV fixtures into the test database.
// It uses PostgreSQL's COPY command for efficient bulk loading.
func (c *PostgresContainer) LoadFixtures(ctx context.Context) error {
	fixtures := map[string]string{
		`"Teams"`:         "teams.csv",
		`"People"`:        "people.csv",
		`"Batting"`:       "batting.csv",
		`"Pitching"`:      "pitching.csv",
		`"Fielding"`:      "fielding.csv",
		`"Parks"`:         "parks.csv",
		`"Appearances"`:   "appearances.csv",
		`"AwardsPlayers"`: "awards_players.csv",
		`"HallOfFame"`:    "hall_of_fame.csv",
		`"Salaries"`:      "salaries.csv",
		`"AllstarFull"`:   "allstar_full.csv",
		"games":           "games.csv",
		"plays":           "plays.csv",
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

// LoadCSV loads a CSV file into a database table using INSERT statements.
// This is simpler than COPY FROM STDIN for small test datasets.
func (c *PostgresContainer) LoadCSV(ctx context.Context, table, csvPath string) error {
	if !fileExists(csvPath) {
		return fmt.Errorf("CSV file not found: %s", csvPath)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	generatedColumns := map[string]bool{
		"game_id": true,
	}

	validIndices := []int{}
	validHeaders := []string{}
	for i, header := range headers {
		if !generatedColumns[header] {
			validIndices = append(validIndices, i)
			validHeaders = append(validHeaders, header)
		}
	}

	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	placeholders := make([]string, len(validHeaders))
	quotedHeaders := make([]string, len(validHeaders))
	for i := range validHeaders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		quotedHeaders[i] = fmt.Sprintf(`"%s"`, validHeaders[i])
	}

	columnList := strings.Join(quotedHeaders, ", ")
	placeholderList := strings.Join(placeholders, ", ")
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnList, placeholderList)

	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare INSERT statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		values := make([]any, len(validIndices))
		for i, idx := range validIndices {
			v := record[idx]
			if v == "" {
				values[i] = nil
			} else {
				values[i] = v
			}
		}

		_, err = stmt.ExecContext(ctx, values...)
		if err != nil {
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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

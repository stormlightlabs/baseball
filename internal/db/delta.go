package db

import (
	"context"
	"fmt"
	"sort"
)

func (db *DB) UpsertDeltaSeasonsForRun(ctx context.Context, runID int64, years []int) error {
	if runID <= 0 || len(years) == 0 {
		return nil
	}

	years = dedupeYears(years)
	_, err := db.ExecContext(ctx, `
		INSERT INTO etl_delta_seasons (run_id, season)
		SELECT $1::bigint, y
		FROM UNNEST($2::int[]) AS y
		ON CONFLICT (run_id, season) DO NOTHING
	`, runID, years)
	if err != nil {
		return fmt.Errorf("failed to upsert changed seasons: %w", err)
	}
	return nil
}

func (db *DB) UpsertDeltaGamesForRun(ctx context.Context, runID int64, years []int) error {
	if runID <= 0 || len(years) == 0 {
		return nil
	}

	years = dedupeYears(years)
	_, err := db.ExecContext(ctx, `
		INSERT INTO etl_delta_games (run_id, game_id, season)
		SELECT DISTINCT
			$1::bigint AS run_id,
			g.game_id,
			CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) AS season
		FROM games g
		WHERE CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) = ANY($2::int[])
		ON CONFLICT (run_id, game_id) DO NOTHING
	`, runID, years)
	if err != nil {
		return fmt.Errorf("failed to upsert changed games: %w", err)
	}
	return nil
}

func (db *DB) UpsertDeltaPlayersForRun(ctx context.Context, runID int64, years []int) error {
	if runID <= 0 || len(years) == 0 {
		return nil
	}

	years = dedupeYears(years)
	_, err := db.ExecContext(ctx, `
		WITH scoped_games AS (
			SELECT
				g.game_id,
				CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) AS season
			FROM games g
			WHERE CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) = ANY($2::int[])
		),
		player_scope AS (
			SELECT DISTINCT
				sg.season,
				v.player_id
			FROM plays p
			JOIN scoped_games sg ON sg.game_id = p.gid
			CROSS JOIN LATERAL (
				VALUES (p.batter), (p.pitcher), (p.f2), (p.f3), (p.f4), (p.f5), (p.f6), (p.f7), (p.f8), (p.f9)
			) AS v(player_id)
			WHERE COALESCE(v.player_id, '') <> ''
		)
		INSERT INTO etl_delta_players (run_id, player_id, season)
		SELECT $1::bigint, player_id, season
		FROM player_scope
		ON CONFLICT (run_id, player_id) DO NOTHING
	`, runID, years)
	if err != nil {
		return fmt.Errorf("failed to upsert changed players: %w", err)
	}
	return nil
}

func (db *DB) DeltaPlayersForRun(ctx context.Context, runID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT player_id
		FROM etl_delta_players
		WHERE run_id = $1
		ORDER BY player_id ASC
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to query changed players for run %d: %w", runID, err)
	}
	defer rows.Close()

	players := make([]string, 0)
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, fmt.Errorf("failed to scan changed player id: %w", err)
		}
		players = append(players, playerID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating changed players rows: %w", err)
	}

	return players, nil
}

func dedupeYears(years []int) []int {
	if len(years) == 0 {
		return nil
	}

	clean := make([]int, 0, len(years))
	for _, year := range years {
		if year > 0 {
			clean = append(clean, year)
		}
	}
	sort.Ints(clean)

	result := make([]int, 0, len(clean))
	last := 0
	for idx, year := range clean {
		if idx == 0 || year != last {
			result = append(result, year)
			last = year
		}
	}
	return result
}

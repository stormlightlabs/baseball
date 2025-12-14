package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/win_expectancy_get.sql
var winExpectancyGetQuery string

//go:embed queries/win_expectancy_get_era.sql
var winExpectancyGetEraQuery string

//go:embed queries/win_expectancy_list_eras.sql
var winExpectancyListErasQuery string

//go:embed queries/win_expectancy_batch_get.sql
var winExpectancyBatchGetQuery string

type WinExpectancyRepository struct {
	db *sql.DB
}

func NewWinExpectancyRepository(db *sql.DB) *WinExpectancyRepository {
	return &WinExpectancyRepository{db: db}
}

// GetWinExpectancy returns the win probability for a specific game state.
// Uses the most recent era available if era parameters are not specified.
func (r *WinExpectancyRepository) GetWinExpectancy(ctx context.Context, state core.GameState) (*core.WinExpectancy, error) {
	inning := min(state.Inning, 9)

	scoreDiff := state.ScoreDiff
	if scoreDiff > 11 {
		scoreDiff = 11
	} else if scoreDiff < -11 {
		scoreDiff = -11
	}

	var we core.WinExpectancy
	var startYear, endYear sql.NullInt64

	err := r.db.QueryRowContext(
		ctx,
		winExpectancyGetQuery,
		inning,
		state.IsBottom,
		state.Outs,
		state.RunnersCode,
		scoreDiff,
	).Scan(
		&we.ID, &we.Inning, &we.IsBottom,
		&we.Outs, &we.RunnersState,
		&we.ScoreDiff, &we.WinProbability, &we.SampleSize,
		&startYear, &endYear,
		&we.CreatedAt, &we.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no win expectancy data for state (inning=%d, bottom=%v, outs=%d, runners=%s, diff=%d)",
			inning, state.IsBottom, state.Outs, state.RunnersCode, scoreDiff)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get win expectancy: %w", err)
	}

	if startYear.Valid {
		year := int(startYear.Int64)
		we.StartYear = &year
	}
	if endYear.Valid {
		year := int(endYear.Int64)
		we.EndYear = &year
	}

	return &we, nil
}

// GetWinExpectancyForEra returns win probability for a specific game state within a time period.
func (r *WinExpectancyRepository) GetWinExpectancyForEra(ctx context.Context, state core.GameState, startYear, endYear *int) (*core.WinExpectancy, error) {
	inning := min(state.Inning, 9)
	scoreDiff := state.ScoreDiff
	if scoreDiff > 11 {
		scoreDiff = 11
	} else if scoreDiff < -11 {
		scoreDiff = -11
	}

	var we core.WinExpectancy
	var startYearDB, endYearDB sql.NullInt64

	err := r.db.QueryRowContext(
		ctx,
		winExpectancyGetEraQuery,
		inning,
		state.IsBottom,
		state.Outs,
		state.RunnersCode,
		scoreDiff,
		startYear,
		endYear,
	).Scan(
		&we.ID,
		&we.Inning, &we.IsBottom, &we.Outs, &we.RunnersState,
		&we.ScoreDiff,
		&we.WinProbability, &we.SampleSize,
		&startYearDB, &endYearDB,
		&we.CreatedAt, &we.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no win expectancy data for state in era %v-%v", startYear, endYear)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get win expectancy for era: %w", err)
	}

	if startYearDB.Valid {
		year := int(startYearDB.Int64)
		we.StartYear = &year
	}
	if endYearDB.Valid {
		year := int(endYearDB.Int64)
		we.EndYear = &year
	}

	return &we, nil
}

// BatchGetWinExpectancy efficiently retrieves win expectancies for multiple game states.
// Useful for computing leverage index across a full game.
// Uses a single database query with UNNEST for optimal performance.
func (r *WinExpectancyRepository) BatchGetWinExpectancy(ctx context.Context, states []core.GameState) ([]core.WinExpectancy, error) {
	if len(states) == 0 {
		return []core.WinExpectancy{}, nil
	}

	innings := make([]int, len(states))
	isBottoms := make([]bool, len(states))
	outs := make([]int, len(states))
	runnersStates := make([]string, len(states))
	scoreDiffs := make([]int, len(states))

	for i, state := range states {
		innings[i] = min(state.Inning, 9)
		isBottoms[i] = state.IsBottom
		outs[i] = state.Outs
		runnersStates[i] = state.RunnersCode

		scoreDiff := state.ScoreDiff
		if scoreDiff > 11 {
			scoreDiff = 11
		} else if scoreDiff < -11 {
			scoreDiff = -11
		}
		scoreDiffs[i] = scoreDiff
	}

	rows, err := r.db.QueryContext(
		ctx, winExpectancyBatchGetQuery, innings,
		isBottoms, outs, runnersStates, scoreDiffs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get win expectancy: %w", err)
	}
	defer rows.Close()

	resultsMap := make(map[int]*core.WinExpectancy)

	for rows.Next() {
		var rowNum int
		var we core.WinExpectancy
		var id sql.NullInt64
		var startYear, endYear sql.NullInt64
		var createdAt, updatedAt sql.NullTime

		err = rows.Scan(
			&rowNum,
			&id,
			&we.Inning,
			&we.IsBottom,
			&we.Outs,
			&we.RunnersState,
			&we.ScoreDiff,
			&we.WinProbability,
			&we.SampleSize,
			&startYear,
			&endYear,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch win expectancy: %w", err)
		}

		if id.Valid {
			we.ID = int(id.Int64)
			if startYear.Valid {
				year := int(startYear.Int64)
				we.StartYear = &year
			}
			if endYear.Valid {
				year := int(endYear.Int64)
				we.EndYear = &year
			}
			if createdAt.Valid {
				we.CreatedAt = createdAt.Time
			}
			if updatedAt.Valid {
				we.UpdatedAt = updatedAt.Time
			}
		}

		resultsMap[rowNum] = &we
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating batch win expectancy: %w", err)
	}

	results := make([]core.WinExpectancy, len(states))
	for i, state := range states {
		// NOTE: SQL ROW_NUMBER() starts at 1
		rowNum := i + 1
		if we, found := resultsMap[rowNum]; found && we.ID != 0 {
			results[i] = *we
		} else {
			results[i] = core.WinExpectancy{
				Inning:         state.Inning,
				IsBottom:       state.IsBottom,
				Outs:           state.Outs,
				RunnersState:   state.RunnersCode,
				ScoreDiff:      state.ScoreDiff,
				WinProbability: 0.5,
				SampleSize:     0,
			}
		}
	}

	return results, nil
}

// ListAvailableEras returns all available historical eras in the win expectancy table.
func (r *WinExpectancyRepository) ListAvailableEras(ctx context.Context) ([]core.WinExpectancyEra, error) {
	rows, err := r.db.QueryContext(ctx, winExpectancyListErasQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list eras: %w", err)
	}
	defer rows.Close()

	var eras []core.WinExpectancyEra

	for rows.Next() {
		var era core.WinExpectancyEra
		var startYear, endYear sql.NullInt64
		var totalSample int64

		err = rows.Scan(
			&startYear,
			&endYear,
			&era.Label,
			&era.StateCount,
			&totalSample,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan era: %w", err)
		}

		if startYear.Valid {
			era.StartYear = int(startYear.Int64)
		}
		if endYear.Valid {
			era.EndYear = int(endYear.Int64)
		}
		era.TotalSample = totalSample

		eras = append(eras, era)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating eras: %w", err)
	}

	return eras, nil
}

// UpsertWinExpectancy inserts or updates win expectancy data.
// Used for populating the table from historical analysis.
func (r *WinExpectancyRepository) UpsertWinExpectancy(ctx context.Context, we *core.WinExpectancy) error {
	query := `
		INSERT INTO win_expectancy_historical (
			inning, is_bottom, outs, runners_state, score_diff,
			win_probability, sample_size, start_year, end_year,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (inning, is_bottom, outs, runners_state, score_diff, start_year, end_year)
		DO UPDATE SET
			win_probability = EXCLUDED.win_probability,
			sample_size = EXCLUDED.sample_size,
			updated_at = NOW()
		RETURNING id
	`

	if err := r.db.QueryRowContext(
		ctx, query,
		we.Inning, we.IsBottom, we.Outs,
		we.RunnersState, we.ScoreDiff, we.WinProbability,
		we.SampleSize, we.StartYear, we.EndYear,
	).Scan(&we.ID); err != nil {
		return fmt.Errorf("failed to upsert win expectancy: %w", err)
	}

	return nil
}

// BuildFromHistoricalData is implemented in the db package
func (r *WinExpectancyRepository) BuildFromHistoricalData(_ context.Context, _, _, _ int) (int64, error) {
	return 0, fmt.Errorf("BuildFromHistoricalData should be called via db.BuildWinExpectancy")
}

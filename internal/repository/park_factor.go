package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

// ParkFactorRepository computes and retrieves park factors.
type ParkFactorRepository struct {
	db *sql.DB
}

// NewParkFactorRepository creates a new ParkFactorRepository.
func NewParkFactorRepository(db *sql.DB) *ParkFactorRepository {
	return &ParkFactorRepository{db: db}
}

// ParkFactor returns park factors for a specific park and season.
func (r *ParkFactorRepository) ParkFactor(ctx context.Context, parkID core.ParkID, season core.SeasonYear) (*core.ParkFactor, error) {
	row := r.db.QueryRowContext(ctx, parkFactorQuery, string(parkID), int(season))

	var pf core.ParkFactor
	var homeTeam string

	err := row.Scan(
		&pf.ParkID,
		&homeTeam,
		&pf.Season,
		&pf.GamesSampled,
		&pf.RunsFactor,
		&pf.HRFactor,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("park factor", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query park factor: %w", err)
	}

	pf.Provider = "internal"
	pf.MultiYear = false
	return &pf, nil
}

// ParkFactorSeries returns park factors over a range of seasons.
func (r *ParkFactorRepository) ParkFactorSeries(ctx context.Context, parkID core.ParkID, fromSeason, toSeason core.SeasonYear) ([]core.ParkFactor, error) {
	rows, err := r.db.QueryContext(ctx, parkFactorSeriesQuery, string(parkID), int(fromSeason), int(toSeason))
	if err != nil {
		return nil, fmt.Errorf("failed to query park factor series: %w", err)
	}
	defer rows.Close()

	var factors []core.ParkFactor
	for rows.Next() {
		var pf core.ParkFactor
		var homeTeam string

		err = rows.Scan(
			&pf.ParkID,
			&homeTeam,
			&pf.Season,
			&pf.GamesSampled,
			&pf.RunsFactor,
			&pf.HRFactor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park factor: %w", err)
		}

		pf.Provider = "internal"
		pf.MultiYear = false

		factors = append(factors, pf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating park factors: %w", err)
	}

	return factors, nil
}

// SeasonParkFactors returns all park factors for a given season.
func (r *ParkFactorRepository) SeasonParkFactors(ctx context.Context, season core.SeasonYear, factorType *string) ([]core.ParkFactor, error) {
	rows, err := r.db.QueryContext(ctx, seasonParkFactorsQuery, int(season))
	if err != nil {
		return nil, fmt.Errorf("failed to query season park factors: %w", err)
	}
	defer rows.Close()

	var factors []core.ParkFactor
	for rows.Next() {
		var pf core.ParkFactor
		var homeTeam string

		err = rows.Scan(
			&pf.ParkID,
			&homeTeam,
			&pf.Season,
			&pf.GamesSampled,
			&pf.RunsFactor,
			&pf.HRFactor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park factor: %w", err)
		}

		pf.Provider = "internal"
		pf.MultiYear = false

		factors = append(factors, pf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating park factors: %w", err)
	}

	return factors, nil
}

// MultiYearParkFactor returns a park factor averaged over multiple seasons.
func (r *ParkFactorRepository) MultiYearParkFactor(ctx context.Context, parkID core.ParkID, fromSeason, toSeason core.SeasonYear) (*core.ParkFactor, error) {
	series, err := r.ParkFactorSeries(ctx, parkID, fromSeason, toSeason)
	if err != nil {
		return nil, err
	}

	if len(series) == 0 {
		return nil, core.NewNotFoundError("park factors", "")
	}

	var totalGames int
	var totalRunsFactor, totalHRFactor float64

	for _, pf := range series {
		totalGames += pf.GamesSampled
		totalRunsFactor += pf.RunsFactor * float64(pf.GamesSampled)
		totalHRFactor += pf.HRFactor * float64(pf.GamesSampled)
	}

	avgPF := &core.ParkFactor{
		ParkID:       string(parkID),
		Season:       int(toSeason),
		RunsFactor:   totalRunsFactor / float64(totalGames),
		HRFactor:     totalHRFactor / float64(totalGames),
		GamesSampled: totalGames,
		Provider:     "internal",
		MultiYear:    true,
	}

	return avgPF, nil
}

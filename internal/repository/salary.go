package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

// SalaryRepository provides access to salary summary data.
type SalaryRepository struct {
	db *sql.DB
}

// NewSalaryRepository creates a new salary repository.
func NewSalaryRepository(db *sql.DB) *SalaryRepository {
	return &SalaryRepository{db: db}
}

// List retrieves all yearly salary aggregates.
func (r *SalaryRepository) List(ctx context.Context) ([]core.SalarySummary, error) {
	query := `
		SELECT year, total, average, median
		FROM salary_summary
		ORDER BY year DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query salary summary: %w", err)
	}
	defer rows.Close()

	var summaries []core.SalarySummary
	for rows.Next() {
		var s core.SalarySummary
		var total, average, median float64

		if err := rows.Scan(&s.Year, &total, &average, &median); err != nil {
			return nil, fmt.Errorf("failed to scan salary summary: %w", err)
		}

		s.Total = total
		s.Average = average
		s.Median = median

		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating salary summary rows: %w", err)
	}
	return summaries, nil
}

// Get retrieves salary aggregate for a specific year.
func (r *SalaryRepository) Get(ctx context.Context, year core.SeasonYear) (*core.SalarySummary, error) {
	query := `
		SELECT year, total, average, median
		FROM salary_summary
		WHERE year = $1
	`

	var s core.SalarySummary
	var total, average, median float64

	err := r.db.QueryRowContext(ctx, query, int(year)).Scan(&s.Year, &total, &average, &median)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query salary summary: %w", err)
	}

	s.Total = total
	s.Average = average
	s.Median = median
	return &s, nil
}

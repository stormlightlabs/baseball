package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type PostseasonRepository struct {
	db *sql.DB
}

func NewPostseasonRepository(db *sql.DB) *PostseasonRepository {
	return &PostseasonRepository{db: db}
}

// ListSeries retrieves all postseason series for a given year.
func (r *PostseasonRepository) ListSeries(ctx context.Context, year core.SeasonYear) ([]core.PostseasonSeries, error) {
	// TODO: move to embedded query
	query := `
		SELECT
			"yearID",
			"round",
			"teamIDwinner",
			"lgIDwinner",
			"teamIDloser",
			"lgIDloser",
			"wins",
			"losses",
			"ties"
		FROM "SeriesPost"
		WHERE "yearID" = $1
		ORDER BY
			CASE "round"
				WHEN 'WS' THEN 1
				WHEN 'ALCS' THEN 2
				WHEN 'NLCS' THEN 2
				WHEN 'ALDS1' THEN 3
				WHEN 'ALDS2' THEN 3
				WHEN 'NLDS1' THEN 3
				WHEN 'NLDS2' THEN 3
				WHEN 'AEDIV' THEN 3
				WHEN 'NEDIV' THEN 3
				WHEN 'ALEWC' THEN 4
				WHEN 'NLWC' THEN 4
				ELSE 5
			END
	`

	rows, err := r.db.QueryContext(ctx, query, int(year))
	if err != nil {
		return nil, fmt.Errorf("failed to list postseason series: %w", err)
	}
	defer rows.Close()

	var series []core.PostseasonSeries
	for rows.Next() {
		var s core.PostseasonSeries
		var winnerTeam sql.NullString
		var winnerLeague sql.NullString
		var loserTeam sql.NullString
		var loserLeague sql.NullString
		var wins sql.NullInt64
		var losses sql.NullInt64
		var ties sql.NullInt64

		err := rows.Scan(
			&s.Year,
			&s.Round,
			&winnerTeam,
			&winnerLeague,
			&loserTeam,
			&loserLeague,
			&wins,
			&losses,
			&ties,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan postseason series: %w", err)
		}

		if winnerTeam.Valid {
			wt := core.TeamID(winnerTeam.String)
			s.WinnerTeam = &wt
		}
		if winnerLeague.Valid {
			wl := core.LeagueID(winnerLeague.String)
			s.WinnerLeague = &wl
		}
		if loserTeam.Valid {
			lt := core.TeamID(loserTeam.String)
			s.LoserTeam = &lt
		}
		if loserLeague.Valid {
			ll := core.LeagueID(loserLeague.String)
			s.LoserLeague = &ll
		}
		if wins.Valid {
			w := int(wins.Int64)
			s.Wins = &w
		}
		if losses.Valid {
			l := int(losses.Int64)
			s.Losses = &l
		}
		if ties.Valid {
			t := int(ties.Int64)
			s.Ties = &t
		}

		series = append(series, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate postseason series: %w", err)
	}

	return series, nil
}

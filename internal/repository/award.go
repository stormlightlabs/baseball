package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type AwardRepository struct {
	db *sql.DB
}

func NewAwardRepository(db *sql.DB) *AwardRepository {
	return &AwardRepository{db: db}
}

func (r *AwardRepository) ListAwards(ctx context.Context) ([]core.Award, error) {
	return nil, nil
}

func (r *AwardRepository) ListAwardResults(ctx context.Context, filter core.AwardFilter) ([]core.AwardResult, error) {
	query := `
		SELECT
			ap."awardID", ap."playerID", ap."yearID", ap."lgID",
			asp."votesFirst", asp."pointsWon"
		FROM "AwardsPlayers" ap
		LEFT JOIN "AwardsSharePlayers" asp
			ON ap."awardID" = asp."awardID"
			AND ap."playerID" = asp."playerID"
			AND ap."yearID" = asp."yearID"
			AND (ap."lgID" = asp."lgID" OR (ap."lgID" IS NULL AND asp."lgID" IS NULL))
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ap.\"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.AwardID != nil {
		query += fmt.Sprintf(" AND ap.\"awardID\" = $%d", argNum)
		args = append(args, string(*filter.AwardID))
		argNum++
	}

	if filter.Year != nil {
		query += fmt.Sprintf(" AND ap.\"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Year))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND ap.\"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += " ORDER BY ap.\"yearID\" DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list award results: %w", err)
	}
	defer rows.Close()

	var results []core.AwardResult
	for rows.Next() {
		var ar core.AwardResult
		var league sql.NullString
		var votesFirst, points sql.NullFloat64

		err := rows.Scan(
			&ar.AwardID, &ar.PlayerID, &ar.Year, &league,
			&votesFirst, &points,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan award result: %w", err)
		}

		if league.Valid {
			lg := core.LeagueID(league.String)
			ar.League = &lg
		}

		if votesFirst.Valid {
			vf := int(votesFirst.Float64)
			ar.VotesFirst = &vf
		}

		if points.Valid {
			p := int(points.Float64)
			ar.Points = &p
		}

		results = append(results, ar)
	}

	return results, nil
}

func (r *AwardRepository) CountAwardResults(ctx context.Context, filter core.AwardFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "AwardsPlayers" ap WHERE 1=1`
	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ap.\"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.AwardID != nil {
		query += fmt.Sprintf(" AND ap.\"awardID\" = $%d", argNum)
		args = append(args, string(*filter.AwardID))
		argNum++
	}

	if filter.Year != nil {
		query += fmt.Sprintf(" AND ap.\"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Year))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND ap.\"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *AwardRepository) HallOfFameByPlayer(ctx context.Context, id core.PlayerID) ([]core.HallOfFameRecord, error) {
	query := `
		SELECT
			"playerID", "yearid", "votedBy", "ballots", "needed", "votes", "inducted"
		FROM "HallOfFame"
		WHERE "playerID" = $1
		ORDER BY "yearid" DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get hall of fame records: %w", err)
	}
	defer rows.Close()

	var records []core.HallOfFameRecord
	for rows.Next() {
		var hof core.HallOfFameRecord
		var ballots, needed, votes sql.NullInt64
		var inducted sql.NullString

		err := rows.Scan(
			&hof.PlayerID, &hof.Year, &hof.VotedBy,
			&ballots, &needed, &votes, &inducted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hall of fame record: %w", err)
		}

		if ballots.Valid {
			b := int(ballots.Int64)
			hof.Ballots = &b
		}

		if needed.Valid {
			n := int(needed.Int64)
			hof.Needed = &n
		}

		if votes.Valid {
			v := int(votes.Int64)
			hof.Votes = &v
		}

		if inducted.Valid {
			hof.Inducted = inducted.String == "Y"
		}

		records = append(records, hof)
	}

	return records, nil
}

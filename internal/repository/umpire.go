package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/umpire_get_by_id.sql
var umpireGetByIDQuery string

//go:embed queries/umpire_list.sql
var umpireListQuery string

//go:embed queries/umpire_games.sql
var umpireGamesQuery string

type UmpireRepository struct {
	db *sql.DB
}

func NewUmpireRepository(db *sql.DB) *UmpireRepository {
	return &UmpireRepository{db: db}
}

// GetByID retrieves an umpire by their ID.
// Since umpires don't have a dedicated table, we extract from games table.
func (r *UmpireRepository) GetByID(ctx context.Context, id core.UmpireID) (*core.Umpire, error) {
	var umpire core.Umpire
	var fullName sql.NullString

	err := r.db.QueryRowContext(ctx, umpireGetByIDQuery, string(id)).Scan(
		&umpire.ID,
		&fullName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("umpire not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get umpire: %w", err)
	}

	if fullName.Valid {
		name := fullName.String
		umpire.FirstName = ""
		umpire.LastName = name
	}

	return &umpire, nil
}

// List retrieves all umpires with pagination.
func (r *UmpireRepository) List(ctx context.Context, p core.Pagination) ([]core.Umpire, error) {
	query := umpireListQuery
	args := []any{}
	if p.PerPage > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list umpires: %w", err)
	}
	defer rows.Close()

	var umpires []core.Umpire
	for rows.Next() {
		var umpire core.Umpire
		var fullName sql.NullString

		err := rows.Scan(
			&umpire.ID,
			&fullName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan umpire: %w", err)
		}

		if fullName.Valid {
			umpire.FirstName = ""
			umpire.LastName = fullName.String
		}

		umpires = append(umpires, umpire)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate umpires: %w", err)
	}

	return umpires, nil
}

// GamesForUmpire retrieves games where the umpire officiated in any position.
func (r *UmpireRepository) GamesForUmpire(ctx context.Context, id core.UmpireID, filter core.GameFilter) ([]core.Game, error) {
	query := umpireGamesQuery

	args := []any{string(id)}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND SUBSTRING(date, 1, 4) = $%d", argNum)
		args = append(args, fmt.Sprintf("%04d", int(*filter.Season)))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, filter.DateFrom.Format("20060102"))
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, filter.DateTo.Format("20060102"))
		argNum++
	}

	if filter.HomeTeam != nil {
		query += fmt.Sprintf(" AND home_team = $%d", argNum)
		args = append(args, string(*filter.HomeTeam))
		argNum++
	}

	if filter.AwayTeam != nil {
		query += fmt.Sprintf(" AND visiting_team = $%d", argNum)
		args = append(args, string(*filter.AwayTeam))
		argNum++
	}

	if filter.ParkID != nil {
		query += fmt.Sprintf(" AND park_id = $%d", argNum)
		args = append(args, string(*filter.ParkID))
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY date DESC, game_number LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list games for umpire: %w", err)
	}
	defer rows.Close()

	var games []core.Game
	for rows.Next() {
		var g core.Game
		var date string
		var gameNumber int
		var attendance, durationMin sql.NullInt64
		var umpHome, umpFirst, umpSecond, umpThird sql.NullString
		var homeTeam, awayTeam, homeLeague, awayLeague, parkID string

		err := rows.Scan(
			&date,
			&gameNumber,
			&awayTeam,
			&homeTeam,
			&awayLeague,
			&homeLeague,
			&g.AwayScore,
			&g.HomeScore,
			&g.Innings,
			&g.DayOfWeek,
			&attendance,
			&durationMin,
			&parkID,
			&umpHome,
			&umpFirst,
			&umpSecond,
			&umpThird,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		g.ID = core.GameID(fmt.Sprintf("%s%d%s", date, gameNumber, homeTeam))
		g.HomeTeam = core.TeamID(homeTeam)
		g.AwayTeam = core.TeamID(awayTeam)
		g.HomeLeague = core.LeagueID(homeLeague)
		g.AwayLeague = core.LeagueID(awayLeague)
		g.ParkID = core.ParkID(parkID)

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		g.Date = parsedDate
		g.Season = core.SeasonYear(parsedDate.Year())

		g.Innings = g.Innings / 3

		if attendance.Valid {
			a := int(attendance.Int64)
			g.Attendance = &a
		}
		if durationMin.Valid {
			d := int(durationMin.Int64)
			g.DurationMin = &d
		}
		if umpHome.Valid {
			u := core.UmpireID(umpHome.String)
			g.UmpHome = &u
		}
		if umpFirst.Valid {
			u := core.UmpireID(umpFirst.String)
			g.UmpFirst = &u
		}
		if umpSecond.Valid {
			u := core.UmpireID(umpSecond.String)
			g.UmpSecond = &u
		}
		if umpThird.Valid {
			u := core.UmpireID(umpThird.String)
			g.UmpThird = &u
		}

		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate games: %w", err)
	}

	return games, nil
}

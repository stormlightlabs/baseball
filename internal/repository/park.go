package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/park_games.sql
var parkGamesQuery string

type ParkRepository struct {
	db *sql.DB
}

func NewParkRepository(db *sql.DB) *ParkRepository {
	return &ParkRepository{db: db}
}

// GetByID retrieves a park by its park key/ID.
func (r *ParkRepository) GetByID(ctx context.Context, id core.ParkID) (*core.Park, error) {
	query := `
		SELECT
			"parkkey",
			"parkname",
			"city",
			"state",
			"country"
		FROM "Parks"
		WHERE "parkkey" = $1
	`

	var park core.Park
	var state sql.NullString
	var country sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&park.ID,
		&park.Name,
		&park.City,
		&state,
		&country,
	)

	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("park", string(id))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get park: %w", err)
	}

	if state.Valid {
		park.State = state.String
	}
	if country.Valid {
		park.Country = country.String
	}

	return &park, nil
}

// List retrieves parks with optional search and pagination.
func (r *ParkRepository) List(ctx context.Context, filter core.ParkFilter) ([]core.Park, error) {
	query := `
		SELECT DISTINCT
			"parkkey",
			"parkname",
			"city",
			"state",
			"country"
		FROM "Parks"
		WHERE "parkkey" IS NOT NULL
		  AND "parkname" IS NOT NULL
	`

	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(` AND (
			"parkname" ILIKE $%d OR
			"city" ILIKE $%d OR
			"state" ILIKE $%d OR
			"parkkey" ILIKE $%d
		)`, argNum, argNum, argNum, argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	query += " ORDER BY \"parkname\""

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list parks: %w", err)
	}
	defer rows.Close()

	var parks []core.Park
	for rows.Next() {
		var park core.Park
		var state sql.NullString
		var country sql.NullString

		err := rows.Scan(
			&park.ID,
			&park.Name,
			&park.City,
			&state,
			&country,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park: %w", err)
		}

		if state.Valid {
			park.State = state.String
		}
		if country.Valid {
			park.Country = country.String
		}

		parks = append(parks, park)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate parks: %w", err)
	}

	return parks, nil
}

// Count returns the total number of parks matching the filter criteria.
func (r *ParkRepository) Count(ctx context.Context, filter core.ParkFilter) (int, error) {
	query := `
		SELECT COUNT(DISTINCT "parkkey")
		FROM "Parks"
		WHERE "parkkey" IS NOT NULL
		  AND "parkname" IS NOT NULL
	`

	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(` AND (
			"parkname" ILIKE $%d OR
			"city" ILIKE $%d OR
			"state" ILIKE $%d OR
			"parkkey" ILIKE $%d
		)`, argNum, argNum, argNum, argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count parks: %w", err)
	}

	return count, nil
}

// GamesAtPark retrieves games played at a specific park.
func (r *ParkRepository) GamesAtPark(ctx context.Context, id core.ParkID, filter core.GameFilter) ([]core.Game, error) {
	query := parkGamesQuery

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

	query += fmt.Sprintf(" ORDER BY date DESC, game_number LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list games at park: %w", err)
	}
	defer rows.Close()

	var games []core.Game
	for rows.Next() {
		var g core.Game
		var date string
		var gameNumber int
		var innings, attendance, durationMin sql.NullInt64
		var dayOfWeek, homeLeague, awayLeague, umpHome, umpFirst, umpSecond, umpThird sql.NullString
		var homeTeam, awayTeam, parkID string

		err := rows.Scan(
			&date,
			&gameNumber,
			&awayTeam,
			&homeTeam,
			&awayLeague,
			&homeLeague,
			&g.AwayScore,
			&g.HomeScore,
			&innings,
			&dayOfWeek,
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
		g.ParkID = core.ParkID(parkID)

		if homeLeague.Valid {
			g.HomeLeague = core.LeagueID(homeLeague.String)
		}
		if awayLeague.Valid {
			g.AwayLeague = core.LeagueID(awayLeague.String)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		g.Date = parsedDate
		g.Season = core.SeasonYear(parsedDate.Year())

		if innings.Valid {
			g.Innings = int(innings.Int64) / 3
		}

		if dayOfWeek.Valid {
			g.DayOfWeek = dayOfWeek.String
		}

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

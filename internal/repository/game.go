package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

// GetByID retrieves a single game by its Retrosheet game ID.
// The game ID format is constructed from date + game number + home team.
func (r *GameRepository) GetByID(ctx context.Context, id core.GameID) (*core.Game, error) {
	query := `
		SELECT
			date,
			visiting_team,
			home_team,
			visiting_team_league,
			home_team_league,
			visiting_score,
			home_score,
			game_length_outs,
			day_of_week,
			attendance,
			game_time_minutes,
			park_id,
			hp_ump_id,
			b1_ump_id,
			b2_ump_id,
			b3_ump_id
		FROM games
		WHERE date || game_number || home_team = $1
	`

	var g core.Game
	var date string
	var attendance, durationMin sql.NullInt64
	var umpHome, umpFirst, umpSecond, umpThird sql.NullString
	var homeTeam, awayTeam, homeLeague, awayLeague, parkID string

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&date,
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

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("game not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	g.ID = id
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

	return &g, nil
}

// List retrieves games based on filter criteria with pagination.
func (r *GameRepository) List(ctx context.Context, filter core.GameFilter) ([]core.Game, error) {
	query := `
		SELECT
			date,
			game_number,
			visiting_team,
			home_team,
			visiting_team_league,
			home_team_league,
			visiting_score,
			home_score,
			game_length_outs,
			day_of_week,
			attendance,
			game_time_minutes,
			park_id,
			hp_ump_id,
			b1_ump_id,
			b2_ump_id,
			b3_ump_id
		FROM games
		WHERE 1=1
	`

	args := []interface{}{}
	argNum := 1

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

	if filter.ParkID != nil {
		query += fmt.Sprintf(" AND park_id = $%d", argNum)
		args = append(args, string(*filter.ParkID))
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY date DESC, game_number LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
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
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	return games, nil
}

// Count returns the total number of games matching the filter.
func (r *GameRepository) Count(ctx context.Context, filter core.GameFilter) (int, error) {
	query := `SELECT COUNT(*) FROM games WHERE 1=1`

	args := []interface{}{}
	argNum := 1

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

	if filter.ParkID != nil {
		query += fmt.Sprintf(" AND park_id = $%d", argNum)
		args = append(args, string(*filter.ParkID))
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count games: %w", err)
	}

	return count, nil
}

// ListByDate retrieves all games played on a specific date.
func (r *GameRepository) ListByDate(ctx context.Context, date time.Time) ([]core.Game, error) {
	filter := core.GameFilter{
		DateFrom: &date,
		DateTo:   &date,
		Pagination: core.Pagination{
			Page:    1,
			PerPage: 100,
		},
	}
	return r.List(ctx, filter)
}

// ListByTeamSeason retrieves all games for a specific team in a season.
func (r *GameRepository) ListByTeamSeason(ctx context.Context, teamID core.TeamID, year core.SeasonYear, p core.Pagination) ([]core.Game, error) {
	filter := core.GameFilter{
		HomeTeam:   &teamID,
		Season:     &year,
		Pagination: p,
	}
	return r.List(ctx, filter)
}

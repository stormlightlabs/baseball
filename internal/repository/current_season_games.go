package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

func (r *GameRepository) shouldUseCurrentSeasonGames(ctx context.Context, filter core.GameFilter) (bool, error) {
	if filter.Season == nil {
		if filter.ID == nil {
			return false, nil
		}
		_, err := strconv.Atoi(string(*filter.ID))
		if err != nil {
			return false, nil
		}
		return true, nil
	}

	hasCurrent, err := r.currentSeasonGameRowsForSeason(ctx, *filter.Season)
	if err != nil {
		return false, err
	}
	if !hasCurrent {
		return false, nil
	}
	if int(*filter.Season) >= currentYear() {
		return true, nil
	}

	var hasHistorical bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM games WHERE SUBSTRING(date, 1, 4) = $1 LIMIT 1)`, fmt.Sprintf("%04d", int(*filter.Season))).Scan(&hasHistorical); err != nil {
		return false, fmt.Errorf("failed checking historical games for season %d: %w", *filter.Season, err)
	}
	return !hasHistorical, nil
}

func (r *GameRepository) currentSeasonGameRowsForSeason(ctx context.Context, season core.SeasonYear) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM current_season.games WHERE season = $1 LIMIT 1)`, int(season)).Scan(&exists); err != nil {
		return false, nil
	}
	return exists, nil
}

func (r *GameRepository) getCurrentSeasonGameByPK(ctx context.Context, gamePK int) (*core.Game, error) {
	query := `
		SELECT
			g.game_pk,
			g.season,
			g.game_date,
			g.status,
			COALESCE(g.away_team_id, atm.local_team_id, ''),
			COALESCE(g.home_team_id, htm.local_team_id, ''),
			COALESCE(atm.local_league, ''),
			COALESCE(htm.local_league, ''),
			g.away_score,
			g.home_score,
			g.innings,
			g.day_night,
			g.venue,
			g.fetched_at
		FROM current_season.games g
		LEFT JOIN team_mlbam_map htm
			ON htm.season = g.season
			AND htm.mlbam_team_id = g.home_mlb_id
		LEFT JOIN team_mlbam_map atm
			ON atm.season = g.season
			AND atm.mlbam_team_id = g.away_mlb_id
		WHERE g.game_pk = $1
	`

	row := r.db.QueryRowContext(ctx, query, gamePK)
	game, err := scanCurrentSeasonGameRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.NewNotFoundError("game", strconv.Itoa(gamePK))
		}
		return nil, err
	}
	return game, nil
}

func (r *GameRepository) listCurrentSeasonGames(ctx context.Context, filter core.GameFilter) ([]core.Game, error) {
	query := `
		SELECT
			g.game_pk,
			g.season,
			g.game_date,
			g.status,
			COALESCE(g.away_team_id, atm.local_team_id, ''),
			COALESCE(g.home_team_id, htm.local_team_id, ''),
			COALESCE(atm.local_league, ''),
			COALESCE(htm.local_league, ''),
			g.away_score,
			g.home_score,
			g.innings,
			g.day_night,
			g.venue,
			g.fetched_at
		FROM current_season.games g
		LEFT JOIN team_mlbam_map htm
			ON htm.season = g.season
			AND htm.mlbam_team_id = g.home_mlb_id
		LEFT JOIN team_mlbam_map atm
			ON atm.season = g.season
			AND atm.mlbam_team_id = g.away_mlb_id
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.ID != nil {
		gamePK, err := strconv.Atoi(string(*filter.ID))
		if err != nil {
			return []core.Game{}, nil
		}
		query += fmt.Sprintf(" AND g.game_pk = $%d", argNum)
		args = append(args, gamePK)
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND g.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.HomeTeam != nil {
		query += fmt.Sprintf(" AND COALESCE(g.home_team_id, htm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.HomeTeam))
		argNum++
	}
	if filter.AwayTeam != nil {
		query += fmt.Sprintf(" AND COALESCE(g.away_team_id, atm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.AwayTeam))
		argNum++
	}
	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND g.game_date >= $%d", argNum)
		args = append(args, filter.DateFrom.Format("2006-01-02"))
		argNum++
	}
	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND g.game_date <= $%d", argNum)
		args = append(args, filter.DateTo.Format("2006-01-02"))
		argNum++
	}
	if filter.ParkID != nil {
		query += " AND 1=0"
	}
	if len(filter.Leagues) > 0 {
		query += fmt.Sprintf(" AND (COALESCE(htm.local_league, '') = ANY($%d) OR COALESCE(atm.local_league, '') = ANY($%d))", argNum, argNum+1)
		leagues := make([]string, len(filter.Leagues))
		for i, league := range filter.Leagues {
			leagues[i] = string(league)
		}
		args = append(args, leagues, leagues)
		argNum += 2
	} else if filter.League != nil {
		query += fmt.Sprintf(" AND (COALESCE(htm.local_league, '') = $%d OR COALESCE(atm.local_league, '') = $%d)", argNum, argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY g.game_date DESC, g.game_pk DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list current-season games: %w", err)
	}
	defer rows.Close()

	games := []core.Game{}
	for rows.Next() {
		game, err := scanCurrentSeasonGameRow(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, *game)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating current-season games: %w", err)
	}
	return games, nil
}

func (r *GameRepository) countCurrentSeasonGames(ctx context.Context, filter core.GameFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM current_season.games g
		LEFT JOIN team_mlbam_map htm
			ON htm.season = g.season
			AND htm.mlbam_team_id = g.home_mlb_id
		LEFT JOIN team_mlbam_map atm
			ON atm.season = g.season
			AND atm.mlbam_team_id = g.away_mlb_id
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.ID != nil {
		gamePK, err := strconv.Atoi(string(*filter.ID))
		if err != nil {
			return 0, nil
		}
		query += fmt.Sprintf(" AND g.game_pk = $%d", argNum)
		args = append(args, gamePK)
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND g.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.HomeTeam != nil {
		query += fmt.Sprintf(" AND COALESCE(g.home_team_id, htm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.HomeTeam))
		argNum++
	}
	if filter.AwayTeam != nil {
		query += fmt.Sprintf(" AND COALESCE(g.away_team_id, atm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.AwayTeam))
		argNum++
	}
	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND g.game_date >= $%d", argNum)
		args = append(args, filter.DateFrom.Format("2006-01-02"))
		argNum++
	}
	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND g.game_date <= $%d", argNum)
		args = append(args, filter.DateTo.Format("2006-01-02"))
		argNum++
	}
	if filter.ParkID != nil {
		query += " AND 1=0"
	}
	if len(filter.Leagues) > 0 {
		query += fmt.Sprintf(" AND (COALESCE(htm.local_league, '') = ANY($%d) OR COALESCE(atm.local_league, '') = ANY($%d))", argNum, argNum+1)
		leagues := make([]string, len(filter.Leagues))
		for i, league := range filter.Leagues {
			leagues[i] = string(league)
		}
		args = append(args, leagues, leagues)
		argNum += 2
	} else if filter.League != nil {
		query += fmt.Sprintf(" AND (COALESCE(htm.local_league, '') = $%d OR COALESCE(atm.local_league, '') = $%d)", argNum, argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count current-season games: %w", err)
	}
	return count, nil
}

func scanCurrentSeasonGameRow(scanner interface {
	Scan(dest ...any) error
}) (*core.Game, error) {
	var gamePK int
	var season int
	var gameDate time.Time
	var status string
	var awayTeam string
	var homeTeam string
	var awayLeague string
	var homeLeague string
	var awayScore sql.NullInt64
	var homeScore sql.NullInt64
	var innings sql.NullInt64
	var dayNight sql.NullString
	var venue sql.NullString
	var fetchedAt time.Time

	if err := scanner.Scan(
		&gamePK,
		&season,
		&gameDate,
		&status,
		&awayTeam,
		&homeTeam,
		&awayLeague,
		&homeLeague,
		&awayScore,
		&homeScore,
		&innings,
		&dayNight,
		&venue,
		&fetchedAt,
	); err != nil {
		return nil, err
	}

	game := &core.Game{
		ID:           core.GameID(strconv.Itoa(gamePK)),
		Season:       core.SeasonYear(season),
		Date:         gameDate,
		DayOfWeek:    gameDate.Weekday().String(),
		Status:       &status,
		Source:       "current_season",
		FetchedAt:    &fetchedAt,
		HomeTeam:     core.TeamID(homeTeam),
		AwayTeam:     core.TeamID(awayTeam),
		HomeLeague:   core.LeagueID(homeLeague),
		AwayLeague:   core.LeagueID(awayLeague),
		IsPostseason: gameDate.Month() >= 10 && gameDate.Month() <= 11,
	}
	if awayScore.Valid {
		game.AwayScore = int(awayScore.Int64)
	}
	if homeScore.Valid {
		game.HomeScore = int(homeScore.Int64)
	}
	if innings.Valid {
		game.Innings = int(innings.Int64)
	}
	if dayNight.Valid {
		v := dayNight.String
		game.DayNight = &v
	}
	if venue.Valid {
		v := venue.String
		game.ParkName = &v
		game.ParkID = core.ParkID(v)
	}
	return game, nil
}

func errorsIsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var nf *core.NotFoundError
	return errors.As(err, &nf)
}

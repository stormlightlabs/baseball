package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/game_boxscore.sql
var gameBoxscoreQuery string

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

	gameNumberFromID := 0
	if len(string(id)) > 8 {
		gameNumberFromID = int(string(id)[8] - '0')
	}
	g.IsPostseason = parsedDate.Month() >= 10 && parsedDate.Month() <= 11 && gameNumberFromID == 0

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

	args := []any{}
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

	if filter.IsPostseason != nil {
		if *filter.IsPostseason {
			query += " AND (SUBSTRING(date, 5, 2) IN ('10', '11') AND game_number = 0)"
		} else {
			query += " AND NOT (SUBSTRING(date, 5, 2) IN ('10', '11') AND game_number = 0)"
		}
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

		g.IsPostseason = parsedDate.Month() >= 10 && parsedDate.Month() <= 11 && gameNumber == 0

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

	args := []any{}
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

	if filter.IsPostseason != nil {
		if *filter.IsPostseason {
			query += " AND (SUBSTRING(date, 5, 2) IN ('10', '11') AND game_number = 0)"
		} else {
			query += " AND NOT (SUBSTRING(date, 5, 2) IN ('10', '11') AND game_number = 0)"
		}
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

// GetBoxscore retrieves detailed boxscore statistics for a game.
func (r *GameRepository) GetBoxscore(ctx context.Context, id core.GameID) (*core.Boxscore, error) {
	query := gameBoxscoreQuery

	var date string
	var visitingTeam, homeTeam string
	var box core.Boxscore
	var vStats, hStats core.TeamGameStats

	var vP1ID, vP1Name sql.NullString
	var vP1Pos sql.NullInt64
	var vP2ID, vP2Name sql.NullString
	var vP2Pos sql.NullInt64
	var vP3ID, vP3Name sql.NullString
	var vP3Pos sql.NullInt64
	var vP4ID, vP4Name sql.NullString
	var vP4Pos sql.NullInt64
	var vP5ID, vP5Name sql.NullString
	var vP5Pos sql.NullInt64
	var vP6ID, vP6Name sql.NullString
	var vP6Pos sql.NullInt64
	var vP7ID, vP7Name sql.NullString
	var vP7Pos sql.NullInt64
	var vP8ID, vP8Name sql.NullString
	var vP8Pos sql.NullInt64
	var vP9ID, vP9Name sql.NullString
	var vP9Pos sql.NullInt64

	var hP1ID, hP1Name sql.NullString
	var hP1Pos sql.NullInt64
	var hP2ID, hP2Name sql.NullString
	var hP2Pos sql.NullInt64
	var hP3ID, hP3Name sql.NullString
	var hP3Pos sql.NullInt64
	var hP4ID, hP4Name sql.NullString
	var hP4Pos sql.NullInt64
	var hP5ID, hP5Name sql.NullString
	var hP5Pos sql.NullInt64
	var hP6ID, hP6Name sql.NullString
	var hP6Pos sql.NullInt64
	var hP7ID, hP7Name sql.NullString
	var hP7Pos sql.NullInt64
	var hP8ID, hP8Name sql.NullString
	var hP8Pos sql.NullInt64
	var hP9ID, hP9Name sql.NullString
	var hP9Pos sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&date,
		&visitingTeam,
		&homeTeam,
		&box.AwayScore,
		&box.HomeScore,
		&vStats.AB, &vStats.H, &vStats.Doubles, &vStats.Triples, &vStats.HR,
		&vStats.RBI, &vStats.SH, &vStats.SF, &vStats.HBP,
		&vStats.BB, &vStats.IBB, &vStats.SO, &vStats.SB,
		&vStats.CS, &vStats.GDP, &vStats.LOB, &vStats.PitchersUsed,
		&vStats.ER, &vStats.WP, &vStats.Balks, &vStats.PO,
		&vStats.A, &vStats.E, &vStats.PB, &vStats.DP,
		&vStats.TP,
		&hStats.AB, &hStats.H, &hStats.Doubles, &hStats.Triples, &hStats.HR,
		&hStats.RBI, &hStats.SH, &hStats.SF, &hStats.HBP,
		&hStats.BB, &hStats.IBB, &hStats.SO, &hStats.SB,
		&hStats.CS, &hStats.GDP, &hStats.LOB, &hStats.PitchersUsed,
		&hStats.ER, &hStats.WP, &hStats.Balks, &hStats.PO,
		&hStats.A, &hStats.E, &hStats.PB, &hStats.DP,
		&hStats.TP,
		&vP1ID, &vP1Name, &vP1Pos,
		&vP2ID, &vP2Name, &vP2Pos,
		&vP3ID, &vP3Name, &vP3Pos,
		&vP4ID, &vP4Name, &vP4Pos,
		&vP5ID, &vP5Name, &vP5Pos,
		&vP6ID, &vP6Name, &vP6Pos,
		&vP7ID, &vP7Name, &vP7Pos,
		&vP8ID, &vP8Name, &vP8Pos,
		&vP9ID, &vP9Name, &vP9Pos,
		&hP1ID, &hP1Name, &hP1Pos,
		&hP2ID, &hP2Name, &hP2Pos,
		&hP3ID, &hP3Name, &hP3Pos,
		&hP4ID, &hP4Name, &hP4Pos,
		&hP5ID, &hP5Name, &hP5Pos,
		&hP6ID, &hP6Name, &hP6Pos,
		&hP7ID, &hP7Name, &hP7Pos,
		&hP8ID, &hP8Name, &hP8Pos,
		&hP9ID, &hP9Name, &hP9Pos,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("game not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get boxscore: %w", err)
	}

	box.GameID = id
	box.HomeTeam = core.TeamID(homeTeam)
	box.AwayTeam = core.TeamID(visitingTeam)

	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}
	box.Date = parsedDate

	vStats.R = box.AwayScore
	hStats.R = box.HomeScore

	box.HomeStats = hStats
	box.AwayStats = vStats

	addPlayerToLineup := func(playerID, playerName sql.NullString, pos sql.NullInt64) *core.LineupPlayer {
		if playerID.Valid && pos.Valid {
			p := &core.LineupPlayer{
				PlayerID: core.PlayerID(playerID.String),
				Position: int(pos.Int64),
			}
			if playerName.Valid {
				p.Name = playerName.String
			}
			return p
		}
		return nil
	}

	awayLineup := []core.LineupPlayer{}
	if p := addPlayerToLineup(vP1ID, vP1Name, vP1Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP2ID, vP2Name, vP2Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP3ID, vP3Name, vP3Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP4ID, vP4Name, vP4Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP5ID, vP5Name, vP5Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP6ID, vP6Name, vP6Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP7ID, vP7Name, vP7Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP8ID, vP8Name, vP8Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}
	if p := addPlayerToLineup(vP9ID, vP9Name, vP9Pos); p != nil {
		awayLineup = append(awayLineup, *p)
	}

	homeLineup := []core.LineupPlayer{}
	if p := addPlayerToLineup(hP1ID, hP1Name, hP1Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP2ID, hP2Name, hP2Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP3ID, hP3Name, hP3Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP4ID, hP4Name, hP4Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP5ID, hP5Name, hP5Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP6ID, hP6Name, hP6Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP7ID, hP7Name, hP7Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP8ID, hP8Name, hP8Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}
	if p := addPlayerToLineup(hP9ID, hP9Name, hP9Pos); p != nil {
		homeLineup = append(homeLineup, *p)
	}

	box.AwayLineup = awayLineup
	box.HomeLineup = homeLineup
	return &box, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) GetByID(ctx context.Context, id core.PlayerID) (*core.Player, error) {
	query := `
		SELECT
			"playerID", "nameFirst", "nameLast", "nameGiven",
			"birthYear", "birthMonth", "birthDay", "birthCity", "birthState", "birthCountry",
			"deathYear", "deathMonth", "deathDay", "deathCity", "deathState", "deathCountry",
			"bats", "throws", "weight", "height",
			"debut", "finalGame", "retroID", "bbrefID"
		FROM "People"
		WHERE "playerID" = $1
	`

	var p core.Player
	var retroID, bbrefID sql.NullString
	var debut, finalGame sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.GivenName,
		&p.BirthYear, &p.BirthMonth, &p.BirthDay, &p.BirthCity, &p.BirthState, &p.BirthCountry,
		&p.DeathYear, &p.DeathMonth, &p.DeathDay, &p.DeathCity, &p.DeathState, &p.DeathCountry,
		&p.Bats, &p.Throws, &p.WeightLbs, &p.HeightInches,
		&debut, &finalGame, &retroID, &bbrefID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if retroID.Valid {
		rid := core.RetroPlayerID(retroID.String)
		p.RetroID = &rid
	}

	return &p, nil
}

func (r *PlayerRepository) List(ctx context.Context, filter core.PlayerFilter) ([]core.Player, error) {
	query := `
		SELECT
			"playerID", "nameFirst", "nameLast", "nameGiven",
			"birthYear", "birthMonth", "birthDay", "birthCity", "birthState", "birthCountry",
			"deathYear", "deathMonth", "deathDay", "deathCity", "deathState", "deathCountry",
			"bats", "throws", "weight", "height",
			"debut", "finalGame", "retroID", "bbrefID"
		FROM "People"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(" AND (LOWER(\"nameFirst\" || ' ' || \"nameLast\") LIKE LOWER($%d))", argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	if filter.DebutYear != nil {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM \"debut\"::date) = $%d", argNum)
		args = append(args, int(*filter.DebutYear))
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY \"nameLast\", \"nameFirst\" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	defer rows.Close()

	var players []core.Player
	for rows.Next() {
		var p core.Player
		var retroID, bbrefID sql.NullString
		var debut, finalGame sql.NullString

		err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.GivenName,
			&p.BirthYear, &p.BirthMonth, &p.BirthDay, &p.BirthCity, &p.BirthState, &p.BirthCountry,
			&p.DeathYear, &p.DeathMonth, &p.DeathDay, &p.DeathCity, &p.DeathState, &p.DeathCountry,
			&p.Bats, &p.Throws, &p.WeightLbs, &p.HeightInches,
			&debut, &finalGame, &retroID, &bbrefID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}

		if retroID.Valid {
			rid := core.RetroPlayerID(retroID.String)
			p.RetroID = &rid
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *PlayerRepository) Count(ctx context.Context, filter core.PlayerFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "People" WHERE 1=1`
	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(" AND (LOWER(\"nameFirst\" || ' ' || \"nameLast\") LIKE LOWER($%d))", argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	if filter.DebutYear != nil {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM \"debut\"::date) = $%d", argNum)
		args = append(args, int(*filter.DebutYear))
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *PlayerRepository) BattingSeasons(ctx context.Context, id core.PlayerID) ([]core.PlayerBattingSeason, error) {
	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"G", "AB", "R", "H", "2B", "3B", "HR", "RBI", "SB", "CS", "BB", "SO", "HBP", "SF"
		FROM "Batting"
		WHERE "playerID" = $1
		ORDER BY "yearID", "stint"
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get batting seasons: %w", err)
	}
	defer rows.Close()

	var seasons []core.PlayerBattingSeason
	for rows.Next() {
		var s core.PlayerBattingSeason
		var doubles, triples, cs, hbp, sf sql.NullInt64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.G, &s.AB, &s.R, &s.H, &doubles, &triples, &s.HR, &s.RBI, &s.SB, &cs, &s.BB, &s.SO, &hbp, &sf,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batting season: %w", err)
		}

		if doubles.Valid {
			s.Doubles = int(doubles.Int64)
		}
		if triples.Valid {
			s.Triples = int(triples.Int64)
		}
		if cs.Valid {
			s.CS = int(cs.Int64)
		}
		if hbp.Valid {
			s.HBP = int(hbp.Int64)
		}
		if sf.Valid {
			s.SF = int(sf.Int64)
		}

		s.PA = s.AB + s.BB + s.HBP + s.SF

		if s.AB > 0 {
			s.AVG = float64(s.H) / float64(s.AB)
			singles := s.H - s.Doubles - s.Triples - s.HR
			totalBases := singles + (s.Doubles * 2) + (s.Triples * 3) + (s.HR * 4)
			s.SLG = float64(totalBases) / float64(s.AB)
		}

		if s.PA > 0 {
			s.OBP = float64(s.H+s.BB+s.HBP) / float64(s.PA)
		}

		s.OPS = s.OBP + s.SLG

		seasons = append(seasons, s)
	}

	return seasons, nil
}

func (r *PlayerRepository) PitchingSeasons(ctx context.Context, id core.PlayerID) ([]core.PlayerPitchingSeason, error) {
	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"W", "L", "G", "GS", "CG", "SHO", "SV", "IPouts", "H", "ER", "HR", "BB", "SO", "HBP", "BK", "WP", "ERA"
		FROM "Pitching"
		WHERE "playerID" = $1
		ORDER BY "yearID", "stint"
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get pitching seasons: %w", err)
	}
	defer rows.Close()

	var seasons []core.PlayerPitchingSeason
	for rows.Next() {
		var s core.PlayerPitchingSeason
		var era sql.NullFloat64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.W, &s.L, &s.G, &s.GS, &s.CG, &s.SHO, &s.SV, &s.IPOuts, &s.H, &s.ER, &s.HR, &s.BB, &s.SO, &s.HBP, &s.BK, &s.WP, &era,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pitching season: %w", err)
		}

		if era.Valid {
			s.ERA = era.Float64
		}

		ip := float64(s.IPOuts) / 3.0
		if ip > 0 {
			s.WHIP = float64(s.H+s.BB) / ip
			s.KPer9 = (float64(s.SO) / ip) * 9.0
			s.BBPer9 = (float64(s.BB) / ip) * 9.0
			s.HRPer9 = (float64(s.HR) / ip) * 9.0
		}

		seasons = append(seasons, s)
	}

	return seasons, nil
}

func (r *PlayerRepository) FieldingSeasons(ctx context.Context, id core.PlayerID) ([]core.PlayerFieldingSeason, error) {
	return nil, nil
}

// GameLogs retrieves games where the player appeared in the starting lineup.
// It uses the Retrosheet player ID to query the games table
func (r *PlayerRepository) GameLogs(ctx context.Context, id core.PlayerID, filter core.GameFilter) ([]core.Game, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player retroID: %w", err)
	}

	// If the player doesn't have a Retrosheet ID, they won't have game logs
	if !retroID.Valid || retroID.String == "" {
		return []core.Game{}, nil
	}

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
		WHERE (
			v_player_1_id = $1 OR v_player_2_id = $1 OR v_player_3_id = $1 OR
			v_player_4_id = $1 OR v_player_5_id = $1 OR v_player_6_id = $1 OR
			v_player_7_id = $1 OR v_player_8_id = $1 OR v_player_9_id = $1 OR
			h_player_1_id = $1 OR h_player_2_id = $1 OR h_player_3_id = $1 OR
			h_player_4_id = $1 OR h_player_5_id = $1 OR h_player_6_id = $1 OR
			h_player_7_id = $1 OR h_player_8_id = $1 OR h_player_9_id = $1
		)
	`

	args := []any{retroID.String}
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

	query += fmt.Sprintf(" ORDER BY date DESC, game_number LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list player game logs: %w", err)
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
			return nil, fmt.Errorf("failed to scan game log: %w", err)
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
		return nil, fmt.Errorf("error iterating game logs: %w", err)
	}

	return games, nil
}

// Appearances retrieves appearance records by position for a player across all seasons.
func (r *PlayerRepository) Appearances(ctx context.Context, id core.PlayerID) ([]core.PlayerAppearance, error) {
	query := `
		SELECT
			"yearID", "teamID", "lgID", "G_all", "GS", "G_batting", "G_defense",
			"G_p", "G_c", "G_1b", "G_2b", "G_3b", "G_ss", "G_lf", "G_cf", "G_rf", "G_of", "G_dh", "G_ph", "G_pr"
		FROM "Appearances"
		WHERE "playerID" = $1
		ORDER BY "yearID" DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to query appearances: %w", err)
	}
	defer rows.Close()

	var appearances []core.PlayerAppearance
	for rows.Next() {
		var a core.PlayerAppearance
		var league sql.NullString

		err := rows.Scan(
			&a.Year, &a.TeamID, &league, &a.GamesAll, &a.GamesStarted, &a.GBatting, &a.GDefense,
			&a.GP, &a.GC, &a.G1B, &a.G2B, &a.G3B, &a.GSS, &a.GLF, &a.GCF, &a.GRF, &a.GOF, &a.GDH, &a.GPH, &a.GPR,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan appearance: %w", err)
		}

		a.PlayerID = id
		if league.Valid {
			lg := core.LeagueID(league.String)
			a.League = &lg
		}

		appearances = append(appearances, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating appearances: %w", err)
	}

	return appearances, nil
}

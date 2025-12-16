package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/player_game_logs.sql
var playerGameLogsQuery string

//go:embed queries/player_appearances.sql
var playerAppearancesQuery string

type PlayerRepository struct {
	db    *sql.DB
	cache *cache.CachedRepository
}

func NewPlayerRepository(db *sql.DB, cacheClient *cache.Client) *PlayerRepository {
	return &PlayerRepository{
		db:    db,
		cache: cache.NewCachedRepository(cacheClient, "player"),
	}
}

func (r *PlayerRepository) GetByID(ctx context.Context, id core.PlayerID) (*core.Player, error) {
	var player core.Player
	if r.cache.Entity.Get(ctx, string(id), &player) {
		return &player, nil
	}

	query := `
		SELECT
			p."playerID", p."nameFirst", p."nameLast", p."nameGiven",
			p."birthYear", p."birthMonth", p."birthDay", p."birthCity", p."birthState", p."birthCountry",
			p."deathYear", p."deathMonth", p."deathDay", p."deathCity", p."deathState", p."deathCountry",
			p."bats", p."throws", p."weight", p."height",
			p."debut", p."finalGame", p."retroID", p."bbrefID"
		FROM "People" p
		LEFT JOIN player_id_map m ON p."playerID" = m.lahman_id
		WHERE p."playerID" = $1 OR m.retro_id = $1
		LIMIT 1
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

	dataSource := "lahman"

	var searchRetroID string
	if p.RetroID != nil {
		searchRetroID = string(*p.RetroID)
	} else {
		searchRetroID = string(id)
	}

	retroQuery := `
		SELECT season, team_id,
			CONCAT_WS(', ',
				NULLIF(CONCAT('P: ', games_p), 'P: 0'),
				NULLIF(CONCAT('C: ', games_c), 'C: 0'),
				NULLIF(CONCAT('1B: ', games_1b), '1B: 0'),
				NULLIF(CONCAT('2B: ', games_2b), '2B: 0'),
				NULLIF(CONCAT('3B: ', games_3b), '3B: 0'),
				NULLIF(CONCAT('SS: ', games_ss), 'SS: 0'),
				NULLIF(CONCAT('OF: ', games_of), 'OF: 0'),
				NULLIF(CONCAT('DH: ', games_dh), 'DH: 0')
			) as positions
		FROM retrosheet_players
		WHERE player_id = $1
		ORDER BY season DESC, last_game DESC
		LIMIT 1
	`

	var season sql.NullInt64
	var teamID, positions sql.NullString
	err = r.db.QueryRowContext(ctx, retroQuery, searchRetroID).Scan(&season, &teamID, &positions)
	if err == nil {
		if season.Valid {
			s := int(season.Int64)
			p.LatestSeason = &s
		}
		if teamID.Valid {
			p.LatestTeam = &teamID.String
		}
		if positions.Valid && positions.String != "" {
			p.Positions = &positions.String
		}
		if p.ID == "" {
			dataSource = "retrosheet"
		} else {
			dataSource = "lahman+retrosheet"
		}
	} else if err != sql.ErrNoRows {
		fmt.Printf("Warning: failed to fetch retrosheet data for %s: %v\n", searchRetroID, err)
	}

	p.DataSource = &dataSource

	_ = r.cache.Entity.Set(ctx, string(id), &p)

	return &p, nil
}

func (r *PlayerRepository) List(ctx context.Context, filter core.PlayerFilter) ([]core.Player, error) {
	args := []any{}
	argNum := 1

	nameFilter := ""
	if filter.NameQuery != "" {
		nameFilter = fmt.Sprintf(" AND (LOWER(\"nameFirst\" || ' ' || \"nameLast\") LIKE LOWER($%d))", argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	debutFilter := ""
	if filter.DebutYear != nil {
		debutFilter = fmt.Sprintf(" AND EXTRACT(YEAR FROM \"debut\"::date) = $%d", argNum)
		args = append(args, int(*filter.DebutYear))
		argNum++
	}

	retroNameFilter := ""
	if filter.NameQuery != "" {
		retroNameFilter = " AND (LOWER(rp.first_name || ' ' || rp.last_name) LIKE LOWER($1))"
	}

	query := fmt.Sprintf(`
		WITH lahman_players AS (
			SELECT
				"playerID", "nameFirst", "nameLast", "nameGiven",
				"birthYear", "birthMonth", "birthDay", "birthCity", "birthState", "birthCountry",
				"deathYear", "deathMonth", "deathDay", "deathCity", "deathState", "deathCountry",
				"bats", "throws", "weight", "height",
				"debut", "finalGame", "retroID", "bbrefID"
			FROM "People"
			WHERE 1=1%s%s
		), retrosheet_only AS (
			SELECT DISTINCT ON (rp.player_id)
				rp.player_id as "playerID",
				rp.first_name as "nameFirst",
				rp.last_name as "nameLast",
				rp.first_name || ' ' || rp.last_name as "nameGiven",
				NULL::int as "birthYear", NULL::int as "birthMonth", NULL::int as "birthDay",
				NULL::text as "birthCity", NULL::text as "birthState", NULL::text as "birthCountry",
				NULL::int as "deathYear", NULL::int as "deathMonth", NULL::int as "deathDay",
				NULL::text as "deathCity", NULL::text as "deathState", NULL::text as "deathCountry",
				rp.bats, rp.throws,
				NULL::int as "weight", NULL::int as "height",
				NULL::text as "debut", NULL::text as "finalGame",
				rp.player_id as "retroID",
				NULL::text as "bbrefID"
			FROM retrosheet_players rp
			LEFT JOIN "People" p ON p."retroID" = rp.player_id
			WHERE p."playerID" IS NULL%s
			ORDER BY rp.player_id, rp.season DESC
		)
		SELECT * FROM (
			SELECT * FROM lahman_players
			UNION ALL
			SELECT * FROM retrosheet_only
		) AS combined
		ORDER BY "nameLast", "nameFirst"
		LIMIT %d OFFSET %d
	`, nameFilter, debutFilter, retroNameFilter, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

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
	args := []any{}
	argNum := 1

	nameFilter := ""
	if filter.NameQuery != "" {
		nameFilter = fmt.Sprintf(" AND (LOWER(\"nameFirst\" || ' ' || \"nameLast\") LIKE LOWER($%d))", argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

	debutFilter := ""
	if filter.DebutYear != nil {
		debutFilter = fmt.Sprintf(" AND EXTRACT(YEAR FROM \"debut\"::date) = $%d", argNum)
		args = append(args, int(*filter.DebutYear))
		argNum++
	}

	retroNameFilter := ""
	if filter.NameQuery != "" {
		retroNameFilter = " AND (LOWER(rp.first_name || ' ' || rp.last_name) LIKE LOWER($1))"
	}

	query := fmt.Sprintf(`
		WITH lahman_count AS (
			SELECT COUNT(*) as cnt FROM "People" WHERE 1=1%s%s
		), retrosheet_count AS (
			SELECT COUNT(DISTINCT rp.player_id) as cnt
			FROM retrosheet_players rp
			LEFT JOIN "People" p ON p."retroID" = rp.player_id
			WHERE p."playerID" IS NULL%s
		)
		SELECT (SELECT cnt FROM lahman_count) + (SELECT cnt FROM retrosheet_count)
	`, nameFilter, debutFilter, retroNameFilter)

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
		s.DataSources = []string{"lahman"}

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

		s.DataSources = []string{"lahman"}

		seasons = append(seasons, s)
	}

	return seasons, nil
}

func (r *PlayerRepository) FieldingSeasons(ctx context.Context, id core.PlayerID) ([]core.PlayerFieldingSeason, error) {
	return nil, nil
}

// GameLogs retrieves games where the player appeared in the starting lineup.
func (r *PlayerRepository) GameLogs(ctx context.Context, id core.PlayerID, filter core.GameFilter) ([]core.Game, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return []core.Game{}, nil
	}

	query := playerGameLogsQuery

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
	query := playerAppearancesQuery

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

func (r *PlayerRepository) Teams(ctx context.Context, id core.PlayerID) ([]core.PlayerTeamSeason, error) {
	query := `
		SELECT
			a."yearID", a."teamID", a."lgID", a."G_all", a."GS",
			t."name"
		FROM "Appearances" a
		LEFT JOIN "Teams" t
			ON t."teamID" = a."teamID" AND t."yearID" = a."yearID"
		WHERE a."playerID" = $1
		ORDER BY a."yearID"
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team history: %w", err)
	}
	defer rows.Close()

	var seasons []core.PlayerTeamSeason
	for rows.Next() {
		var season core.PlayerTeamSeason
		var league sql.NullString
		var teamName sql.NullString
		var games, starts sql.NullInt64

		if err := rows.Scan(&season.Year, &season.TeamID, &league, &games, &starts, &teamName); err != nil {
			return nil, fmt.Errorf("failed to scan team season: %w", err)
		}

		season.PlayerID = id
		if league.Valid {
			lg := core.LeagueID(league.String)
			season.League = &lg
		}
		if teamName.Valid {
			name := teamName.String
			season.TeamName = &name
		}
		if games.Valid {
			season.Games = int(games.Int64)
		}
		if starts.Valid {
			season.GamesStarted = int(starts.Int64)
		}

		seasons = append(seasons, season)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate team history: %w", err)
	}

	return seasons, nil
}

func (r *PlayerRepository) Salaries(ctx context.Context, id core.PlayerID) ([]core.PlayerSalary, error) {
	query := `
		SELECT "yearID", "teamID", "lgID", "salary"
		FROM "Salaries"
		WHERE "playerID" = $1
		ORDER BY "yearID"
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch salaries: %w", err)
	}
	defer rows.Close()

	var salaries []core.PlayerSalary
	for rows.Next() {
		var salary core.PlayerSalary
		var league sql.NullString
		var amount sql.NullInt64

		if err := rows.Scan(&salary.Year, &salary.TeamID, &league, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan salary row: %w", err)
		}

		salary.PlayerID = id
		if league.Valid {
			lg := core.LeagueID(league.String)
			salary.League = &lg
		}
		if amount.Valid {
			salary.Salary = amount.Int64
		}

		salaries = append(salaries, salary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate salaries: %w", err)
	}

	return salaries, nil
}

// BattingGameLogs returns per-game batting statistics for a player from the materialized view
func (r *PlayerRepository) BattingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) ([]core.PlayerGameBattingLog, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return []core.PlayerGameBattingLog{}, nil
	}

	query := `
		SELECT
			player_id, game_id, date, season, team_id,
			pa, ab, h, singles, doubles, triples, hr, r, rbi,
			bb, so, hbp, sf, sh, sb, cs, ibb, gdp,
			avg, obp, slg
		FROM player_game_batting_stats
		WHERE player_id = $1
	`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.MinHR != nil {
		query += fmt.Sprintf(" AND hr >= $%d", argNum)
		args = append(args, *filter.MinHR)
		argNum++
	}

	if filter.MinH != nil {
		query += fmt.Sprintf(" AND h >= $%d", argNum)
		args = append(args, *filter.MinH)
		argNum++
	}

	if filter.MinRBI != nil {
		query += fmt.Sprintf(" AND rbi >= $%d", argNum)
		args = append(args, *filter.MinRBI)
		argNum++
	}

	if filter.MinPA != nil {
		query += fmt.Sprintf(" AND pa >= $%d", argNum)
		args = append(args, *filter.MinPA)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY date DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batting game logs: %w", err)
	}
	defer rows.Close()

	var logs []core.PlayerGameBattingLog
	for rows.Next() {
		var log core.PlayerGameBattingLog
		var playerID, gameID, date, teamID string
		var season int

		if err := rows.Scan(
			&playerID, &gameID, &date, &season, &teamID,
			&log.PA, &log.AB, &log.H, &log.Singles, &log.Doubles, &log.Triples, &log.HR, &log.R, &log.RBI,
			&log.BB, &log.SO, &log.HBP, &log.SF, &log.SH, &log.SB, &log.CS, &log.IBB, &log.GDP,
			&log.AVG, &log.OBP, &log.SLG,
		); err != nil {
			return nil, fmt.Errorf("failed to scan batting game log: %w", err)
		}

		log.PlayerID = id
		log.GameID = core.GameID(gameID)
		log.Date = date
		log.Season = core.SeasonYear(season)
		log.TeamID = core.TeamID(teamID)
		log.OPS = log.OBP + log.SLG

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate batting game logs: %w", err)
	}

	return logs, nil
}

// CountBattingGameLogs returns the count of batting game logs for a player
func (r *PlayerRepository) CountBattingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) (int, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return 0, nil
	}

	query := `SELECT COUNT(*) FROM player_game_batting_stats WHERE player_id = $1`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.MinHR != nil {
		query += fmt.Sprintf(" AND hr >= $%d", argNum)
		args = append(args, *filter.MinHR)
		argNum++
	}

	if filter.MinH != nil {
		query += fmt.Sprintf(" AND h >= $%d", argNum)
		args = append(args, *filter.MinH)
		argNum++
	}

	if filter.MinRBI != nil {
		query += fmt.Sprintf(" AND rbi >= $%d", argNum)
		args = append(args, *filter.MinRBI)
		argNum++
	}

	if filter.MinPA != nil {
		query += fmt.Sprintf(" AND pa >= $%d", argNum)
		args = append(args, *filter.MinPA)
		argNum++
	}

	var count int
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count batting game logs: %w", err)
	}

	return count, nil
}

// PitchingGameLogs returns per-game pitching statistics for a player from the materialized view
func (r *PlayerRepository) PitchingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) ([]core.PlayerGamePitchingLog, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return []core.PlayerGamePitchingLog{}, nil
	}

	query := `
		SELECT
			player_id, game_id, date, season, team_id,
			ip, pa, ab, h, r, er, bb, so, hr, hbp, ibb, wp, bk, sh, sf,
			era, whip, k9, bb9
		FROM player_game_pitching_stats
		WHERE player_id = $1
	`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.MinSO != nil {
		query += fmt.Sprintf(" AND so >= $%d", argNum)
		args = append(args, *filter.MinSO)
		argNum++
	}

	if filter.MinIP != nil {
		query += fmt.Sprintf(" AND ip >= $%d", argNum)
		args = append(args, *filter.MinIP)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY date DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pitching game logs: %w", err)
	}
	defer rows.Close()

	var logs []core.PlayerGamePitchingLog
	for rows.Next() {
		var log core.PlayerGamePitchingLog
		var playerID, gameID, date, teamID string
		var season int

		if err := rows.Scan(
			&playerID, &gameID, &date, &season, &teamID,
			&log.IP, &log.PA, &log.AB, &log.H, &log.R, &log.ER, &log.BB, &log.SO, &log.HR,
			&log.HBP, &log.IBB, &log.WP, &log.BK, &log.SH, &log.SF,
			&log.ERA, &log.WHIP, &log.K9, &log.BB9,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pitching game log: %w", err)
		}

		log.PlayerID = id
		log.GameID = core.GameID(gameID)
		log.Date = date
		log.Season = core.SeasonYear(season)
		log.TeamID = core.TeamID(teamID)

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pitching game logs: %w", err)
	}

	return logs, nil
}

// CountPitchingGameLogs returns the count of pitching game logs for a player
func (r *PlayerRepository) CountPitchingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) (int, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return 0, nil
	}

	query := `SELECT COUNT(*) FROM player_game_pitching_stats WHERE player_id = $1`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.MinSO != nil {
		query += fmt.Sprintf(" AND so >= $%d", argNum)
		args = append(args, *filter.MinSO)
		argNum++
	}

	if filter.MinIP != nil {
		query += fmt.Sprintf(" AND ip >= $%d", argNum)
		args = append(args, *filter.MinIP)
		argNum++
	}

	var count int
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pitching game logs: %w", err)
	}

	return count, nil
}

// FieldingGameLogs returns per-game fielding statistics for a player from the materialized view
func (r *PlayerRepository) FieldingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) ([]core.PlayerGameFieldingLog, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return []core.PlayerGameFieldingLog{}, nil
	}

	query := `
		SELECT
			player_id, game_id, date, season, team_id, position,
			po, a, e, tc, fpct
		FROM player_game_fielding_stats
		WHERE player_id = $1
	`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.Position != nil {
		query += fmt.Sprintf(" AND position = $%d", argNum)
		args = append(args, *filter.Position)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY date DESC, position LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fielding game logs: %w", err)
	}
	defer rows.Close()

	var logs []core.PlayerGameFieldingLog
	for rows.Next() {
		var log core.PlayerGameFieldingLog
		var playerID, gameID, date, teamID string
		var season int

		if err := rows.Scan(
			&playerID, &gameID, &date, &season, &teamID, &log.Position,
			&log.PO, &log.A, &log.E, &log.TC, &log.Fpct,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fielding game log: %w", err)
		}

		log.PlayerID = id
		log.GameID = core.GameID(gameID)
		log.Date = date
		log.Season = core.SeasonYear(season)
		log.TeamID = core.TeamID(teamID)

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate fielding game logs: %w", err)
	}

	return logs, nil
}

// CountFieldingGameLogs returns the count of fielding game logs for a player
func (r *PlayerRepository) CountFieldingGameLogs(ctx context.Context, id core.PlayerID, filter core.PlayerGameLogFilter) (int, error) {
	var retroID sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "retroID" FROM "People" WHERE "playerID" = $1`, string(id)).Scan(&retroID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("player not found: %s", id)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get player retroID: %w", err)
	}

	if !retroID.Valid || retroID.String == "" {
		return 0, nil
	}

	query := `SELECT COUNT(*) FROM player_game_fielding_stats WHERE player_id = $1`

	args := []any{retroID.String}
	argNum := 2

	if filter.Season != nil {
		query += fmt.Sprintf(" AND season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.Position != nil {
		query += fmt.Sprintf(" AND position = $%d", argNum)
		args = append(args, *filter.Position)
		argNum++
	}

	var count int
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count fielding game logs: %w", err)
	}

	return count, nil
}
